package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"terraform-provider-teamcity/models"
)

func (c *Client) NewPool(p models.PoolJson) (*models.PoolJson, error) {
	var actual models.PoolJson

	rb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	err = c.PostRequest("/agentPools", bytes.NewReader(rb), &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetPool(name string) (*models.PoolJson, error) {
	var pool models.PoolJson
	endpoint := fmt.Sprintf("/agentPools/name:%s", name)

	err := c.GetRequest(endpoint, "", &pool)

	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &pool, nil
}

func (c *Client) DeletePool(id string) error {
	endpoint := fmt.Sprintf("/agentPools/id:%s", id)

	err := c.DeleteRequest(endpoint)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SetPoolProjects(name string, p *models.ProjectsJson) (*models.ProjectsJson, error) {
	var actual models.ProjectsJson

	rb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/agentPools/name:%s/projects", name)
	err = c.PutRequest(endpoint, bytes.NewReader(rb), &actual)

	if err != nil {
		return nil, err
	}

	return &actual, nil
}
