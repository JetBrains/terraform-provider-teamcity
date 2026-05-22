package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

func (c *Client) NewAgentRequirement(buildTypeId string, ar models.AgentRequirementJson) (*models.AgentRequirementJson, error) {
	rb, err := json.Marshal(ar)
	if err != nil {
		return nil, err
	}

	var actual models.AgentRequirementJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/agent-requirements", buildTypeId)
	if err := c.PostRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) GetAgentRequirement(buildTypeId, arId string) (*models.AgentRequirementJson, error) {
	var actual models.AgentRequirementJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/agent-requirements/%s", buildTypeId, arId)
	err := c.GetRequest(endpoint, "", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) UpdateAgentRequirement(buildTypeId, arId string, ar models.AgentRequirementJson) (*models.AgentRequirementJson, error) {
	rb, err := json.Marshal(ar)
	if err != nil {
		return nil, err
	}

	var actual models.AgentRequirementJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/agent-requirements/%s", buildTypeId, arId)
	if err := c.PutRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) DeleteAgentRequirement(buildTypeId, arId string) error {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/agent-requirements/%s", buildTypeId, arId)
	return c.DeleteRequest(endpoint)
}
