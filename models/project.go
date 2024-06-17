package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProjectsJson struct {
	Project []ProjectJson `json:"project,omitempty"`
}

type ProjectJson struct {
	Name            string           `json:"name"`
	Id              *string          `json:"id,omitempty"`
	ProjectFeatures *ProjectFeaturesJson `json:"projectFeatures,omitempty"`
}

type ProjectResourceModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

type ProjectFeaturesJson struct {
	ProjectFeature []ProjectFeatureJson `json:"projectFeature,omitempty"`
}
type ProjectFeatureJson struct {
	Id         *string    `json:"id,omitempty"`
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
}

type VersionedSettingsJson struct {
	SynchronizationMode         string  `json:"synchronizationMode"`
	VcsRootId                   *string `json:"vcsRootId"`
	Format                      *string `json:"format"`
	AllowUIEditing              *bool   `json:"allowUIEditing"`
	StoreSecureValuesOutsideVcs *bool   `json:"storeSecureValuesOutsideVcs"`
	BuildSettingsMode           *string `json:"buildSettingsMode"`
	ShowSettingsChanges         *bool   `json:"showSettingsChanges"`
	ImportDecision              *string `json:"importDecision"`
}

type VersionedSettingsModel struct {
	ProjectId      types.String `tfsdk:"project_id"`
	VcsRoot        types.String `tfsdk:"vcsroot_id"`
	AllowUIEditing types.Bool   `tfsdk:"allow_ui_editing"`
	Settings       types.String `tfsdk:"settings"`
	ShowChanges    types.Bool   `tfsdk:"show_changes"`
}
