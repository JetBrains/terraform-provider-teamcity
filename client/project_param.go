package client

import (
	"errors"
	"fmt"
)

func (c *Client) SetParam(project, name, value string) error {
	// Use SetField to PUT text/plain value; leverages retryableRequestWithType under the hood
	_, err := c.SetField("projects", project, fmt.Sprintf("parameters/%s", name), &value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetParam(project, name string) (*string, error) {
	endpoint := fmt.Sprintf("/projects/id:%s/parameters/%s", project, name)
	body, err := c.GetTextRequest(endpoint, "")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &body, nil
}

func (c *Client) DeleteParam(project, name string) error {
	endpoint := fmt.Sprintf("/projects/id:%s/parameters/%s", project, name)
	return c.DeleteRequest(endpoint)
}
