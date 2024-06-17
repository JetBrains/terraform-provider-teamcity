package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-teamcity/models"
)

// TODO: refactor other methods in the same way
func (c *Client) NewProject(p models.ProjectJson) (models.ProjectJson, error) {
	rb, err := json.Marshal(p)
	if err != nil {
		return models.ProjectJson{}, err
	}

	var newPool = models.ProjectJson{}
	endpoint := "/projects"
	err = c.PostRequest(endpoint, bytes.NewReader(rb), &newPool)
	if err != nil {
		return models.ProjectJson{}, err
	}

	return newPool, nil
}

func (c *Client) GetProject(id string) (*models.ProjectJson, error) {
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

	actual := models.ProjectJson{}
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

func (c *Client) NewProjectFeature(id string, feature models.ProjectFeatureJson) (models.ProjectFeatureJson, error) {
	rb, err := json.Marshal(feature)
	if err != nil {
		return models.ProjectFeatureJson{}, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects/id:%s/projectFeatures", c.RestURL, id), bytes.NewReader(rb))
	if err != nil {
		return models.ProjectFeatureJson{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return models.ProjectFeatureJson{}, err
	}

	actual := models.ProjectFeatureJson{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return models.ProjectFeatureJson{}, err
	}

	return actual, nil
}

func (c *Client) GetProjectFeature(projectId, featureId string) (*models.ProjectFeatureJson, error) {
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

	actual := models.ProjectFeatureJson{}
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

func (c *Client) GetVersionedSettings(projectId string) (*models.VersionedSettingsJson, error) {
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

	actual := models.VersionedSettingsJson{}
	err = json.Unmarshal(resp.Body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) SetVersionedSettings(projectId string, settings models.VersionedSettingsJson) (*models.VersionedSettingsJson, error) {
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

	actual := models.VersionedSettingsJson{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}
