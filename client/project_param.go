package client

import (
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) SetParam(project, name, value string) error {
	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/projects/id:%s/parameters/%s", c.RestURL, project, name),
		strings.NewReader(value),
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

func (c *Client) GetParam(project, name string) (*string, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/projects/id:%s/parameters/%s", c.RestURL, project, name),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.requestWithType(req, "text/plain")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	body := string(resp.Body)
	return &body, nil
}

func (c *Client) DeleteParam(project, name string) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/projects/id:%s/parameters/%s", c.RestURL, project, name),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
