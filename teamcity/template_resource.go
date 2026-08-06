package teamcity

import (
	"context"
	"fmt"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &templateResource{}
	_ resource.ResourceWithConfigure   = &templateResource{}
	_ resource.ResourceWithImportState = &templateResource{}
)

func NewTemplateResource() resource.Resource {
	return &templateResource{}
}

type templateResource struct {
	client *client.Client
}

func (r *templateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A build configuration template is a set of reusable build configuration settings that other build configurations can inherit. More info [here](https://www.jetbrains.com/help/teamcity/build-configuration-template.html)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "ID of the build configuration template. If not provided, it will be generated from the name.",
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
				MarkdownDescription: "ID of the project where the build configuration template will be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
		},
	}
}

func (r *templateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.BuildTypeTemplateDataModel
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
	}

	result, err := r.client.NewBuildTypeTemplate(btJson)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating build configuration template",
			"Could not create build configuration template: "+err.Error(),
		)
		return
	}

	// Fetch full data to ensure all fields are populated
	result, err = r.client.GetBuildTypeTemplate(result.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching build configuration template after creation",
			err.Error(),
		)
		return
	}
	if result == nil {
		resp.Diagnostics.AddError(
			"Error fetching build configuration template after creation",
			"Build configuration template was not found after creation",
		)
		return
	}

	r.mapJsonToDataModel(result, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.BuildTypeTemplateDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetBuildTypeTemplate(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading build configuration template",
			"Could not read build configuration template: "+err.Error(),
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

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.BuildTypeTemplateDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.BuildTypeTemplateDataModel
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

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.BuildTypeTemplateDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBuildTypeTemplate(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting build configuration template",
			"Could not delete build configuration template: "+err.Error(),
		)
		return
	}
}

func (r *templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *templateResource) mapJsonToDataModel(result *models.BuildTypeJson, model *models.BuildTypeTemplateDataModel) {
	model.ID = types.StringValue(result.ID)
	model.Name = types.StringValue(result.Name)
	model.ProjectID = types.StringValue(result.GetProjectID())
	model.Description = types.StringValue(result.Description)
}
