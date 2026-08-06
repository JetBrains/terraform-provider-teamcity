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
	_ resource.Resource                = &bcTemplateResource{}
	_ resource.ResourceWithConfigure   = &bcTemplateResource{}
	_ resource.ResourceWithImportState = &bcTemplateResource{}
)

func NewBuildConfigurationTemplateResource() resource.Resource {
	return &bcTemplateResource{}
}

type bcTemplateResource struct {
	client *client.Client
}

func (r *bcTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration_template"
}

func (r *bcTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Attaches an existing build configuration template to a build configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier in the form 'build_configuration_id/template_id'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_configuration_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the build configuration the template is attached to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the template to attach.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *bcTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bcTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.BuildTypeTemplateEntryDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	templateId := plan.TemplateId.ValueString()

	_, err := r.client.AttachBuildTypeTemplate(buildTypeId, templateId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching template",
			"Could not attach template to build configuration: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, templateId))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.BuildTypeTemplateEntryDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	templateId := state.TemplateId.ValueString()

	actual, err := r.client.GetBuildTypeTemplateEntry(buildTypeId, templateId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading template attachment",
			"Could not read template attachment: "+err.Error(),
		)
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, templateId))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *bcTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Both build_configuration_id and template_id require replacement, so
	// there is nothing to update in place.
	var plan models.BuildTypeTemplateEntryDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bcTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.BuildTypeTemplateEntryDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DetachBuildTypeTemplate(state.BuildConfigurationId.ValueString(), state.TemplateId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error detaching template",
			"Could not detach template from build configuration: "+err.Error(),
		)
		return
	}
}

func (r *bcTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: build_configuration_id/template_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("build_configuration_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("template_id"), idParts[1])...)
}
