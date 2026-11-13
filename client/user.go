package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-teamcity/models"
)

type User struct {
	Id         *int64             `json:"id,omitempty"`
	Username   string             `json:"username"`
	Password   *string            `json:"password,omitempty"`
	Roles      *RoleAssignments   `json:"roles,omitempty"`
	Properties *models.Properties `json:"properties,omitempty"`
}

type RoleAssignments struct {
	RoleAssignment []RoleAssignment `json:"role"`
}

type RoleAssignment struct {
	Id    string `json:"roleId"`
	Scope string `json:"scope"`
}

func (c *Client) NewUser(user User) (*User, error) {
	body, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/users", c.RestURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	result, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := User{}
	err = json.Unmarshal(result, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetUser(id string) (*User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/id:%s", c.RestURL, id), nil)
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

	actual := User{}
	err = json.Unmarshal(resp.Body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetUserByName(username string) (*User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/username:%s", c.RestURL, username), nil)
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

	actual := User{}
	err = json.Unmarshal(resp.Body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) SetUser(user User) (*User, error) {
	rb, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/users/id:%d", c.RestURL, *user.Id), bytes.NewReader(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := User{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) DeleteUser(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/users/id:%s", c.RestURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// AddUserRole gives the user one role, leaving the user's other roles alone, so that assignments
// made in parallel do not overwrite each other. Adding a role the user already has succeeds.
func (c *Client) AddUserRole(userId, roleId, scope string) error {
	role := models.RoleAssignmentJson{
		Id:    roleId,
		Scope: scope,
	}

	rb, err := json.Marshal(role)
	if err != nil {
		return err
	}

	return c.PostRequest(fmt.Sprintf("/users/id:%s/roles", userId), bytes.NewReader(rb), nil)
}

// RemoveUserRole takes one role from the user, leaving the user's other roles alone. Removing a role
// the user does not have, or a role of a user that no longer exists, succeeds.
func (c *Client) RemoveUserRole(userId, roleId, scope string) error {
	return c.DeleteRequest(fmt.Sprintf("/users/id:%s/roles/%s/%s", userId, roleId, scope))
}
