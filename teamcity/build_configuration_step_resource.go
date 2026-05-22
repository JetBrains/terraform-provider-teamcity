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
	_ resource.Resource                = &bcStepResource{}
	_ resource.ResourceWithConfigure   = &bcStepResource{}
	_ resource.ResourceWithImportState = &bcStepResource{}
)

func NewBuildConfigurationStepResource() resource.Resource {
	return &bcStepResource{}
}

type bcStepResource struct {
	client *client.Client
}

func (r *bcStepResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration_step"
}

func (r *bcStepResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A build step in a TeamCity build configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (Step ID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the build step.",
			},
			"build_configuration_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the build runner (e.g., simpleRunner, Maven2, Ant).",
			},
			"properties": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Properties for the build runner.",
			},
		},
	}
}

func (r *bcStepResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *bcStepResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.BuildStepDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	step := models.BuildStepJson{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}

	if !plan.ID.IsUnknown() && !plan.ID.IsNull() {
		step.ID = plan.ID.ValueString()
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
		step.Properties = &models.Properties{Property: props}
	}

	actual, err := r.client.NewBuildTypeStep(buildTypeId, step)
	if err != nil {
		resp.Diagnostics.AddError("Error creating build step", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	plan.Name = types.StringValue(actual.Name)
	plan.Type = types.StringValue(actual.Type)

	plan.Properties = mergePropertiesFromServer(actual.Properties, plan.Properties, &resp.Diagnostics)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcStepResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.BuildStepDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	stepId := idParts[len(idParts)-1]

	actual, err := r.client.GetBuildTypeStep(buildTypeId, stepId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading build step", err.Error())
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	state.Name = types.StringValue(actual.Name)
	state.Type = types.StringValue(actual.Type)

	state.Properties = mergePropertiesFromServer(actual.Properties, state.Properties, &resp.Diagnostics)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *bcStepResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.BuildStepDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.BuildStepDataModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	stepId := idParts[len(idParts)-1]

	step := models.BuildStepJson{
		ID:   stepId,
		Name: plan.Name.ValueString(),
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
		step.Properties = &models.Properties{Property: props}
	}

	actual, err := r.client.UpdateBuildTypeStep(buildTypeId, stepId, step)
	if err != nil {
		resp.Diagnostics.AddError("Error updating build step", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	plan.Name = types.StringValue(actual.Name)
	plan.Type = types.StringValue(actual.Type)

	plan.Properties = mergePropertiesFromServer(actual.Properties, plan.Properties, &resp.Diagnostics)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcStepResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.BuildStepDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	stepId := idParts[len(idParts)-1]

	err := r.client.DeleteBuildTypeStep(buildTypeId, stepId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting build step", err.Error())
		return
	}
}

func (r *bcStepResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: build_configuration_id/step_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("build_configuration_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
