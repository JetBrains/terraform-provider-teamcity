package client

import (
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

// SetBuildTypeParam sets/updates a regular (text) build configuration parameter value using text/plain PUT.
func (c *Client) SetBuildTypeParam(buildTypeId, name, value string) error {
	_, err := c.SetField("buildTypes", buildTypeId, fmt.Sprintf("parameters/%s", name), &value)
	if err != nil {
		return err
	}
	return nil
}

// SecureSetBuildTypeParam sets/updates a secure (password) build configuration parameter using JSON payload
// with type.rawValue set to "password display='normal'" as required by TeamCity REST API.
func (c *Client) SecureSetBuildTypeParam(buildTypeId, name, value string) error {
	payload := struct {
		Name      string `json:"name"`
		Value     string `json:"value"`
		Inherited bool   `json:"inherited"`
		Type      *struct {
			RawValue string `json:"rawValue"`
		} `json:"type,omitempty"`
	}{
		Name:      name,
		Value:     value,
		Inherited: false,
		Type: &struct {
			RawValue string `json:"rawValue"`
		}{RawValue: models.SecureParamRawType},
	}

	// Use JSON PUT to the parameters endpoint
	_, err := c.SetFieldJson("buildTypes", buildTypeId, fmt.Sprintf("parameters/%s", name), payload)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetBuildTypeParam(buildTypeId, name string) (*string, error) {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/parameters/%s", buildTypeId, name)
	body, err := c.GetTextRequest(endpoint, "")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &body, nil
}

func (c *Client) DeleteBuildTypeParam(buildTypeId, name string) error {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/parameters/%s", buildTypeId, name)
	return c.DeleteRequest(endpoint)
}
