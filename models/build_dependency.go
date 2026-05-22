package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type SourceBuildTypeJson struct {
	ID string `json:"id,omitempty"`
}

type SnapshotDependencyJson struct {
	ID              string               `json:"id,omitempty"`
	Type            string               `json:"type,omitempty"`
	SourceBuildType *SourceBuildTypeJson `json:"source-buildType,omitempty"`
	Properties      *Properties          `json:"properties,omitempty"`
}

type SnapshotDependencyDataModel struct {
	ID                   types.String `tfsdk:"id"`
	BuildConfigurationId types.String `tfsdk:"build_configuration_id"`
	DependsOnId          types.String `tfsdk:"depends_on_id"`
	Properties           types.Map    `tfsdk:"properties"`
}

type ArtifactDependencyJson struct {
	ID              string               `json:"id,omitempty"`
	Type            string               `json:"type,omitempty"`
	SourceBuildType *SourceBuildTypeJson `json:"source-buildType,omitempty"`
	Properties      *Properties          `json:"properties,omitempty"`
}

type ArtifactDependencyDataModel struct {
	ID                   types.String `tfsdk:"id"`
	BuildConfigurationId types.String `tfsdk:"build_configuration_id"`
	DependsOnId          types.String `tfsdk:"depends_on_id"`
	Properties           types.Map    `tfsdk:"properties"`
}
