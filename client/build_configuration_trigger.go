package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

func (c *Client) NewBuildTypeTrigger(buildTypeId string, trigger models.BuildTriggerJson) (*models.BuildTriggerJson, error) {
	rb, err := json.Marshal(trigger)
	if err != nil {
		return nil, err
	}

	var actual models.BuildTriggerJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/triggers", buildTypeId)
	if err := c.PostRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) GetBuildTypeTrigger(buildTypeId, triggerId string) (*models.BuildTriggerJson, error) {
	var actual models.BuildTriggerJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/triggers/%s", buildTypeId, triggerId)
	err := c.GetRequest(endpoint, "", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) UpdateBuildTypeTrigger(buildTypeId, triggerId string, trigger models.BuildTriggerJson) (*models.BuildTriggerJson, error) {
	rb, err := json.Marshal(trigger)
	if err != nil {
		return nil, err
	}

	var actual models.BuildTriggerJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/triggers/%s", buildTypeId, triggerId)
	if err := c.PutRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) DeleteBuildTypeTrigger(buildTypeId, triggerId string) error {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/triggers/%s", buildTypeId, triggerId)
	return c.DeleteRequest(endpoint)
}
