package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

type Group struct {
	Key     string           `json:"key"`
	Name    string           `json:"name"`
	Roles   *RoleAssignments `json:"roles,omitempty"`
	Parents *ParentGroups    `json:"parent-groups,omitempty"`
}

type ParentGroups struct {
	Group []Group `json:"group"`
}

func (c *Client) NewGroup(group Group) (*Group, error) {
	if group.Key == "" {
		group.Key = generateKey(group.Name)
	}

	body, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/userGroups", c.RestURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	result, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := Group{}
	err = json.Unmarshal(result, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func generateKey(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToUpper(r)
		}
		return -1
	}, name)
}

func (c *Client) GetGroup(id string) (*Group, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/userGroups/%s", c.RestURL, id), nil)
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

	actual := Group{}
	err = json.Unmarshal(resp.Body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetGroupByName(name string) (*Group, error) {
	encodedName := url.QueryEscape(name)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/userGroups/name:%s", c.RestURL, encodedName), nil)
	if err != nil {
		// If direct name lookup fails, fall back to listing all groups
		return nil, err
	}

	resp, err := c.request(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("group not found")
	}

	var group Group
	err = json.Unmarshal(resp.Body, &group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (c *Client) DeleteGroup(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/userGroups/%s", c.RestURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoveGroupRole(groupId, roleId, scope string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/userGroups/%s/roles/%s/%s", c.RestURL, groupId, roleId, scope), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AddGroupRole(groupId, roleId, scope string) error {
	role := RoleAssignment{
		Id:    roleId,
		Scope: scope,
	}

	body, err := json.Marshal(role)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/userGroups/%s/roles", c.RestURL, groupId), bytes.NewReader(body))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SetGroupParents(groupId string, parents []string) error {
	groups := ParentGroups{}

	for _, i := range parents {
		groups.Group = append(groups.Group, Group{Key: i})
	}

	body, err := json.Marshal(groups)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/userGroups/%s/parent-groups", c.RestURL, groupId), bytes.NewReader(body))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AddGroupMember(groupId, username string) error {
	group := Group{
		Key: groupId,
	}

	rb, err := json.Marshal(group)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/users/username:%s/groups", c.RestURL, username), bytes.NewReader(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CheckGroupMember(groupId, username string) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/username:%s/groups/%s", c.RestURL, username, groupId), nil)
	if err != nil {
		return false, err
	}

	resp, err := c.request(req)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return true, nil
}

func (c *Client) DeleteGroupMember(groupId, username string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/users/username:%s/groups/%s", c.RestURL, username, groupId), nil)
	if err != nil {
		return err
	}

	_, err = c.request(req)
	if err != nil {
		return err
	}

	return nil
}
