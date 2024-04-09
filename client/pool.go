package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"terraform-provider-teamcity/models"
	"time"
)

func (c *Client) NewPool(p models.PoolJson) (*models.PoolJson, error) {
	var actual models.PoolJson

	rb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*60))
	defer cancel()

	err = c.PostRequest(ctx, "/agentPools", bytes.NewReader(rb), &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetPool(name string) (*models.PoolJson, error) {
	var pool models.PoolJson
	endpoint := fmt.Sprintf("/agentPools/name:%s", name)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*60))
	defer cancel()

	err := c.GetRequest(ctx, endpoint, "", &pool)

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*60))
	defer cancel()

	err := c.DeleteRequest(ctx, endpoint)
	if err != nil {
		return err
	}

	return nil
}
