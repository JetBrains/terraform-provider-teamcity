package teamcity

import (
	"context"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &cleanupResource{}
	_ resource.ResourceWithConfigure = &cleanupResource{}
)

func NewCleanupResource() resource.Resource {
	return &cleanupResource{}
}

type cleanupResource struct {
	client *client.Client
}
type cleanupResourceModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

func (r *cleanupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cleanup"
}

func (r *cleanupResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"enabled": {
				Type:     types.BoolType,
				Required: true,
			},
		},
	}, nil
}

func (r *cleanupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *cleanupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cleanupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var enabled = plan.Enabled.Value

	settings := client.Settings{
		Enabled: enabled,
	}
	result, err := r.client.SetCleanup(settings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting cleanup",
			"Cannot set cleanup, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Enabled = types.Bool{Value: result.Enabled}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cleanupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cleanupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actual, err := r.client.GetCleanup()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Cleanup",
			"Could not read cleanup settings: "+err.Error(),
		)
		return
	}

	state.Enabled = types.Bool{Value: actual.Enabled}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cleanupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *cleanupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
