package teamcity

import (
	"context"
	"fmt"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &buildConfigurationResource{}
	_ resource.ResourceWithConfigure   = &buildConfigurationResource{}
	_ resource.ResourceWithImportState = &buildConfigurationResource{}
)

func NewBuildConfigurationResource() resource.Resource {
	return &buildConfigurationResource{}
}

type buildConfigurationResource struct {
	client *client.Client
}

func (r *buildConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration"
}

func (r *buildConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A build configuration is a collection of settings used to start a build and group the sequence of the builds. More info [here](https://www.jetbrains.com/help/teamcity/creating-and-editing-build-configurations.html)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "ID of the build configuration. If not provided, it will be generated from the name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the project where the build configuration will be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"build_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Type of the build configuration. Possible values: regular, composite, deployment. Default: regular.",
				Default:             stringdefault.StaticString("regular"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"paused": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the build configuration is paused.",
			},
		},
	}
}

func (r *buildConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *buildConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.BuildTypeDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	btJson := models.BuildTypeJson{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		ProjectID:   plan.ProjectID.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        plan.BuildType.ValueString(),
		Paused:      plan.Paused.ValueBool(),
	}

	result, err := r.client.NewBuildType(btJson)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating build configuration",
			"Could not create build configuration: "+err.Error(),
		)
		return
	}

	// Fetch full data to ensure all fields (like type) are populated
	result, err = r.client.GetBuildType(result.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching build configuration after creation",
			err.Error(),
		)
		return
	}

	r.mapJsonToDataModel(result, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *buildConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.BuildTypeDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetBuildType(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading build configuration",
			"Could not read build configuration: "+err.Error(),
		)
		return
	}

	if result == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapJsonToDataModel(result, &state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *buildConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.BuildTypeDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.BuildTypeDataModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		result, err := r.client.SetField("buildTypes", id, "name", &name)
		if err != nil {
			resp.Diagnostics.AddError("Error updating name", err.Error())
			return
		}
		state.Name = types.StringValue(result)
	}

	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		result, err := r.client.SetField("buildTypes", id, "description", &desc)
		if err != nil {
			resp.Diagnostics.AddError("Error updating description", err.Error())
			return
		}
		state.Description = types.StringValue(result)
	}

	if !plan.Paused.Equal(state.Paused) {
		pausedStr := "false"
		if plan.Paused.ValueBool() {
			pausedStr = "true"
		}
		result, err := r.client.SetField("buildTypes", id, "paused", &pausedStr)
		if err != nil {
			resp.Diagnostics.AddError("Error updating paused", err.Error())
			return
		}
		state.Paused = types.BoolValue(result == "true")
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *buildConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.BuildTypeDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBuildType(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting build configuration",
			"Could not delete build configuration: "+err.Error(),
		)
		return
	}
}

func (r *buildConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *buildConfigurationResource) mapJsonToDataModel(result *models.BuildTypeJson, model *models.BuildTypeDataModel) {
	model.ID = types.StringValue(result.ID)
	model.Name = types.StringValue(result.Name)
	model.ProjectID = types.StringValue(result.GetProjectID())
	model.Description = types.StringValue(result.Description)
	if result.Type != "" {
		model.BuildType = types.StringValue(result.Type)
	} else {
		model.BuildType = types.StringValue("regular")
	}
	model.Paused = types.BoolValue(result.Paused)
}
