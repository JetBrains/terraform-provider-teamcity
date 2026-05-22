package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

func (c *Client) NewBuildType(bt models.BuildTypeJson) (*models.BuildTypeJson, error) {
	// For creation, we need to wrap project id into a project object if it's not already there
	if bt.ProjectID != "" && bt.Project == nil {
		id := bt.ProjectID
		bt.Project = &models.ProjectJson{
			Id: &id,
		}
		bt.ProjectID = "" // Clear it so it's not sent twice
	}

	rb, err := json.Marshal(bt)
	if err != nil {
		return nil, err
	}

	var actual models.BuildTypeJson
	if err := c.PostRequest("/buildTypes", bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) GetBuildType(id string) (*models.BuildTypeJson, error) {
	var actual models.BuildTypeJson
	err := c.GetRequest(fmt.Sprintf("/buildTypes/id:%s", id), "", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) UpdateBuildType(id string, bt models.BuildTypeJson) (*models.BuildTypeJson, error) {
	rb, err := json.Marshal(bt)
	if err != nil {
		return nil, err
	}

	var actual models.BuildTypeJson
	if err := c.PutRequest(fmt.Sprintf("/buildTypes/id:%s", id), bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) DeleteBuildType(id string) error {
	return c.DeleteRequest(fmt.Sprintf("/buildTypes/id:%s", id))
}
