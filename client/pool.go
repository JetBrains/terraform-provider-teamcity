package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"terraform-provider-teamcity/models"
)

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
