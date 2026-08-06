package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

func (c *Client) AttachBuildTypeTemplate(buildTypeId, templateId string) (*models.BuildTypeJson, error) {
	payload := models.BuildTypeJson{
		ID: templateId,
	}
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var actual models.BuildTypeJson
	if err := c.PostRequest(fmt.Sprintf("/buildTypes/id:%s/templates", buildTypeId), bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) GetBuildTypeTemplateEntry(buildTypeId, templateId string) (*models.BuildTypeJson, error) {
	var actual models.BuildTypeJson
	err := c.GetRequest(fmt.Sprintf("/buildTypes/id:%s/templates/id:%s", buildTypeId, templateId), "fields=id,name,projectId,project(id)", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) DetachBuildTypeTemplate(buildTypeId, templateId string) error {
	err := c.DeleteRequest(fmt.Sprintf("/buildTypes/id:%s/templates/id:%s", buildTypeId, templateId))
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}
