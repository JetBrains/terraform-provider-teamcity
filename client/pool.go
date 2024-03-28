package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"terraform-provider-teamcity/models"
)

func (c *Client) NewPool(p models.PoolJson) (*models.PoolJson, error) {

	rb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/agentPools", c.RestURL)

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := models.PoolJson{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetPool(name string) (*models.PoolJson, error) {

	endpoint := fmt.Sprintf("%s/agentPools/name:%s", c.RestURL, name)

	req, err := http.NewRequest("GET", endpoint, nil)
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

	actual := models.PoolJson{}
	err = json.Unmarshal(resp.Body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) DeletePool(id string) error {

	endpoint := fmt.Sprintf("%s/agentPools/id:%s", c.RestURL, id)

	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.request(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return nil
}
