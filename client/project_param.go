package client

import (
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

// SetParam sets/updates a regular (text) project parameter value using text/plain PUT.
func (c *Client) SetParam(project, name, value string) error {
	_, err := c.SetField("projects", project, fmt.Sprintf("parameters/%s", name), &value)
	if err != nil {
		return err
	}
	return nil
}

// SecureSetParam sets/updates a secure (password) project parameter using JSON payload
// with type.rawValue set to "password display='normal'" as required by TeamCity REST API.
func (c *Client) SecureSetParam(project, name, value string) error {
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
	_, err := c.SetFieldJson("projects", project, fmt.Sprintf("parameters/%s", name), payload)
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
