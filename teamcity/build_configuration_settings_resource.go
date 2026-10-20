package teamcity

import (
	"context"
	"errors"
	"strconv"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &bcSettingsResource{}
	_ resource.ResourceWithConfigure   = &bcSettingsResource{}
	_ resource.ResourceWithImportState = &bcSettingsResource{}
)

func NewBuildConfigurationSettingsResource() resource.Resource {
	return &bcSettingsResource{}
}

type bcSettingsResource struct {
	client *client.Client
}

type bcSettingsResourceModel struct {
	Id                   types.String `tfsdk:"id"`
	BuildConfigurationId types.String `tfsdk:"build_configuration_id"`
	BuildNumberCounter   types.Int64  `tfsdk:"build_number_counter"`
	BuildNumberPattern   types.String `tfsdk:"build_number_pattern"`
	ArtifactRules        types.String `tfsdk:"artifact_rules"`
}

func (r *bcSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration_settings"
}

func (r *bcSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "General settings for a build configuration, including build number counter, pattern, and artifact rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (same as build_configuration_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_configuration_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the build configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"build_number_counter": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The next build number to be used.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"build_number_pattern": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The pattern for the build number.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"artifact_rules": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Rules for artifacts produced by the build.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *bcSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *bcSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bcSettingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()

	if !plan.BuildNumberCounter.IsNull() && !plan.BuildNumberCounter.IsUnknown() {
		err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberCounter", strconv.FormatInt(plan.BuildNumberCounter.ValueInt64(), 10))
		if err != nil {
			resp.Diagnostics.AddError("Error setting buildNumberCounter", err.Error())
			return
		}
	}

	if !plan.BuildNumberPattern.IsNull() && !plan.BuildNumberPattern.IsUnknown() {
		err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberPattern", plan.BuildNumberPattern.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error setting buildNumberPattern", err.Error())
			return
		}
	}

	if !plan.ArtifactRules.IsNull() && !plan.ArtifactRules.IsUnknown() {
		err := r.client.SetBuildTypeSetting(buildTypeId, "artifactRules", plan.ArtifactRules.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error setting artifactRules", err.Error())
			return
		}
	}

	plan.Id = types.StringValue(buildTypeId)

	if _, ok := r.readSettings(buildTypeId, &plan, &resp.Diagnostics); !ok {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bcSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	state.Id = types.StringValue(buildTypeId)

	found, ok := r.readSettings(buildTypeId, &state, &resp.Diagnostics)
	if !ok {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *bcSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bcSettingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state bcSettingsResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()

	if !plan.BuildNumberCounter.IsNull() && !plan.BuildNumberCounter.IsUnknown() && !plan.BuildNumberCounter.Equal(state.BuildNumberCounter) {
		err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberCounter", strconv.FormatInt(plan.BuildNumberCounter.ValueInt64(), 10))
		if err != nil {
			resp.Diagnostics.AddError("Error updating buildNumberCounter", err.Error())
			return
		}
	}

	if !plan.BuildNumberPattern.IsNull() && !plan.BuildNumberPattern.IsUnknown() && !plan.BuildNumberPattern.Equal(state.BuildNumberPattern) {
		err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberPattern", plan.BuildNumberPattern.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error updating buildNumberPattern", err.Error())
			return
		}
	}

	if !plan.ArtifactRules.IsNull() && !plan.ArtifactRules.IsUnknown() && !plan.ArtifactRules.Equal(state.ArtifactRules) {
		err := r.client.SetBuildTypeSetting(buildTypeId, "artifactRules", plan.ArtifactRules.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error updating artifactRules", err.Error())
			return
		}
	}

	if _, ok := r.readSettings(buildTypeId, &plan, &resp.Diagnostics); !ok {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bcSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()

	// Settings cannot be deleted, only reset to defaults.
	// We reset them to TeamCity defaults.
	if err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberCounter", "1"); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error resetting buildNumberCounter", err.Error())
	}
	if err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberPattern", "%build.counter%"); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error resetting buildNumberPattern", err.Error())
	}
	if err := r.client.SetBuildTypeSetting(buildTypeId, "artifactRules", ""); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error resetting artifactRules", err.Error())
	}
}

func (r *bcSettingsResource) readSettings(buildTypeId string, model *bcSettingsResourceModel, diags *diag.Diagnostics) (bool, bool) {
	counter, err := r.client.GetBuildTypeSetting(buildTypeId, "buildNumberCounter")
	if err != nil {
		diags.AddError("Error reading buildNumberCounter", err.Error())
		return false, false
	}

	pattern, err := r.client.GetBuildTypeSetting(buildTypeId, "buildNumberPattern")
	if err != nil {
		diags.AddError("Error reading buildNumberPattern", err.Error())
		return false, false
	}

	rules, err := r.client.GetBuildTypeSetting(buildTypeId, "artifactRules")
	if err != nil {
		diags.AddError("Error reading artifactRules", err.Error())
		return false, false
	}

	if counter == nil && pattern == nil && rules == nil {
		return false, true
	}

	if counter != nil {
		val, err := strconv.ParseInt(*counter, 10, 64)
		if err != nil {
			diags.AddError("Error parsing buildNumberCounter", err.Error())
			return true, false
		}
		model.BuildNumberCounter = types.Int64Value(val)
	}

	if pattern != nil {
		model.BuildNumberPattern = types.StringValue(*pattern)
	}

	if rules != nil {
		model.ArtifactRules = types.StringValue(*rules)
	}

	return true, true
}

func (r *bcSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("build_configuration_id"), req.ID)...)
}
