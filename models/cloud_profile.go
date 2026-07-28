package models

// CloudProfileJson is the TeamCity REST representation of a cloud profile.
type CloudProfileJson struct {
	Id              string           `json:"id,omitempty"`
	Name            string           `json:"name,omitempty"`
	CloudProviderId string           `json:"cloudProviderId,omitempty"`
	Project         *ProjectJson     `json:"project,omitempty"`
	Properties      *Properties      `json:"properties,omitempty"`
	Images          *CloudImagesJson `json:"images,omitempty"`
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
