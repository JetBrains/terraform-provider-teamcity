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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &bcTriggerResource{}
	_ resource.ResourceWithConfigure   = &bcTriggerResource{}
	_ resource.ResourceWithImportState = &bcTriggerResource{}
)

func NewBuildConfigurationTriggerResource() resource.Resource {
	return &bcTriggerResource{}
}

type bcTriggerResource struct {
	client *client.Client
}

func (r *bcTriggerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration_trigger"
}

func (r *bcTriggerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A build trigger in a TeamCity build configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (Trigger ID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_configuration_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the build configuration to which this trigger belongs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the build trigger (e.g., vcsTrigger, schedulingTrigger).",
			},
			"properties": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Properties for the build trigger.",
			},
		},
	}
}

func (r *bcTriggerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *bcTriggerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.BuildTriggerDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	trigger := models.BuildTriggerJson{
		Type: plan.Type.ValueString(),
	}

	if !plan.ID.IsUnknown() && !plan.ID.IsNull() {
		trigger.ID = plan.ID.ValueString()
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
		trigger.Properties = &models.Properties{Property: props}
	}

	actual, err := r.client.NewBuildTypeTrigger(buildTypeId, trigger)
	if err != nil {
		resp.Diagnostics.AddError("Error creating build trigger", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	plan.Type = types.StringValue(actual.Type)

	if actual.Properties != nil {
		propsMap := make(map[string]attr.Value)
		for _, p := range actual.Properties.Property {
			propsMap[p.Name] = types.StringValue(p.Value)
		}
		props, diags := types.MapValue(types.StringType, propsMap)
		resp.Diagnostics.Append(diags...)
		if !diags.HasError() {
			plan.Properties = props
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcTriggerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.BuildTriggerDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	triggerId := idParts[len(idParts)-1]

	actual, err := r.client.GetBuildTypeTrigger(buildTypeId, triggerId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading build trigger", err.Error())
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	state.Type = types.StringValue(actual.Type)

	if actual.Properties != nil {
		propsMap := make(map[string]attr.Value)
		for _, p := range actual.Properties.Property {
			propsMap[p.Name] = types.StringValue(p.Value)
		}
		props, diags := types.MapValue(types.StringType, propsMap)
		resp.Diagnostics.Append(diags...)
		if !diags.HasError() {
			state.Properties = props
		}
	} else {
		state.Properties = types.MapNull(types.StringType)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *bcTriggerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.BuildTriggerDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.BuildTriggerDataModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	triggerId := idParts[len(idParts)-1]

	trigger := models.BuildTriggerJson{
		ID:   triggerId,
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
		trigger.Properties = &models.Properties{Property: props}
	}

	actual, err := r.client.UpdateBuildTypeTrigger(buildTypeId, triggerId, trigger)
	if err != nil {
		resp.Diagnostics.AddError("Error updating build trigger", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	plan.Type = types.StringValue(actual.Type)

	if actual.Properties != nil {
		propsMap := make(map[string]attr.Value)
		for _, p := range actual.Properties.Property {
			propsMap[p.Name] = types.StringValue(p.Value)
		}
		props, diags := types.MapValue(types.StringType, propsMap)
		resp.Diagnostics.Append(diags...)
		if !diags.HasError() {
			plan.Properties = props
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcTriggerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.BuildTriggerDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	triggerId := idParts[len(idParts)-1]

	err := r.client.DeleteBuildTypeTrigger(buildTypeId, triggerId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting build trigger", err.Error())
		return
	}
}

func (r *bcTriggerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: build_configuration_id/trigger_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("build_configuration_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
