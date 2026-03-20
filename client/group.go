package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"terraform-provider-teamcity/models"
)

func (c *Client) NewGroup(group models.GroupJson) (*models.GroupJson, error) {
	if group.Key == "" {
		group.Key = generateKey(group.Name)
	}

	var actual models.GroupJson
	rb, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	err = c.PostRequest("/userGroups", bytes.NewReader(rb), &actual)
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

func (c *Client) GetGroup(id string) (*models.GroupJson, error) {
	var actual models.GroupJson
	err := c.GetRequest(fmt.Sprintf("/userGroups/%s", id), "", &actual)

	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetGroupByName(name string) (*models.GroupJson, error) {
	encodedName := url.QueryEscape(name)
	var group models.GroupJson
	err := c.GetRequest(fmt.Sprintf("/userGroups/name:%s", encodedName), "", &group)

	if errors.Is(err, ErrNotFound) {
		return nil, errors.New("group not found")
	}
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (c *Client) DeleteGroup(id string) error {
	return c.DeleteRequest(fmt.Sprintf("/userGroups/%s", id))
}

func (c *Client) RemoveGroupRole(groupId, roleId, scope string) error {
	return c.DeleteRequest(fmt.Sprintf("/userGroups/%s/roles/%s/%s", groupId, roleId, scope))
}

func (c *Client) AddGroupRole(groupId, roleId, scope string) error {
	role := models.RoleAssignmentJson{
		Id:    roleId,
		Scope: scope,
	}

	rb, err := json.Marshal(role)
	if err != nil {
		return err
	}

	return c.PostRequest(fmt.Sprintf("/userGroups/%s/roles", groupId), bytes.NewReader(rb), nil)
}

func (c *Client) SetGroupParents(groupId string, parents []string) error {
	groups := models.ParentGroupsJson{}

	for _, i := range parents {
		groups.Group = append(groups.Group, models.GroupJson{Key: i})
	}

	rb, err := json.Marshal(groups)
	if err != nil {
		return err
	}

	return c.PutRequest(fmt.Sprintf("/userGroups/%s/parent-groups", groupId), bytes.NewReader(rb), nil)
}

func (c *Client) AddGroupMember(groupId, username string) error {
	group := models.GroupJson{
		Key: groupId,
	}

	rb, err := json.Marshal(group)
	if err != nil {
		return err
	}

	return c.PostRequest(fmt.Sprintf("/users/username:%s/groups", username), bytes.NewReader(rb), nil)
}

func (c *Client) CheckGroupMember(groupId, username string) (bool, error) {
	err := c.GetRequest(fmt.Sprintf("/users/username:%s/groups/%s", username, groupId), "", nil)

	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Client) DeleteGroupMember(groupId, username string) error {
	return c.DeleteRequest(fmt.Sprintf("/users/username:%s/groups/%s", username, groupId))
}
