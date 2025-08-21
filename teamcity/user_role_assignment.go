package teamcity

import (
	"context"
	"fmt"
	_ "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
	"strings"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource                = &userRoleAssignmentResource{}
	_ resource.ResourceWithConfigure   = &userRoleAssignmentResource{}
	_ resource.ResourceWithImportState = &userRoleAssignmentResource{}
)

type userRoleAssignmentResource struct {
	client *client.Client
}

func NewUserRoleAssignmentResource() resource.Resource {
	return &userRoleAssignmentResource{}
}

func (r *userRoleAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_role_assignment"
}

type userRoleAssignmentResourceModel struct {
	Id       types.String `tfsdk:"id"`
	UserId   types.String `tfsdk:"user_id"`
	Username types.String `tfsdk:"username"`
	RoleId   types.String `tfsdk:"role_id"`
	Scope    types.String `tfsdk:"scope"`
}

func (r *userRoleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages role assignments for TeamCity users. This resource allows you to assign roles to users either globally or for specific projects.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the role assignment (computed).",
			},
			"user_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the user to assign the role to. Either user_id or username must be specified.",
			},
			"username": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The username of the user to assign the role to. Either user_id or username must be specified.",
			},
			"role_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The ID of the role to assign.",
			},
			"scope": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The scope of the role assignment. Use 'g' for global scope or 'p:PROJECT_ID' for project-specific scope. Defaults to global if not specified.",
			},
		},
	}
}

func (r *userRoleAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *userRoleAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userRoleAssignmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either user_id or username is specified
	if plan.UserId.IsNull() && plan.Username.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid configuration",
			"Either 'user_id' or 'username' must be specified",
		)
		return
	}

	// Get user to obtain ID if username was provided
	var user *client.User
	var err error
	var userId string

	if !plan.Username.IsNull() {
		user, err = r.client.GetUserByName(plan.Username.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting user",
				err.Error(),
			)
			return
		}
		if user == nil {
			resp.Diagnostics.AddError(
				"User not found",
				"User with username '"+plan.Username.ValueString()+"' not found",
			)
			return
		}
		userId = strconv.FormatInt(*user.Id, 10)
	} else {
		userId = plan.UserId.ValueString()
		// Verify user exists
		user, err = r.client.GetUser(userId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting user",
				err.Error(),
			)
			return
		}
		if user == nil {
			resp.Diagnostics.AddError(
				"User not found",
				"User with ID '"+userId+"' not found",
			)
			return
		}
	}

	// Determine scope
	scope := "g" // default to global
	if !plan.Scope.IsNull() {
		scope = plan.Scope.ValueString()
	}

	// Build updated user with new role
	updatedUser := client.User{
		Id:       user.Id,
		Username: user.Username,
	}

	// Copy existing roles and add new one
	updatedUser.Roles = &client.RoleAssignments{
		RoleAssignment: []client.RoleAssignment{},
	}

	// Copy existing roles
	if user.Roles != nil {
		for _, role := range user.Roles.RoleAssignment {
			updatedUser.Roles.RoleAssignment = append(updatedUser.Roles.RoleAssignment, role)
		}
	}

	// Add new role
	updatedUser.Roles.RoleAssignment = append(updatedUser.Roles.RoleAssignment, client.RoleAssignment{
		Id:    plan.RoleId.ValueString(),
		Scope: scope,
	})

	// Update user
	_, err = r.client.SetUser(updatedUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error assigning role to user",
			err.Error(),
		)
		return
	}

	// Set state
	state := userRoleAssignmentResourceModel{
		Id:       types.StringValue(fmt.Sprintf("%s_%s_%s", userId, plan.RoleId.ValueString(), scope)),
		UserId:   types.StringValue(userId),
		Username: types.StringValue(user.Username),
		RoleId:   plan.RoleId,
		Scope:    types.StringValue(scope),
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *userRoleAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userRoleAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get user to verify role assignment still exists
	user, err := r.client.GetUser(state.UserId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading user",
			err.Error(),
		)
		return
	}

	if user == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update username in case it changed
	state.Username = types.StringValue(user.Username)

	// Check if role assignment still exists
	found := false
	if user.Roles != nil {
		for _, role := range user.Roles.RoleAssignment {
			if role.Id == state.RoleId.ValueString() && role.Scope == state.Scope.ValueString() {
				found = true
				break
			}
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Role assignment exists, update state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *userRoleAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource doesn't support updates - all attributes have RequiresReplace
	// The framework will handle this by destroying and recreating the resource
	resp.Diagnostics.AddError(
		"Update not supported",
		"This resource does not support updates. All changes require resource replacement.",
	)
}

func (r *userRoleAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userRoleAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current user
	user, err := r.client.GetUser(state.UserId.ValueString())
	if err != nil {
		// If user doesn't exist, consider it deleted
		if strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError(
			"Error getting user",
			err.Error(),
		)
		return
	}

	if user == nil {
		// User doesn't exist, nothing to delete
		return
	}

	// Build updated user without the role
	updatedUser := client.User{
		Id:       user.Id,
		Username: user.Username,
	}

	updatedUser.Roles = &client.RoleAssignments{
		RoleAssignment: []client.RoleAssignment{},
	}

	// Copy existing roles except the one being deleted
	if user.Roles != nil {
		for _, role := range user.Roles.RoleAssignment {
			if !(role.Id == state.RoleId.ValueString() && role.Scope == state.Scope.ValueString()) {
				updatedUser.Roles.RoleAssignment = append(updatedUser.Roles.RoleAssignment, role)
			}
		}
	}

	// Update user
	_, err = r.client.SetUser(updatedUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing role from user",
			err.Error(),
		)
	}
}

func (r *userRoleAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: user_id/role_id/scope or username/role_id/scope
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format: user_id/role_id/scope or username/role_id/scope",
		)
		return
	}

	// Try to determine if first part is user_id or username
	var userId string
	var username string

	// Check if it's a numeric ID
	if _, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
		// It's a user ID
		userId = parts[0]
		// Get username
		user, err := r.client.GetUser(userId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting user",
				err.Error(),
			)
			return
		}
		if user != nil {
			username = user.Username
		}
	} else {
		// It's a username
		username = parts[0]
		// Get user ID
		user, err := r.client.GetUserByName(username)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error getting user",
				err.Error(),
			)
			return
		}
		if user != nil {
			userId = strconv.FormatInt(*user.Id, 10)
		}
	}

	state := userRoleAssignmentResourceModel{
		Id:       types.StringValue(fmt.Sprintf("%s_%s_%s", userId, parts[1], parts[2])),
		UserId:   types.StringValue(userId),
		Username: types.StringValue(username),
		RoleId:   types.StringValue(parts[1]),
		Scope:    types.StringValue(parts[2]),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
