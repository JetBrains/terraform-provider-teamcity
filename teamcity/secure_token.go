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
	_ resource.Resource              = &tokenResource{}
	_ resource.ResourceWithConfigure = &tokenResource{}
)

func NewTokenResource() resource.Resource {
	return &tokenResource{}
}

type tokenResource struct {
	client *client.Client
}

type tokenResourceModel struct {
	Id      types.String `tfsdk:"id"`
	Project types.String `tfsdk:"project_id"`
	Value   types.String `tfsdk:"value"`
}

func (r *tokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secure_token"
}

func (r *tokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *tokenResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *tokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tokenResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.AddSecureToken(plan.Project.ValueString(), plan.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding secure token",
			err.Error(),
		)
		return
	}

	var newState tokenResourceModel
	newState.Project = plan.Project
	newState.Id = types.StringValue(*id)
	newState.Value = plan.Value

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetSecureTokens(state.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading secure tokens",
			err.Error(),
		)
		return
	}

	for _, token := range result {
		if token == state.Id.ValueString() {
			var newState tokenResourceModel
			newState.Project = state.Project
			newState.Id = state.Id
			newState.Value = state.Value

			diags = resp.State.Set(ctx, newState)
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *tokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *tokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSecureToken(state.Project.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting secure token",
			err.Error(),
		)
		return
	}
}
