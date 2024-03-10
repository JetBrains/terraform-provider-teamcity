package client

import (
	"context"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

func (c *Client) GetPool(name string) (*models.PoolJson, error) {
	var pool models.PoolJson

	endpoint := fmt.Sprintf("/agentPools/name:%s", name)
	err := c.GetRequest(context.Background(), endpoint, "", &pool)

	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &pool, nil
}
