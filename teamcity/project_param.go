package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource              = &paramResource{}
	_ resource.ResourceWithConfigure = &paramResource{}
)

func NewParamResource() resource.Resource {
	return &paramResource{}
}

type paramResource struct {
	client *client.Client
}

type paramResourceModel struct {
	ProjectId types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Value     types.String `tfsdk:"value"`
}

func (r *paramResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_parameter"
}

func (r *paramResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *paramResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *paramResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan paramResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetParam(plan.ProjectId.ValueString(), plan.Name.ValueString(), plan.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding project parameter",
			err.Error(),
		)
		return
	}

	var newState paramResourceModel
	newState.ProjectId = plan.ProjectId
	newState.Name = plan.Name
	newState.Value = plan.Value

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *paramResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState paramResourceModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetParam(oldState.ProjectId.ValueString(), oldState.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading group param",
			err.Error(),
		)
		return
	}

	if result == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var newState paramResourceModel
	newState.ProjectId = oldState.ProjectId
	newState.Name = oldState.Name
	newState.Value = types.StringValue(*result)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *paramResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan paramResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var oldState paramResourceModel
	diags = req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Value.Equal(oldState.Value) {
		err := r.client.SetParam(plan.ProjectId.ValueString(), plan.Name.ValueString(), plan.Value.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating project param",
				err.Error(),
			)
			return
		}
	}

	var newState paramResourceModel
	newState.ProjectId = plan.ProjectId
	newState.Name = plan.Name
	newState.Value = plan.Value

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *paramResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state paramResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteParam(state.ProjectId.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting project param",
			err.Error(),
		)
		return
	}
}
