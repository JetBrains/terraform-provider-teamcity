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
	_ resource.Resource                = &bcFeatureResource{}
	_ resource.ResourceWithConfigure   = &bcFeatureResource{}
	_ resource.ResourceWithImportState = &bcFeatureResource{}
)

func NewBuildConfigurationFeatureResource() resource.Resource {
	return &bcFeatureResource{}
}

type bcFeatureResource struct {
	client *client.Client
}

func (r *bcFeatureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration_feature"
}

func (r *bcFeatureResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A build feature in a TeamCity build configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (Feature ID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_configuration_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the build configuration to which this feature belongs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the build feature (e.g., swabra, freeDiskSpace, xml-report-plugin).",
			},
			"properties": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Properties for the build feature.",
			},
		},
	}
}

func (r *bcFeatureResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *bcFeatureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.BuildFeatureDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	feature := models.BuildFeatureJson{
		Type: plan.Type.ValueString(),
	}

	if !plan.ID.IsUnknown() && !plan.ID.IsNull() {
		feature.ID = plan.ID.ValueString()
	}

	if !plan.Properties.IsNull() {
		propsMap := make(map[string]string)
		diags = plan.Properties.ElementsAs(ctx, &propsMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		props := make([]models.Property, 0, len(propsMap))
		for k, v := range propsMap {
			props = append(props, models.Property{Name: k, Value: v})
		}
		feature.Properties = &models.Properties{Property: props}
	}

	actual, err := r.client.NewBuildTypeFeature(buildTypeId, feature)
	if err != nil {
		resp.Diagnostics.AddError("Error creating build feature", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	plan.Type = types.StringValue(actual.Type)

	plan.Properties = mergePropertiesFromServer(actual.Properties, plan.Properties, &resp.Diagnostics)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcFeatureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.BuildFeatureDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	featureId := idParts[len(idParts)-1]

	actual, err := r.client.GetBuildTypeFeature(buildTypeId, featureId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading build feature", err.Error())
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	state.Type = types.StringValue(actual.Type)

	state.Properties = mergePropertiesFromServer(actual.Properties, state.Properties, &resp.Diagnostics)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *bcFeatureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.BuildFeatureDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.BuildFeatureDataModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	featureId := idParts[len(idParts)-1]

	feature := models.BuildFeatureJson{
		ID:   featureId,
		Type: plan.Type.ValueString(),
	}

	if !plan.Properties.IsNull() {
		propsMap := make(map[string]string)
		diags = plan.Properties.ElementsAs(ctx, &propsMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		props := make([]models.Property, 0, len(propsMap))
		for k, v := range propsMap {
			props = append(props, models.Property{Name: k, Value: v})
		}
		feature.Properties = &models.Properties{Property: props}
	}

	actual, err := r.client.UpdateBuildTypeFeature(buildTypeId, featureId, feature)
	if err != nil {
		resp.Diagnostics.AddError("Error updating build feature", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	plan.Type = types.StringValue(actual.Type)

	plan.Properties = mergePropertiesFromServer(actual.Properties, plan.Properties, &resp.Diagnostics)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcFeatureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.BuildFeatureDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	featureId := idParts[len(idParts)-1]

	err := r.client.DeleteBuildTypeFeature(buildTypeId, featureId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting build feature", err.Error())
		return
	}
}

func (r *bcFeatureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: build_configuration_id/feature_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("build_configuration_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
