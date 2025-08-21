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
	"strings"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource                = &groupRoleAssignmentResource{}
	_ resource.ResourceWithConfigure   = &groupRoleAssignmentResource{}
	_ resource.ResourceWithImportState = &groupRoleAssignmentResource{}
)

type groupRoleAssignmentResource struct {
	client *client.Client
}

func NewGroupRoleAssignmentResource() resource.Resource {
	return &groupRoleAssignmentResource{}
}

func (r *groupRoleAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_role_assignment"
}

type groupRoleAssignmentResourceModel struct {
	Id      types.String `tfsdk:"id"`
	GroupId types.String `tfsdk:"group_id"`
	RoleId  types.String `tfsdk:"role_id"`
	Scope   types.String `tfsdk:"scope"`
}

func (r *groupRoleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages role assignments for TeamCity groups. This resource allows you to assign roles to groups either globally or for specific projects.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the role assignment (computed).",
			},
			"group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The ID or key of the group to assign the role to.",
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

func (r *groupRoleAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *groupRoleAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupRoleAssignmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine scope
	scope := "g" // default to global
	if !plan.Scope.IsNull() {
		scope = plan.Scope.ValueString()
	}

	// Add role to group
	err := r.client.AddGroupRole(plan.GroupId.ValueString(), plan.RoleId.ValueString(), scope)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error assigning role to group",
			err.Error(),
		)
		return
	}

	// Set state
	state := groupRoleAssignmentResourceModel{
		Id:      types.StringValue(fmt.Sprintf("%s_%s_%s", plan.GroupId.ValueString(), plan.RoleId.ValueString(), scope)),
		GroupId: plan.GroupId,
		RoleId:  plan.RoleId,
		Scope:   types.StringValue(scope),
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *groupRoleAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupRoleAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get group to verify role assignment still exists
	group, err := r.client.GetGroup(state.GroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading group",
			err.Error(),
		)
		return
	}

	if group == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Check if role assignment still exists
	found := false
	if group.Roles != nil {
		for _, role := range group.Roles.RoleAssignment {
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

	// Role assignment exists, keep current state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *groupRoleAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource doesn't support updates - all attributes have RequiresReplace
	// The framework will handle this by destroying and recreating the resource
	resp.Diagnostics.AddError(
		"Update not supported",
		"This resource does not support updates. All changes require resource replacement.",
	)
}

func (r *groupRoleAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupRoleAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove role from group
	err := r.client.RemoveGroupRole(state.GroupId.ValueString(), state.RoleId.ValueString(), state.Scope.ValueString())
	if err != nil {
		// Check if it's a 404 error (already deleted)
		if !strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Error removing role from group",
				err.Error(),
			)
		}
	}
}

func (r *groupRoleAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: group_id/role_id/scope
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format: group_id/role_id/scope",
		)
		return
	}

	state := groupRoleAssignmentResourceModel{
		Id:      types.StringValue(fmt.Sprintf("%s_%s_%s", parts[0], parts[1], parts[2])),
		GroupId: types.StringValue(parts[0]),
		RoleId:  types.StringValue(parts[1]),
		Scope:   types.StringValue(parts[2]),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
