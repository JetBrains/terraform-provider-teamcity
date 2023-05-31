package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type VcsRoot struct {
	Name            string         `json:"name"`
	Id              *string        `json:"id"`
	VcsName         string         `json:"vcsName"`
	PollingInterval *int           `json:"modificationCheckInterval,omitempty"`
	Project         ProjectLocator `json:"project"`
	Properties      Properties     `json:"properties"`
}

func (c *Client) NewVcsRoot(p VcsRoot) (VcsRoot, error) {
	rb, err := json.Marshal(p)
	if err != nil {
		return VcsRoot{}, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/vcs-roots", c.RestURL), bytes.NewReader(rb))
	if err != nil {
		return VcsRoot{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return VcsRoot{}, err
	}

	actual := VcsRoot{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return VcsRoot{}, err
	}

	return actual, nil
}

func (c *Client) GetVcsRoot(id string) (VcsRoot, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/vcs-roots/id:%s", c.RestURL, id), nil)
	if err != nil {
		return VcsRoot{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return VcsRoot{}, err
	}

	actual := VcsRoot{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return VcsRoot{}, err
	}

	return actual, nil
}

func (c *Client) DeleteVcsRoot(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/vcs-roots/id:%s", c.RestURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
