package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Role struct {
	Name        *string      `json:"name,omitempty"`
	Id          *string      `json:"id,omitempty"`
	Included    *Included    `json:"included,omitempty"`
	Permissions *Permissions `json:"permissions,omitempty"`
}

type Included struct {
	Role []*Role `json:"role,omitempty"`
}

type Permissions struct {
	Permission []*Permission `json:"permission,omitempty"`
}

type Permission struct {
	Id     *string `json:"id,omitempty"`
	Name   *string `json:"name,omitempty"`
	Global bool    `json:"global,omitempty"`
}

func (c *Client) NewRole(role Role) (*Role, error) {
	body, err := json.Marshal(role)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/roles", c.HostURL), strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	result, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := Role{}
	err = json.Unmarshal(result, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetRole(id string) (*Role, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/roles/id:%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	result, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := Role{}
	err = json.Unmarshal(result, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) DeleteRole(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/roles/id:%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AddIncludedRole(roleId, includedId string) (*Role, error) {
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/roles/id:%s/included/%s", c.HostURL, roleId, includedId), nil)
	if err != nil {
		return nil, err
	}

	result, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := Role{}
	err = json.Unmarshal(result, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) RemoveIncludedRole(roleId, includedId string) (*Role, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/roles/id:%s/included/%s", c.HostURL, roleId, includedId), nil)
	if err != nil {
		return nil, err
	}

	result, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := Role{}
	err = json.Unmarshal(result, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) AddPermission(roleId, permId string) (*Role, error) {
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/roles/id:%s/permissions/%s", c.HostURL, roleId, permId), nil)
	if err != nil {
		return nil, err
	}

	result, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := Role{}
	err = json.Unmarshal(result, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) RemovePermission(roleId, permId string) (*Role, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/roles/id:%s/permissions/%s", c.HostURL, roleId, permId), nil)
	if err != nil {
		return nil, err
	}

	result, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := Role{}
	err = json.Unmarshal(result, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}
