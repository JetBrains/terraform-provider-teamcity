package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

func (c *Client) NewBuildTypeFeature(buildTypeId string, feature models.BuildFeatureJson) (*models.BuildFeatureJson, error) {
	rb, err := json.Marshal(feature)
	if err != nil {
		return nil, err
	}

	var actual models.BuildFeatureJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/features", buildTypeId)
	if err := c.PostRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) GetBuildTypeFeature(buildTypeId, featureId string) (*models.BuildFeatureJson, error) {
	var actual models.BuildFeatureJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/features/%s", buildTypeId, featureId)
	err := c.GetRequest(endpoint, "", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) UpdateBuildTypeFeature(buildTypeId, featureId string, feature models.BuildFeatureJson) (*models.BuildFeatureJson, error) {
	rb, err := json.Marshal(feature)
	if err != nil {
		return nil, err
	}

	var actual models.BuildFeatureJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/features/%s", buildTypeId, featureId)
	if err := c.PutRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) DeleteBuildTypeFeature(buildTypeId, featureId string) error {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/features/%s", buildTypeId, featureId)
	return c.DeleteRequest(endpoint)
}
