package teamcity

import (
	"context"
	"strings"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &roleResource{}
	_ resource.ResourceWithConfigure = &roleResource{}
)

type roleResource struct {
	client *client.Client
}

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

type roleResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Id          types.String `tfsdk:"id"`
	Included    types.Set    `tfsdk:"included"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func (r *roleResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:     types.StringType,
				Required: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"included": {
				Type:     types.SetType{ElemType: types.StringType},
				Optional: true,
			},
			"permissions": {
				Type:     types.SetType{ElemType: types.StringType},
				Optional: true,
			},
		},
	}, nil
}

func (r *roleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	val := plan.Name.ValueString()
	role := client.Role{
		Name: &val,
	}
	if !plan.Included.IsNull() {
		var roles []types.String
		diags = plan.Included.ElementsAs(ctx, &roles, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		role.Included = &client.Included{Role: []*client.Role{}}
		for _, i := range roles {
			val := i.ValueString()
			role.Included.Role = append(
				role.Included.Role,
				&client.Role{Id: &val},
			)
		}
	}

	if !plan.Permissions.IsNull() {
		var perms []types.String
		diags = plan.Permissions.ElementsAs(ctx, &perms, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		role.Permissions = &client.Permissions{Permission: []*client.Permission{}}
		for _, i := range perms {
			v := i.ValueString()
			role.Permissions.Permission = append(
				role.Permissions.Permission,
				&client.Permission{Id: &v},
			)
		}
	}

	actual, err := r.client.NewRole(role)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting role",
			"Cannot set role, unexpected error: "+err.Error(),
		)
		return
	}

	var newState roleResourceModel
	readRoleState(&newState, actual)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actual, err := r.client.GetRole(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading role",
			"Could not read role settings: "+err.Error(),
		)
		return
	}

	var newState roleResourceModel
	readRoleState(&newState, actual)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state roleResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState roleResourceModel

	var stateIncluded []types.String
	state.Included.ElementsAs(ctx, &stateIncluded, false)
	var planIncluded []types.String
	plan.Included.ElementsAs(ctx, &planIncluded, false)

	// items present in old state but missing in a plan -> remove
	for _, i := range stateIncluded {
		if !contains(planIncluded, i) {
			actual, err := r.client.RemoveIncludedRole(plan.Id.ValueString(), i.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error removing included role",
					"Unexpected error: "+err.Error(),
				)
				return
			}
			readRoleState(&newState, actual)
		}
	}

	// items missing in old state but present in a plan -> add
	for _, i := range planIncluded {
		if !contains(stateIncluded, i) {
			actual, err := r.client.AddIncludedRole(plan.Id.ValueString(), i.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error adding included role",
					"Unexpected error: "+err.Error(),
				)
				return
			}
			readRoleState(&newState, actual)
		}
	}

	var statePerms []types.String
	state.Permissions.ElementsAs(ctx, &statePerms, false)
	var planPerms []types.String
	plan.Permissions.ElementsAs(ctx, &planPerms, false)

	// items present in old state but missing in a plan -> remove
	for _, i := range statePerms {
		if !contains(planPerms, i) {
			actual, err := r.client.RemovePermission(plan.Id.ValueString(), i.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error removing permission",
					"Unexpected error: "+err.Error(),
				)
				return
			}
			readRoleState(&newState, actual)
		}
	}

	// items missing in old state but present in a plan -> add
	for _, i := range planPerms {
		if !contains(statePerms, i) {
			actual, err := r.client.AddPermission(plan.Id.ValueString(), i.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error adding permission",
					"Unexpected error: "+err.Error(),
				)
				return
			}
			readRoleState(&newState, actual)
		}
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func contains(items []types.String, item types.String) bool {
	for _, i := range items {
		if i.Equal(item) {
			return true
		}
	}
	return false
}

func readRoleState(newState *roleResourceModel, actual *client.Role) {
	newState.Name = types.StringValue(*actual.Name)
	newState.Id = types.StringValue(*actual.Id)

	newState.Included = types.SetNull(types.StringType)
	for _, i := range actual.Included.Role {
		newState.Included, _ = types.SetValue(
			types.StringType,
			append(newState.Included.Elements(), types.StringValue(*i.Id)),
		)
	}

	newState.Permissions = types.SetNull(types.StringType)
	for _, i := range actual.Permissions.Permission {
		v := strings.ToUpper(*i.Id) //TODO bug in REST
		newState.Permissions, _ = types.SetValue(
			types.StringType,
			append(newState.Permissions.Elements(), types.StringValue(v)),
		)
	}
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRole(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting role",
			"Could not delete role, unexpected error: "+err.Error(),
		)
		return
	}
}
