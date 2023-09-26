package teamcity

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource                = &memberResource{}
	_ resource.ResourceWithConfigure   = &memberResource{}
	_ resource.ResourceWithImportState = &memberResource{}
)

func NewMemberResource() resource.Resource {
	return &memberResource{}
}

type memberResource struct {
	client *client.Client
}

type memberResourceModel struct {
	GroupId  types.String `tfsdk:"group_id"`
	Username types.String `tfsdk:"username"`
}

func (r *memberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_member"
}

func (r *memberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *memberResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *memberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan memberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddGroupMember(plan.GroupId.ValueString(), plan.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding group member",
			err.Error(),
		)
		return
	}

	var newState memberResourceModel
	newState.GroupId = plan.GroupId
	newState.Username = plan.Username

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *memberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState memberResourceModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ok, err := r.client.CheckGroupMember(oldState.GroupId.ValueString(), oldState.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading group member",
			err.Error(),
		)
		return
	}

	if !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	var newState memberResourceModel
	newState.GroupId = oldState.GroupId
	newState.Username = oldState.Username

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *memberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *memberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state memberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGroupMember(state.GroupId.ValueString(), state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting group member",
			err.Error(),
		)
		return
	}
}

func (r *memberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: group_id/username. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), idParts[1])...)
}
