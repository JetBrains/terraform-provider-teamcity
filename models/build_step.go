package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type BuildStepJson struct {
	ID         string      `json:"id,omitempty"`
	Name       string      `json:"name,omitempty"`
	Type       string      `json:"type,omitempty"`
	Properties *Properties `json:"properties,omitempty"`
}

type BuildStepDataModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	BuildConfigurationId types.String `tfsdk:"build_configuration_id"`
	Type                 types.String `tfsdk:"type"`
	Properties           types.Map    `tfsdk:"properties"`
}
