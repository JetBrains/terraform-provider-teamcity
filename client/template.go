package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

func (c *Client) NewBuildTypeTemplate(bt models.BuildTypeJson) (*models.BuildTypeJson, error) {
	// For creation, we need to wrap project id into a project object if it's not already there
	if bt.ProjectID != "" && bt.Project == nil {
		id := bt.ProjectID
		bt.Project = &models.ProjectJson{
			Id: &id,
		}
		bt.ProjectID = "" // Clear it so it's not sent twice
	}
	bt.TemplateFlag = true

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

func (c *Client) GetBuildTypeTemplate(id string) (*models.BuildTypeJson, error) {
	var actual models.BuildTypeJson
	err := c.GetRequest(fmt.Sprintf("/buildTypes/id:%s", id), "fields=id,name,projectId,project(id),description,templateFlag", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// The id namespace is shared with build configurations; only report templates
	if !actual.TemplateFlag {
		return nil, nil
	}
	return &actual, nil
}

func (c *Client) DeleteBuildTypeTemplate(id string) error {
	err := c.DeleteRequest(fmt.Sprintf("/buildTypes/id:%s", id))
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}
