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
	_ resource.Resource                = &licenseResource{}
	_ resource.ResourceWithConfigure   = &licenseResource{}
	_ resource.ResourceWithImportState = &licenseResource{}
)

type licenseResource struct {
	client *client.Client
}

func NewLicenseResource() resource.Resource {
	return &licenseResource{}
}

func (r *licenseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

type licenseResourceModel struct {
	Key types.String `tfsdk:"key"`
}

func (r *licenseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *licenseResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *licenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan licenseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.NewLicense(plan.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding license key",
			err.Error(),
		)
		return
	}

	var newState licenseResourceModel
	newState.Key = plan.Key

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *licenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState licenseResourceModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ok, err := r.client.CheckLicense(oldState.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading license key",
			err.Error(),
		)
		return
	}

	if ok == false {
		resp.State.RemoveResource(ctx)
		return
	}

	var newState licenseResourceModel
	newState.Key = oldState.Key

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *licenseResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *licenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state licenseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLicense(state.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting license key",
			err.Error(),
		)
		return
	}
}

func (r *licenseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
