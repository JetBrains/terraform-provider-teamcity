package client

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-teamcity/models"
)

func (c *Client) GetPool(name string) (*models.PoolJson, error) {
	var pool models.PoolJson
	// Do GET request
	endpoint := fmt.Sprintf("/agentPools/name:%s", name)
	resp, err := c.GetRequest(context.Background(), endpoint, "", &pool)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	// Return pool
	return &pool, nil
}
