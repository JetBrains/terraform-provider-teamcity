package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"terraform-provider-teamcity/models"
)

const cloudProfileFieldsQuery = "fields=id,name,cloudProviderId,project(id),properties(property(name,value)),images(image(id,name,agentPoolId,properties(property(name,value))))"

func (c *Client) CreateCloudProfile(profile models.CloudProfileJson) (*models.CloudProfileJson, error) {
	body, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}

	var actual models.CloudProfileJson
	if err := c.PostRequest("/cloud/profiles", bytes.NewReader(body), &actual); err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetCloudProfile(id string) (*models.CloudProfileJson, error) {
	var actual models.CloudProfileJson
	endpoint := fmt.Sprintf("/cloud/profiles/id:%s", id)

	if err := c.GetRequest(endpoint, cloudProfileFieldsQuery, &actual); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &actual, nil
}

func (c *Client) UpdateCloudProfile(id string, profile models.CloudProfileJson) (*models.CloudProfileJson, error) {
	body, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}

	var actual models.CloudProfileJson
	endpoint := fmt.Sprintf("/cloud/profiles/id:%s", id)
	if err := c.PutRequest(endpoint, bytes.NewReader(body), &actual); err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) DeleteCloudProfile(id string) error {
	return c.DeleteRequest(fmt.Sprintf("/cloud/profiles/id:%s", id))
}
