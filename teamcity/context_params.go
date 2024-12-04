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
	_ resource.Resource                = &contextParamsResource{}
	_ resource.ResourceWithConfigure   = &contextParamsResource{}
	_ resource.ResourceWithImportState = &contextParamsResource{}
)

func NewContextParamsResource() resource.Resource {
	return &contextParamsResource{}
}

type contextParamsResource struct {
	client *client.Client
}

type contextParamsResourceModel struct {
	Project types.String `tfsdk:"project_id"`
	Params  types.Map    `tfsdk:"params"`
}

func (r *contextParamsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_context_parameters"
}

func (r *contextParamsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "With context parameters, it is possible to maintain a single Kotlin DSL code and use it in different projects on the same TeamCity server. Each of these projects can have own values of context parameters. More details [here](https://www.jetbrains.com/help/teamcity/kotlin-dsl.html#Use+Context+Parameters+in+DSL))",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"params": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *contextParamsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *contextParamsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan contextParamsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var params map[string]string
	diags = plan.Params.ElementsAs(ctx, &params, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.SetContextParams(plan.Project.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting context parameters",
			err.Error(),
		)
		return
	}
	v, diags := types.MapValueFrom(ctx, types.StringType, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState contextParamsResourceModel
	newState.Project = plan.Project
	newState.Params = v

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *contextParamsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state contextParamsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetContextParams(state.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading context parameters",
			err.Error(),
		)
		return
	}

	v, diags := types.MapValueFrom(ctx, types.StringType, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState contextParamsResourceModel
	newState.Project = state.Project
	newState.Params = v

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *contextParamsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan contextParamsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var params map[string]string
	diags = plan.Params.ElementsAs(ctx, &params, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.SetContextParams(plan.Project.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting context parameters",
			err.Error(),
		)
		return
	}
	v, diags := types.MapValueFrom(ctx, types.StringType, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState contextParamsResourceModel
	newState.Project = plan.Project
	newState.Params = v

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *contextParamsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state contextParamsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var params map[string]string
	_, err := r.client.SetContextParams(state.Project.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting context parameters",
			err.Error(),
		)
		return
	}
}

func (r *contextParamsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("project_id"), req, resp)
}
