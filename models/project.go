package models

type ProjectsJson struct {
	Project []Project `json:"project,omitempty"`
}

type Project struct {
	Name            string           `json:"name"`
	Id              *string          `json:"id,omitempty"`
	ProjectFeatures *ProjectFeatures `json:"projectFeatures,omitempty"`
}

type ProjectFeatures struct {
	ProjectFeature []ProjectFeature `json:"projectFeature,omitempty"`
}
type ProjectFeature struct {
	Id         *string    `json:"id,omitempty"`
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
}

type VersionedSettings struct {
	SynchronizationMode         string  `json:"synchronizationMode"`
	VcsRootId                   *string `json:"vcsRootId"`
	Format                      *string `json:"format"`
	AllowUIEditing              *bool   `json:"allowUIEditing"`
	StoreSecureValuesOutsideVcs *bool   `json:"storeSecureValuesOutsideVcs"`
	BuildSettingsMode           *string `json:"buildSettingsMode"`
	ShowSettingsChanges         *bool   `json:"showSettingsChanges"`
	ImportDecision              *string `json:"importDecision"`
}
