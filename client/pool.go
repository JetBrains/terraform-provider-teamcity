package client

import (
	"bytes"
	"encoding/json"
	"context"
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

    err = c.PostRequest(context.Background(), "/agentPools", bytes.NewReader(rb), &actual)
    if err != nil {
        return nil, err
    }

    return &actual, nil
}


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

func (c *Client) DeletePool(id string) error {
	endpoint := fmt.Sprintf("/agentPools/id:%s", id)

	err := c.DeleteRequest(context.Background(), endpoint)
	if err != nil {
		return err
	}

    return nil
}
