package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

func (c *Client) NewBuildTypeStep(buildTypeId string, step models.BuildStepJson) (*models.BuildStepJson, error) {
	rb, err := json.Marshal(step)
	if err != nil {
		return nil, err
	}

	var actual models.BuildStepJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/steps", buildTypeId)
	if err := c.PostRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) GetBuildTypeStep(buildTypeId, stepId string) (*models.BuildStepJson, error) {
	var actual models.BuildStepJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/steps/%s", buildTypeId, stepId)
	err := c.GetRequest(endpoint, "", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) UpdateBuildTypeStep(buildTypeId, stepId string, step models.BuildStepJson) (*models.BuildStepJson, error) {
	rb, err := json.Marshal(step)
	if err != nil {
		return nil, err
	}

	var actual models.BuildStepJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/steps/%s", buildTypeId, stepId)
	if err := c.PutRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) DeleteBuildTypeStep(buildTypeId, stepId string) error {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/steps/%s", buildTypeId, stepId)
	return c.DeleteRequest(endpoint)
}
