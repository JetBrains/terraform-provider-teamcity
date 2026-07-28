package teamcity

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"
)

var (
	_ resource.Resource                = &cloudProfileResource{}
	_ resource.ResourceWithConfigure   = &cloudProfileResource{}
	_ resource.ResourceWithImportState = &cloudProfileResource{}
)

type cloudProfileResource struct {
	client *client.Client
}

func NewCloudProfileResource() resource.Resource {
	return &cloudProfileResource{}
}

func (r *cloudProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_profile"
}

func (r *cloudProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TeamCity cloud profile and its cloud images.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The TeamCity cloud profile ID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The public cloud profile name displayed in TeamCity.",
			},
			"cloud_provider_id": schema.StringAttribute{
				Required:    true,
				Description: "The TeamCity cloud provider ID, for example amazon.",
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The TeamCity project that owns the cloud profile.",
			},
			"properties": schema.MapAttribute{
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
				Description: "Cloud provider-specific properties. The complete map is sensitive because TeamCity cloud profiles can contain secure values.",
			},
		},
		Blocks: map[string]schema.Block{
			"image": schema.ListNestedBlock{
				Description: "Cloud images managed by this profile.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Description: "The TeamCity cloud image ID.",
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The cloud image name displayed in TeamCity.",
						},
						"properties": schema.MapAttribute{
							Optional:    true,
							Sensitive:   true,
							ElementType: types.StringType,
							Description: "Image-specific provider properties. The complete map is sensitive because it can contain secure values.",
						},
						"agent_pool_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The agent pool that receives agents started from this image.",
						},
					},
				},
			},
		},
	}
}

func (r *cloudProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configuredClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = configuredClient
}

func (r *cloudProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.CloudProfileDataModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profile := r.modelToJSON(ctx, plan, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateCloudProfile(plan.ProjectId.ValueString(), profile)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cloud profile", err.Error())
		return
	}
	if created.Id == "" {
		resp.Diagnostics.AddError("Error creating cloud profile", "TeamCity returned a cloud profile without an ID")
		return
	}

	refreshed, err := r.client.GetCloudProfile(plan.ProjectId.ValueString(), created.Id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading created cloud profile", err.Error())
		return
	}
	if refreshed == nil {
		resp.Diagnostics.AddError("Error reading created cloud profile", "TeamCity did not return the newly created cloud profile")
		return
	}

	r.jsonToModel(ctx, refreshed, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *cloudProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.CloudProfileDataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profile, err := r.client.GetCloudProfile(state.ProjectId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading cloud profile", err.Error())
		return
	}
	if profile == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.jsonToModel(ctx, profile, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *cloudProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.CloudProfileDataModel
	var state models.CloudProfileDataModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Id.IsNull() || state.Id.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Cloud profile ID is unavailable", "The resource cannot update a cloud profile without its TeamCity ID.")
		return
	}

	profile := r.modelToJSON(ctx, plan, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateCloudProfile(state.ProjectId.ValueString(), state.Id.ValueString(), profile)
	if err != nil {
		resp.Diagnostics.AddError("Error updating cloud profile", err.Error())
		return
	}

	profileID := state.Id.ValueString()
	if updated.Id != "" {
		profileID = updated.Id
	}
	refreshed, err := r.client.GetCloudProfile(state.ProjectId.ValueString(), profileID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated cloud profile", err.Error())
		return
	}
	if refreshed == nil {
		resp.Diagnostics.AddError("Error reading updated cloud profile", "TeamCity did not return the updated cloud profile")
		return
	}

	r.jsonToModel(ctx, refreshed, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *cloudProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.CloudProfileDataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.Id.IsNull() || state.Id.IsUnknown() {
		return
	}

	if err := r.client.DeleteCloudProfile(state.ProjectId.ValueString(), state.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting cloud profile", err.Error())
	}
}

func (r *cloudProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid cloud profile import ID", "Use <project_id>/<cloud_profile_feature_id>.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *cloudProfileResource) modelToJSON(ctx context.Context, plan models.CloudProfileDataModel, previous *models.CloudProfileDataModel, diags *diag.Diagnostics) models.CloudProfileJson {
	projectID := plan.ProjectId.ValueString()
	profile := models.CloudProfileJson{
		Name:            plan.Name.ValueString(),
		CloudProviderId: plan.CloudProviderId.ValueString(),
		Project:         &models.CloudProfileProjectJson{Id: &projectID},
		Properties:      propertiesFromMap(ctx, plan.Properties, diags),
		Images:          &models.CloudImagesJson{CloudImage: make([]models.CloudImageJson, 0, len(plan.Images))},
	}

	for index, imagePlan := range plan.Images {
		image := models.CloudImageJson{
			Name:       imagePlan.Name.ValueString(),
			Properties: propertiesFromMap(ctx, imagePlan.Properties, diags),
		}
		if !imagePlan.AgentPoolId.IsNull() && !imagePlan.AgentPoolId.IsUnknown() {
			poolID := int(imagePlan.AgentPoolId.ValueInt64())
			image.AgentPoolId = &poolID
		}
		if previous != nil && index < len(previous.Images) && !previous.Images[index].Id.IsNull() && !previous.Images[index].Id.IsUnknown() {
			image.Id = previous.Images[index].Id.ValueString()
		}
		profile.Images.CloudImage = append(profile.Images.CloudImage, image)
	}

	return profile
}

func (r *cloudProfileResource) jsonToModel(ctx context.Context, profile *models.CloudProfileJson, state *models.CloudProfileDataModel, diags *diag.Diagnostics) {
	state.Id = types.StringValue(profile.Id)
	state.Name = types.StringValue(profile.Name)
	state.CloudProviderId = types.StringValue(profile.CloudProviderId)
	if profile.Project != nil && profile.Project.Id != nil {
		state.ProjectId = types.StringValue(*profile.Project.Id)
	}
	state.Properties = mapFromProperties(ctx, profile.Properties, state.Properties, diags)

	if profile.Images == nil {
		state.Images = nil
		return
	}

	images := make([]models.CloudImageDataModel, 0, len(profile.Images.CloudImage))
	for index, image := range profile.Images.CloudImage {
		var previousImage *models.CloudImageDataModel
		if index < len(state.Images) {
			previousImage = &state.Images[index]
		}

		imageState := models.CloudImageDataModel{
			Id:   types.StringValue(image.Id),
			Name: types.StringValue(image.Name),
		}
		if previousImage == nil {
			imageState.Properties = mapFromProperties(ctx, image.Properties, types.MapNull(types.StringType), diags)
		} else {
			imageState.Properties = mapFromProperties(ctx, image.Properties, previousImage.Properties, diags)
		}
		if image.AgentPoolId != nil {
			imageState.AgentPoolId = types.Int64Value(int64(*image.AgentPoolId))
		} else {
			imageState.AgentPoolId = types.Int64Null()
		}
		images = append(images, imageState)
	}
	state.Images = images
}

func propertiesFromMap(ctx context.Context, value types.Map, diags *diag.Diagnostics) *models.Properties {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	properties := map[string]string{}
	diags.Append(value.ElementsAs(ctx, &properties, false)...)
	if diags.HasError() {
		return nil
	}

	result := &models.Properties{Property: make([]models.Property, 0, len(properties))}
	for name, propertyValue := range properties {
		result.Property = append(result.Property, models.Property{Name: name, Value: propertyValue})
	}
	return result
}

func mapFromProperties(ctx context.Context, properties *models.Properties, previous types.Map, diags *diag.Diagnostics) types.Map {
	if properties == nil || len(properties.Property) == 0 {
		if previous.IsNull() {
			return types.MapNull(types.StringType)
		}
		empty, emptyDiags := types.MapValueFrom(ctx, types.StringType, map[string]string{})
		diags.Append(emptyDiags...)
		return empty
	}

	previousValues := map[string]string{}
	if !previous.IsNull() && !previous.IsUnknown() {
		diags.Append(previous.ElementsAs(ctx, &previousValues, false)...)
		if diags.HasError() {
			return types.MapNull(types.StringType)
		}
	}

	result := make(map[string]string, len(properties.Property))
	for _, property := range properties.Property {
		if strings.HasPrefix(property.Name, "secure:") && property.Value == "" {
			if previousValue, ok := previousValues[property.Name]; ok {
				result[property.Name] = previousValue
				continue
			}
		}
		result[property.Name] = property.Value
	}
	mapped, mapDiags := types.MapValueFrom(ctx, types.StringType, result)
	diags.Append(mapDiags...)
	return mapped
}
