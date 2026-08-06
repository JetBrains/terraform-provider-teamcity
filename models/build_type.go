package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type BuildTypesJson struct {
	BuildType []BuildTypeJson `json:"buildType,omitempty"`
}

type BuildTypeJson struct {
	ID           string       `json:"id,omitempty"`
	Name         string       `json:"name,omitempty"`
	ProjectID    string       `json:"projectId,omitempty"`
	Description  string       `json:"description,omitempty"`
	Type         string       `json:"type,omitempty"`
	Paused       bool         `json:"paused,omitempty"`
	TemplateFlag bool         `json:"templateFlag,omitempty"`
	Project      *ProjectJson `json:"project,omitempty"`
}

func (bt *BuildTypeJson) GetProjectID() string {
	if bt.Project != nil && bt.Project.Id != nil {
		return *bt.Project.Id
	}
	return bt.ProjectID
}

type BuildTypeTemplateEntryDataModel struct {
	ID                   types.String `tfsdk:"id"`
	BuildConfigurationId types.String `tfsdk:"build_configuration_id"`
	TemplateId           types.String `tfsdk:"template_id"`
}

type BuildTypeTemplateDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	Description types.String `tfsdk:"description"`
}

type BuildTypeDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	Description types.String `tfsdk:"description"`
	BuildType   types.String `tfsdk:"build_type"`
	Paused      types.Bool   `tfsdk:"paused"`
}
