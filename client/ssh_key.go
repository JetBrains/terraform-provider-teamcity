package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SshKeys struct {
	Key []SshKey `json:"sshKey"`
}

type SshKey struct {
	Name string `json:"name"`
}

func (c *Client) NewSshKey(project, name, key string) error {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/projects/id:%s/sshKeys?fileName=%s", c.RestURL, project, name),
		strings.NewReader(key),
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithType(req, "text/plain")
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetSshKeys(projectId string) ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/id:%s/sshKeys", c.RestURL, projectId), nil)
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

	actual := SshKeys{}
	err = json.Unmarshal(resp.Body, &actual)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, i := range actual.Key {
		names = append(names, i.Name)
	}

	return names, nil
}

func (c *Client) DeleteSshKey(projectId, keyName string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/projects/id:%s/sshKeys/%s", c.RestURL, projectId, keyName), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
