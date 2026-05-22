package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

func (c *Client) NewBuildTypeVcsRootEntry(buildTypeId string, entry models.VcsRootEntryJson) (*models.VcsRootEntryJson, error) {
	rb, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	var actual models.VcsRootEntryJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/vcs-root-entries", buildTypeId)
	if err := c.PostRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) GetBuildTypeVcsRootEntry(buildTypeId, vcsRootId string) (*models.VcsRootEntryJson, error) {
	var actual models.VcsRootEntryJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/vcs-root-entries/%s", buildTypeId, vcsRootId)
	err := c.GetRequest(endpoint, "", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) UpdateBuildTypeVcsRootEntry(buildTypeId, vcsRootId string, entry models.VcsRootEntryJson) (*models.VcsRootEntryJson, error) {
	rb, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	var actual models.VcsRootEntryJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/vcs-root-entries/%s", buildTypeId, vcsRootId)
	if err := c.PutRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) DeleteBuildTypeVcsRootEntry(buildTypeId, vcsRootId string) error {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/vcs-root-entries/%s", buildTypeId, vcsRootId)
	return c.DeleteRequest(endpoint)
}
