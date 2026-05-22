package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type AgentRequirementJson struct {
	ID         string      `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"`
	Properties *Properties `json:"properties,omitempty"`
}

type AgentRequirementDataModel struct {
	ID                   types.String `tfsdk:"id"`
	BuildConfigurationId types.String `tfsdk:"build_configuration_id"`
	Condition            types.String `tfsdk:"condition"`
	Name                 types.String `tfsdk:"name"`
	Value                types.String `tfsdk:"value"`
}
