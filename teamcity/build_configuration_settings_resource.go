package teamcity

import (
	"context"
	"strconv"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &bcSettingsResource{}
	_ resource.ResourceWithConfigure = &bcSettingsResource{}
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
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"build_number_counter": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The next build number to be used.",
			},
			"build_number_pattern": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The pattern for the build number.",
			},
			"artifact_rules": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Rules for artifacts produced by the build.",
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

	if !plan.BuildNumberCounter.IsNull() {
		err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberCounter", strconv.FormatInt(plan.BuildNumberCounter.ValueInt64(), 10))
		if err != nil {
			resp.Diagnostics.AddError("Error setting buildNumberCounter", err.Error())
			return
		}
	}

	if !plan.BuildNumberPattern.IsNull() {
		err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberPattern", plan.BuildNumberPattern.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error setting buildNumberPattern", err.Error())
			return
		}
	}

	if !plan.ArtifactRules.IsNull() {
		err := r.client.SetBuildTypeSetting(buildTypeId, "artifactRules", plan.ArtifactRules.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error setting artifactRules", err.Error())
			return
		}
	}

	plan.Id = types.StringValue(buildTypeId)

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

	counter, err := r.client.GetBuildTypeSetting(buildTypeId, "buildNumberCounter")
	if err != nil {
		resp.Diagnostics.AddError("Error reading buildNumberCounter", err.Error())
		return
	}
	if counter != nil {
		val, _ := strconv.ParseInt(*counter, 10, 64)
		state.BuildNumberCounter = types.Int64Value(val)
	}

	pattern, err := r.client.GetBuildTypeSetting(buildTypeId, "buildNumberPattern")
	if err != nil {
		resp.Diagnostics.AddError("Error reading buildNumberPattern", err.Error())
		return
	}
	if pattern != nil {
		state.BuildNumberPattern = types.StringValue(*pattern)
	}

	rules, err := r.client.GetBuildTypeSetting(buildTypeId, "artifactRules")
	if err != nil {
		resp.Diagnostics.AddError("Error reading artifactRules", err.Error())
		return
	}
	if rules != nil {
		state.ArtifactRules = types.StringValue(*rules)
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

	if !plan.BuildNumberCounter.Equal(state.BuildNumberCounter) {
		err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberCounter", strconv.FormatInt(plan.BuildNumberCounter.ValueInt64(), 10))
		if err != nil {
			resp.Diagnostics.AddError("Error updating buildNumberCounter", err.Error())
			return
		}
	}

	if !plan.BuildNumberPattern.Equal(state.BuildNumberPattern) {
		err := r.client.SetBuildTypeSetting(buildTypeId, "buildNumberPattern", plan.BuildNumberPattern.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error updating buildNumberPattern", err.Error())
			return
		}
	}

	if !plan.ArtifactRules.Equal(state.ArtifactRules) {
		err := r.client.SetBuildTypeSetting(buildTypeId, "artifactRules", plan.ArtifactRules.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error updating artifactRules", err.Error())
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Settings cannot be deleted, only reset to defaults.
	// We don't have a good way to reset to defaults via individual field PUTs if we don't know the defaults.
	// For now, we just leave them as is or we could try to set them to empty strings if appropriate.
}
