package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// CloudProfileJson is the TeamCity REST representation of a cloud profile.
type CloudProfileJson struct {
	Id              string                   `json:"id,omitempty"`
	Name            string                   `json:"name,omitempty"`
	CloudProviderId string                   `json:"cloudProviderId,omitempty"`
	Project         *CloudProfileProjectJson `json:"project,omitempty"`
	Properties      *Properties              `json:"properties,omitempty"`
	Images          *CloudImagesJson         `json:"images,omitempty"`
}

// CloudProfileProjectJson identifies the TeamCity project that owns a cloud profile.
// It deliberately does not reuse ProjectJson because the cloud-profile REST payload
// only needs an ID and must not send ProjectJson.Name as an empty required field.
type CloudProfileProjectJson struct {
	Id *string `json:"id,omitempty"`
}

// CloudImagesJson is the TeamCity REST collection wrapper for cloud images.
type CloudImagesJson struct {
	CloudImage []CloudImageJson `json:"image,omitempty"`
}

// CloudImageJson is the TeamCity REST representation of an image in a cloud profile.
type CloudImageJson struct {
	Id          string      `json:"id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Properties  *Properties `json:"properties,omitempty"`
	AgentPoolId *int        `json:"agentPoolId,omitempty"`
}

// CloudProfileDataModel is the Terraform state and plan model for a cloud profile.
type CloudProfileDataModel struct {
	Id              types.String          `tfsdk:"id"`
	Name            types.String          `tfsdk:"name"`
	CloudProviderId types.String          `tfsdk:"cloud_provider_id"`
	ProjectId       types.String          `tfsdk:"project_id"`
	Properties      types.Map             `tfsdk:"properties"`
	Images          []CloudImageDataModel `tfsdk:"image"`
}

// CloudImageDataModel is the Terraform nested state and plan model for a cloud image.
type CloudImageDataModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Properties  types.Map    `tfsdk:"properties"`
	AgentPoolId types.Int64  `tfsdk:"agent_pool_id"`
}
