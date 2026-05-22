package teamcity

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &artifactDependencyResource{}
	_ resource.ResourceWithConfigure   = &artifactDependencyResource{}
	_ resource.ResourceWithImportState = &artifactDependencyResource{}
)

func NewArtifactDependencyResource() resource.Resource {
	return &artifactDependencyResource{}
}

type artifactDependencyResource struct {
	client *client.Client
}

func (r *artifactDependencyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration_artifact_dependency"
}

func (r *artifactDependencyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An artifact dependency in a TeamCity build configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (Artifact Dependency ID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_configuration_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the build configuration to which this dependency belongs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"depends_on_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the build configuration on which this one depends.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"properties": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Properties for the artifact dependency (e.g., pathRules, revisionName, revisionValue).",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *artifactDependencyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *artifactDependencyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.ArtifactDependencyDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	dep := models.ArtifactDependencyJson{
		SourceBuildType: &models.SourceBuildTypeJson{
			ID: plan.DependsOnId.ValueString(),
		},
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
		dep.Properties = &models.Properties{Property: props}
	}

	actual, err := r.client.NewArtifactDependency(buildTypeId, dep)
	if err != nil {
		resp.Diagnostics.AddError("Error creating artifact dependency", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	plan.DependsOnId = types.StringValue(actual.SourceBuildType.ID)

	props, err := r.filterProperties(ctx, actual.Properties, plan.Properties)
	if err != nil {
		resp.Diagnostics.AddError("Error filtering properties", err.Error())
		return
	}
	plan.Properties = props

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *artifactDependencyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.ArtifactDependencyDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	depId := idParts[len(idParts)-1]

	actual, err := r.client.GetArtifactDependency(buildTypeId, depId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading artifact dependency", err.Error())
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	state.DependsOnId = types.StringValue(actual.SourceBuildType.ID)

	props, err := r.filterProperties(ctx, actual.Properties, state.Properties)
	if err != nil {
		resp.Diagnostics.AddError("Error filtering properties", err.Error())
		return
	}
	state.Properties = props

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *artifactDependencyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.ArtifactDependencyDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.ArtifactDependencyDataModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	depId := idParts[len(idParts)-1]

	dep := models.ArtifactDependencyJson{
		ID: depId,
		SourceBuildType: &models.SourceBuildTypeJson{
			ID: plan.DependsOnId.ValueString(),
		},
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
		dep.Properties = &models.Properties{Property: props}
	}

	actual, err := r.client.UpdateArtifactDependency(buildTypeId, depId, dep)
	if err != nil {
		resp.Diagnostics.AddError("Error updating artifact dependency", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	plan.DependsOnId = types.StringValue(actual.SourceBuildType.ID)

	props, err := r.filterProperties(ctx, actual.Properties, plan.Properties)
	if err != nil {
		resp.Diagnostics.AddError("Error filtering properties", err.Error())
		return
	}
	plan.Properties = props

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *artifactDependencyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.ArtifactDependencyDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	depId := idParts[len(idParts)-1]

	err := r.client.DeleteArtifactDependency(buildTypeId, depId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting artifact dependency", err.Error())
		return
	}
}

func (r *artifactDependencyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: build_configuration_id/artifact_dependency_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("build_configuration_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *artifactDependencyResource) filterProperties(ctx context.Context, actual *models.Properties, requested types.Map) (types.Map, error) {
	if actual == nil {
		return types.MapNull(types.StringType), nil
	}

	propsMap := make(map[string]attr.Value)
	if requested.IsNull() || requested.IsUnknown() {
		// During Import or if not specified, we take everything from the server
		for _, p := range actual.Property {
			propsMap[p.Name] = types.StringValue(p.Value)
		}
	} else {
		requestedMap := make(map[string]string)
		if diags := requested.ElementsAs(ctx, &requestedMap, false); diags.HasError() {
			return types.MapNull(types.StringType), fmt.Errorf("%v", diags)
		}

		for _, p := range actual.Property {
			if _, ok := requestedMap[p.Name]; ok {
				propsMap[p.Name] = types.StringValue(p.Value)
			}
		}
	}

	if len(propsMap) == 0 && (requested.IsNull() || requested.IsUnknown()) {
		return types.MapNull(types.StringType), nil
	}

	res, diags := types.MapValue(types.StringType, propsMap)
	if diags.HasError() {
		return types.MapNull(types.StringType), fmt.Errorf("error creating map: %v", diags)
	}
	return res, nil
}
