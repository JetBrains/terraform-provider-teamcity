package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource                   = &userResource{}
	_ resource.ResourceWithConfigure      = &userResource{}
	_ resource.ResourceWithValidateConfig = &userResource{}
	_ resource.ResourceWithImportState    = &userResource{}
)

type userResource struct {
	client *client.Client
}

func NewUserResource() resource.Resource {
	return &userResource{}
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

type userResourceModel struct {
	Id       types.String     `tfsdk:"id"`
	Username types.String     `tfsdk:"username"`
	Password types.String     `tfsdk:"password"`
	Github   types.String     `tfsdk:"github_username"`
	Roles    []roleAssignment `tfsdk:"roles"`
}

type roleAssignment struct {
	Id      types.String `tfsdk:"id"`
	Global  types.Bool   `tfsdk:"global"`
	Project types.String `tfsdk:"project"`
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"github_username": schema.StringAttribute{
				Optional: true,
			},
			"roles": schema.SetNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
						},
						"global": schema.BoolAttribute{
							Optional: true,
						},
						"project": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *userResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config userResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, role := range config.Roles {
		if role.Global.IsNull() && role.Project.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("roles"), //TODO path to specific set item
				"Either 'global' or 'project' must be specified",
				"",
			)
		}
		if !role.Global.IsNull() && !role.Project.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("roles"), //TODO path to specific set item
				"'global' and 'project' cannot be specified together",
				"",
			)
		}
		if !role.Global.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("roles"), //TODO path to specific set item
				"'global' must be set to 'true'",
				"",
			)
		}
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := r.update(plan)
	actual, err := r.client.NewUser(user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting user",
			"Cannot set user, unexpected error: "+err.Error(),
		)
		return
	}

	newState := r.readState(actual)
	newState.Password = plan.Password

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState userResourceModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var actual *client.User
	var err error
	if oldState.Id.IsNull() != true {
		actual, err = r.client.GetUser(oldState.Id.ValueString())
	} else {
		actual, err = r.client.GetUserByName(oldState.Username.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading user",
			"Could not read user settings: "+err.Error(),
		)
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := r.readState(actual)
	newState.Password = oldState.Password

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.Id.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing user id",
			err.Error(),
		)
		return
	}
	user := r.update(plan)
	user.Id = &id

	result, err := r.client.SetUser(user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating user",
			err.Error(),
		)
		return
	}

	newState := r.readState(result)
	newState.Password = plan.Password

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *userResource) update(plan userResourceModel) client.User {
	user := client.User{
		Username: plan.Username.ValueString(),
	}

	if plan.Password.IsNull() != true {
		password := plan.Password.ValueString()
		user.Password = &password
	}

	if plan.Github.IsNull() != true {
		user.Properties = &client.Properties{
			Property: []client.Property{
				{
					Name:  "plugin:auth:GitHubApp-oauth:userName",
					Value: plan.Github.ValueString(),
				},
			},
		}
	}

	user.Roles = &client.RoleAssignments{
		RoleAssignment: []client.RoleAssignment{},
	}

	if plan.Roles != nil {
		for _, role := range plan.Roles {
			assignment := client.RoleAssignment{
				Id: role.Id.ValueString(),
			}
			if role.Global.ValueBool() {
				assignment.Scope = "g"
			} else {
				assignment.Scope = "p:" + role.Project.ValueString()
			}
			user.Roles.RoleAssignment = append(user.Roles.RoleAssignment, assignment)
		}
	}
	return user
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting user",
			"Could not delete user, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("username"), req, resp)
}

func (r *userResource) readState(actual *client.User) userResourceModel {
	var newState userResourceModel
	newState.Id = types.StringValue(strconv.FormatInt(*actual.Id, 10))
	newState.Username = types.StringValue(actual.Username)

	for _, p := range actual.Properties.Property {
		if p.Name == "plugin:auth:GitHubApp-oauth:userName" {
			newState.Github = types.StringValue(p.Value)
			break
		}
	}

	if actual.Roles != nil && len(actual.Roles.RoleAssignment) > 0 {
		newState.Roles = []roleAssignment{}
		for _, role := range actual.Roles.RoleAssignment {
			assignment := roleAssignment{
				Id: types.StringValue(role.Id),
			}
			if role.Scope == "g" {
				assignment.Global = types.BoolValue(role.Scope == "g")
			} else {
				assignment.Project = types.StringValue(role.Scope[2:])
			}
			newState.Roles = append(newState.Roles, assignment)
		}
	}

	return newState
}
