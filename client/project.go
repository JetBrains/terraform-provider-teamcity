package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-teamcity/models"
)

type Project struct {
	Name            string           `json:"name"`
	Id              *string          `json:"id,omitempty"`
	ParentProject   *Project         `json:"parentProject,omitempty"`
	ProjectFeatures *ProjectFeatures `json:"projectFeatures,omitempty"`
}

type ProjectFeatures struct {
	ProjectFeature []ProjectFeature `json:"projectFeature,omitempty"`
}
type ProjectFeature struct {
	Id         *string           `json:"id,omitempty"`
	Type       string            `json:"type"`
	Properties models.Properties `json:"properties"`
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

// TODO: refactor other methods in the same way
func (c *Client) NewProject(p Project) (Project, error) {
	rb, err := json.Marshal(p)
	if err != nil {
		return Project{}, err
	}

	var newPool = Project{}
	endpoint := "/projects"
	err = c.PostRequest(endpoint, bytes.NewReader(rb), &newPool)
	if err != nil {
		return Project{}, err
	}

	return newPool, nil
}

func (c *Client) GetProject(id string) (*Project, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/id:%s", c.RestURL, id), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	actual := Project{}
	err = json.Unmarshal(resp.Body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) DeleteProject(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/projects/id:%s", c.RestURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) NewProjectFeature(id string, feature ProjectFeature) (ProjectFeature, error) {
	rb, err := json.Marshal(feature)
	if err != nil {
		return ProjectFeature{}, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects/id:%s/projectFeatures", c.RestURL, id), bytes.NewReader(rb))
	if err != nil {
		return ProjectFeature{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return ProjectFeature{}, err
	}

	actual := ProjectFeature{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return ProjectFeature{}, err
	}

	return actual, nil
}

func (c *Client) GetProjectFeature(projectId, featureId string) (*ProjectFeature, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/id:%s/projectFeatures/id:%s", c.RestURL, projectId, featureId), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	actual := ProjectFeature{}
	err = json.Unmarshal(resp.Body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) DeleteProjectFeature(projectId, featureId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/projects/id:%s/projectFeatures/id:%s", c.RestURL, projectId, featureId), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

type ProjectLocator struct {
	Id string `json:"id"`
}

func (c *Client) GetVersionedSettings(projectId string) (*VersionedSettings, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/id:%s/versionedSettings/config", c.RestURL, projectId), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	actual := VersionedSettings{}
	err = json.Unmarshal(resp.Body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) SetVersionedSettings(projectId string, settings VersionedSettings) (*VersionedSettings, error) {
	rb, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/projects/id:%s/versionedSettings/config", c.RestURL, projectId), bytes.NewReader(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := VersionedSettings{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}
