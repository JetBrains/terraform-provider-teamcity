package teamcity

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &bcVcsRootResource{}
	_ resource.ResourceWithConfigure   = &bcVcsRootResource{}
	_ resource.ResourceWithImportState = &bcVcsRootResource{}
)

func NewBuildConfigurationVcsRootResource() resource.Resource {
	return &bcVcsRootResource{}
}

type bcVcsRootResource struct {
	client *client.Client
}

func (r *bcVcsRootResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration_vcs_root"
}

func (r *bcVcsRootResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Attaches a VCS root to a build configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier in the form 'build_configuration_id/vcs_root_id'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_configuration_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vcs_root_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"checkout_rules": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Checkout rules for the VCS root.",
			},
		},
	}
}

func (r *bcVcsRootResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *bcVcsRootResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.VcsRootEntryDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	vcsRootId := plan.VcsRootId.ValueString()

	entry := models.VcsRootEntryJson{
		VcsRoot: &models.VcsRootJson{
			ID: &vcsRootId,
		},
		CheckoutRules: plan.CheckoutRules.ValueString(),
	}

	actual, err := r.client.NewBuildTypeVcsRootEntry(buildTypeId, entry)
	if err != nil {
		resp.Diagnostics.AddError("Error attaching VCS root", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, vcsRootId))
	if actual.CheckoutRules != "" {
		plan.CheckoutRules = types.StringValue(actual.CheckoutRules)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcVcsRootResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.VcsRootEntryDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	vcsRootId := state.VcsRootId.ValueString()

	actual, err := r.client.GetBuildTypeVcsRootEntry(buildTypeId, vcsRootId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading VCS root attachment", err.Error())
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.CheckoutRules = types.StringValue(actual.CheckoutRules)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *bcVcsRootResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.VcsRootEntryDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.VcsRootEntryDataModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	vcsRootId := plan.VcsRootId.ValueString()

	if !plan.CheckoutRules.Equal(state.CheckoutRules) {
		entry := models.VcsRootEntryJson{
			VcsRoot: &models.VcsRootJson{
				ID: &vcsRootId,
			},
			CheckoutRules: plan.CheckoutRules.ValueString(),
		}
		_, err := r.client.UpdateBuildTypeVcsRootEntry(buildTypeId, vcsRootId, entry)
		if err != nil {
			resp.Diagnostics.AddError("Error updating VCS root attachment", err.Error())
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcVcsRootResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.VcsRootEntryDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	vcsRootId := state.VcsRootId.ValueString()

	err := r.client.DeleteBuildTypeVcsRootEntry(buildTypeId, vcsRootId)
	if err != nil {
		resp.Diagnostics.AddError("Error detaching VCS root", err.Error())
		return
	}
}

func (r *bcVcsRootResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: build_configuration_id/vcs_root_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("build_configuration_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vcs_root_id"), idParts[1])...)
}
