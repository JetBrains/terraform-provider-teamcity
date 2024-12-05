package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"terraform-provider-teamcity/models"
)

func (c *Client) NewProject(p models.ProjectJson) (models.ProjectJson, error) {
	rb, err := json.Marshal(p)
	if err != nil {
		return models.ProjectJson{}, err
	}

	var newProject = models.ProjectJson{}
	endpoint := "/projects"
	err = c.PostRequest(endpoint, bytes.NewReader(rb), &newProject)
	if err != nil {
		return models.ProjectJson{}, err
	}

	return newProject, nil
}

func (c *Client) GetProject(id string) (*models.ProjectJson, error) {
	var actual models.ProjectJson
	endpoint := fmt.Sprintf("/projects/id:%s", id)

	err := c.GetRequest(endpoint, "", &actual)

	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) DeleteProject(id string) error {
	endpoint := fmt.Sprintf("/projects/id:%s", id)

	err := c.DeleteRequest(endpoint)
	if err != nil {
		return err
	}

	return nil
}

// TODO: refactor other methods in the same way as the New/Get/DeleteProject
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

	res, err := c.request(req)
	if err != nil {
		return nil, err
	}

	actual := models.VersionedSettingsJson{}
	err = json.Unmarshal(res.Body, &actual)
	if err != nil {
		return nil, err
	}

	//TODO TeamCIty rest api bug TW-90967, remove when fixed
	if settings.ShowSettingsChanges != nil && actual.ShowSettingsChanges != settings.ShowSettingsChanges {
		corrected, err := c.SetVersionedSettingsProperty(projectId, "showSettingsChanges", strconv.FormatBool(*settings.ShowSettingsChanges))
		if err != nil {
			return nil, fmt.Errorf("could not correct showSettingsChanges property, potentially not enough time for VersionedSettings to be applied: %w", err)
		}
		var correctedShowSettingsChanges bool
		err = json.Unmarshal(corrected, &correctedShowSettingsChanges)
		if err != nil {
			return nil, fmt.Errorf("malformed showSettingsChanges property, more info: %w", err)
		}
		actual.ShowSettingsChanges = &correctedShowSettingsChanges
	}

	return &actual, nil
}

func (c *Client) SetVersionedSettingsProperty(projectId string, property string, value string) ([]byte, error) {
	paramValue := []byte(value)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/projects/id:%s/versionedSettings/config/parameters/%s", c.RestURL, projectId, property), bytes.NewReader(paramValue))
	if err != nil {
		return nil, err
	}

	resp, err := c.retryableRequestWithType(req, "text/plain", retryPolicy)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func retryPolicy(_ context.Context, resp *http.Response, _ error) (bool, error) {
	if resp.StatusCode == http.StatusInternalServerError { // wait the main object/feature to be ready for a property change
		return true, nil
	} else {
		return false, nil
	}
}
