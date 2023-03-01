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
		fmt.Sprintf("%s/projects/id:%s/sshKeys?fileName=%s", c.HostURL, project, name),
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

func (c *Client) GetSshKeys(id string) ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/id:%s/sshKeys", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := SshKeys{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, i := range actual.Key {
		names = append(names, i.Name)
	}

	return names, nil
}
