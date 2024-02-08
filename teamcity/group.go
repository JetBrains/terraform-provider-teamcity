package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

type groupResource struct {
	client *client.Client
}

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

type groupResourceModel struct {
	Id           types.String     `tfsdk:"id"`
	Name         types.String     `tfsdk:"name"`
	Roles        []roleAssignment `tfsdk:"roles"`
	ParentGroups types.Set        `tfsdk:"parent_groups"`
}

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "User groups help manage user accounts more efficiently via roles and notification rules. More details [here](https://www.jetbrains.com/help/teamcity/creating-and-managing-user-groups.html).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				//TODO Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"parent_groups": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := client.Group{
		Name: plan.Name.ValueString(),
	}
	group.Roles = &client.RoleAssignments{
		RoleAssignment: []client.RoleAssignment{},
	}

	if plan.Roles != nil {
		for _, role := range plan.Roles {
			assignment := client.RoleAssignment{
				Id: role.Id.ValueString(),
			}
			assignment.Scope = scope(role)
			group.Roles.RoleAssignment = append(group.Roles.RoleAssignment, assignment)
		}
	}

	if !plan.ParentGroups.IsNull() {
		var parents []types.String
		diags = plan.ParentGroups.ElementsAs(ctx, &parents, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		group.Parents = &client.ParentGroups{}
		for _, i := range parents {
			val := i.ValueString()
			group.Parents.Group = append(
				group.Parents.Group,
				client.Group{Key: val},
			)
		}
	}

	actual, err := r.client.NewGroup(group)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding group",
			err.Error(),
		)
		return
	}

	newState := r.readState(actual)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState groupResourceModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actual, err := r.client.GetGroup(oldState.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading group",
			err.Error(),
		)
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := r.readState(actual)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var oldState groupResourceModel
	diags = req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// items present in old state but missing in a plan -> remove
	for _, i := range oldState.Roles {
		if !contains3(plan.Roles, i) {
			err := r.client.RemoveGroupRole(plan.Id.ValueString(), i.Id.ValueString(), scope(i))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error removing group role",
					err.Error(),
				)
				return
			}
		}
	}

	// items missing in old state but present in a plan -> add
	for _, i := range plan.Roles {
		if !contains3(oldState.Roles, i) {
			err := r.client.AddGroupRole(plan.Id.ValueString(), i.Id.ValueString(), scope(i))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error adding group role",
					err.Error(),
				)
				return
			}
		}
	}

	if !plan.ParentGroups.Equal(oldState.ParentGroups) {
		var parents []string
		diags = plan.ParentGroups.ElementsAs(ctx, &parents, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		err := r.client.SetGroupParents(plan.Id.ValueString(), parents)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading group",
				err.Error(),
			)
			return
		}

	}

	actual, err := r.client.GetGroup(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading group",
			err.Error(),
		)
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := r.readState(actual)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGroup(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting group",
			err.Error(),
		)
		return
	}
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *groupResource) readState(actual *client.Group) groupResourceModel {
	var newState groupResourceModel
	newState.Id = types.StringValue(actual.Key)
	newState.Name = types.StringValue(actual.Name)

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

	newState.ParentGroups = types.SetNull(types.StringType)
	for _, parent := range actual.Parents.Group {
		newState.ParentGroups, _ = types.SetValue(
			types.StringType,
			append(newState.ParentGroups.Elements(), types.StringValue(parent.Key)),
		)
	}

	return newState
}

func scope(i roleAssignment) string {
	var scope string
	if i.Global.ValueBool() {
		scope = "g"
	} else {
		scope = "p:" + i.Project.ValueString()
	}
	return scope
}

func contains3(items []roleAssignment, item roleAssignment) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}
	return false
}
