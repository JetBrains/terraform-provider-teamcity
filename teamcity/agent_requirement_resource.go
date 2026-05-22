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
	_ resource.Resource                = &agentRequirementResource{}
	_ resource.ResourceWithConfigure   = &agentRequirementResource{}
	_ resource.ResourceWithImportState = &agentRequirementResource{}
)

func NewAgentRequirementResource() resource.Resource {
	return &agentRequirementResource{}
}

type agentRequirementResource struct {
	client *client.Client
}

func (r *agentRequirementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration_agent_requirement"
}

func (r *agentRequirementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An agent requirement in a TeamCity build configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (Requirement ID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_configuration_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the build configuration to which this requirement belongs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"condition": schema.StringAttribute{
				Required:    true,
				Description: "The condition of the agent requirement (e.g., equals, exists, contains).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the agent parameter to check.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Optional:    true,
				Description: "The value to compare against (not required for all conditions like 'exists').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *agentRequirementResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *agentRequirementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.AgentRequirementDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	ar := models.AgentRequirementJson{
		Type: plan.Condition.ValueString(),
		Properties: &models.Properties{
			Property: []models.Property{
				{Name: "property-name", Value: plan.Name.ValueString()},
			},
		},
	}

	if !plan.Value.IsNull() && !plan.Value.IsUnknown() {
		ar.Properties.Property = append(ar.Properties.Property, models.Property{
			Name:  "property-value",
			Value: plan.Value.ValueString(),
		})
	}

	actual, err := r.client.NewAgentRequirement(buildTypeId, ar)
	if err != nil {
		resp.Diagnostics.AddError("Error creating agent requirement", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	// In case TeamCity changed the condition (unlikely but possible)
	plan.Condition = types.StringValue(actual.Type)

	// Map back properties
	if actual.Properties != nil {
		for _, p := range actual.Properties.Property {
			if p.Name == "property-name" {
				plan.Name = types.StringValue(p.Value)
			}
			if p.Name == "property-value" {
				plan.Value = types.StringValue(p.Value)
			}
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *agentRequirementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.AgentRequirementDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	arId := idParts[len(idParts)-1]

	actual, err := r.client.GetAgentRequirement(buildTypeId, arId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading agent requirement", err.Error())
		return
	}

	if actual == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	state.Condition = types.StringValue(actual.Type)

	if actual.Properties != nil {
		for _, p := range actual.Properties.Property {
			if p.Name == "property-name" {
				state.Name = types.StringValue(p.Value)
			}
			if p.Name == "property-value" {
				state.Value = types.StringValue(p.Value)
			}
		}
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *agentRequirementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.AgentRequirementDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state models.AgentRequirementDataModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := plan.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	arId := idParts[len(idParts)-1]

	ar := models.AgentRequirementJson{
		ID:   arId,
		Type: plan.Condition.ValueString(),
		Properties: &models.Properties{
			Property: []models.Property{
				{Name: "property-name", Value: plan.Name.ValueString()},
			},
		},
	}

	if !plan.Value.IsNull() && !plan.Value.IsUnknown() {
		ar.Properties.Property = append(ar.Properties.Property, models.Property{
			Name:  "property-value",
			Value: plan.Value.ValueString(),
		})
	}

	actual, err := r.client.UpdateAgentRequirement(buildTypeId, arId, ar)
	if err != nil {
		resp.Diagnostics.AddError("Error updating agent requirement", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", buildTypeId, actual.ID))
	plan.Condition = types.StringValue(actual.Type)

	if actual.Properties != nil {
		for _, p := range actual.Properties.Property {
			if p.Name == "property-name" {
				plan.Name = types.StringValue(p.Value)
			}
			if p.Name == "property-value" {
				plan.Value = types.StringValue(p.Value)
			}
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *agentRequirementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.AgentRequirementDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildTypeId := state.BuildConfigurationId.ValueString()
	idParts := strings.Split(state.ID.ValueString(), "/")
	arId := idParts[len(idParts)-1]

	err := r.client.DeleteAgentRequirement(buildTypeId, arId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting agent requirement", err.Error())
		return
	}
}

func (r *agentRequirementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: build_configuration_id/requirement_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("build_configuration_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
