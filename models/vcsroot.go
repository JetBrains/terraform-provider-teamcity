package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type VcsRootJson struct {
	ID                        *string      `json:"id,omitempty"`
	Name                      string       `json:"name,omitempty"`
	VcsName                   string       `json:"vcsName,omitempty"`
	ModificationCheckInterval *int         `json:"modificationCheckInterval,omitempty"`
	Project                   *ProjectJson `json:"project,omitempty"`
	Properties                *Properties  `json:"properties,omitempty"`
}

type VcsRootDataModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ProjectId       types.String `tfsdk:"project_id"`
	PollingInterval types.Int64  `tfsdk:"polling_interval"`
}
