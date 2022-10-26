package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type VcsRoot struct {
	Name       *string        `json:"name"`
	Id         *string        `json:"id"`
	VcsName    string         `json:"vcsName"`
	Project    ProjectLocator `json:"project"`
	Properties VcsProperties  `json:"properties"`
}

type VcsProperties struct {
	Property []VcsProperty `json:"property"`
}

type VcsProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *Client) NewVcsRoot(p VcsRoot) (*VcsRoot, error) {
	rb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/vcs-roots", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := VcsRoot{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetVcsRoot(id string) (*VcsRoot, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/vcs-roots/id:%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := VcsRoot{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) DeleteVcsRoot(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/vcs-roots/id:%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
