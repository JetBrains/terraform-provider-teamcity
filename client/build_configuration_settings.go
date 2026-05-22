package client

import (
	"errors"
	"fmt"
)

// SetBuildTypeSetting sets a build configuration setting value.
func (c *Client) SetBuildTypeSetting(buildTypeId, name, value string) error {
	_, err := c.SetField("buildTypes", buildTypeId, fmt.Sprintf("settings/%s", name), &value)
	return err
}

// GetBuildTypeSetting retrieves a build configuration setting value.
func (c *Client) GetBuildTypeSetting(buildTypeId, name string) (*string, error) {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/settings/%s", buildTypeId, name)
	body, err := c.GetTextRequest(endpoint, "")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &body, nil
}
