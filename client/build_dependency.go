package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"terraform-provider-teamcity/models"
)

// Snapshot Dependencies

func (c *Client) NewSnapshotDependency(buildTypeId string, dep models.SnapshotDependencyJson) (*models.SnapshotDependencyJson, error) {
	dep.Type = "snapshot_dependency"
	rb, err := json.Marshal(dep)
	if err != nil {
		return nil, err
	}

	var actual models.SnapshotDependencyJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/snapshot-dependencies", buildTypeId)
	if err := c.PostRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) GetSnapshotDependency(buildTypeId, depId string) (*models.SnapshotDependencyJson, error) {
	var actual models.SnapshotDependencyJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/snapshot-dependencies/%s", buildTypeId, depId)
	err := c.GetRequest(endpoint, "", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) UpdateSnapshotDependency(buildTypeId, depId string, dep models.SnapshotDependencyJson) (*models.SnapshotDependencyJson, error) {
	dep.Type = "snapshot_dependency"
	rb, err := json.Marshal(dep)
	if err != nil {
		return nil, err
	}

	var actual models.SnapshotDependencyJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/snapshot-dependencies/%s", buildTypeId, depId)
	if err := c.PutRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) DeleteSnapshotDependency(buildTypeId, depId string) error {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/snapshot-dependencies/%s", buildTypeId, depId)
	return c.DeleteRequest(endpoint)
}

// Artifact Dependencies

func (c *Client) NewArtifactDependency(buildTypeId string, dep models.ArtifactDependencyJson) (*models.ArtifactDependencyJson, error) {
	dep.Type = "artifact_dependency"
	rb, err := json.Marshal(dep)
	if err != nil {
		return nil, err
	}

	var actual models.ArtifactDependencyJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/artifact-dependencies", buildTypeId)
	if err := c.PostRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) GetArtifactDependency(buildTypeId, depId string) (*models.ArtifactDependencyJson, error) {
	var actual models.ArtifactDependencyJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/artifact-dependencies/%s", buildTypeId, depId)
	err := c.GetRequest(endpoint, "", &actual)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) UpdateArtifactDependency(buildTypeId, depId string, dep models.ArtifactDependencyJson) (*models.ArtifactDependencyJson, error) {
	dep.Type = "artifact_dependency"
	rb, err := json.Marshal(dep)
	if err != nil {
		return nil, err
	}

	var actual models.ArtifactDependencyJson
	endpoint := fmt.Sprintf("/buildTypes/id:%s/artifact-dependencies/%s", buildTypeId, depId)
	if err := c.PutRequest(endpoint, bytes.NewReader(rb), &actual); err != nil {
		return nil, err
	}
	return &actual, nil
}

func (c *Client) DeleteArtifactDependency(buildTypeId, depId string) error {
	endpoint := fmt.Sprintf("/buildTypes/id:%s/artifact-dependencies/%s", buildTypeId, depId)
	return c.DeleteRequest(endpoint)
}
