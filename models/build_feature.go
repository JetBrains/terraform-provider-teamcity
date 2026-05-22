package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type BuildFeatureJson struct {
	ID         string      `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"`
	Properties *Properties `json:"properties,omitempty"`
}

type BuildFeatureDataModel struct {
	ID                   types.String `tfsdk:"id"`
	BuildConfigurationId types.String `tfsdk:"build_configuration_id"`
	Type                 types.String `tfsdk:"type"`
	Properties           types.Map    `tfsdk:"properties"`
}
