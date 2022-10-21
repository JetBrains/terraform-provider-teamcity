package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Project struct {
	Name string  `json:"name"`
	Id   *string `json:"id"`
}

func (c *Client) NewProject(p Project) (*Project, error) {
	rb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := Project{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetProject(id string) (*Project, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := Project{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) RenameProject(id, name string) (*Project, error) {
	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/projects/%s/name", c.HostURL, id),
		strings.NewReader(name),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithType(req, "text/plain")
	if err != nil {
		return nil, err
	}

	actual := Project{
		Name: string(body),
		Id:   &id,
	}

	return &actual, nil
}

func (c *Client) DeleteProject(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/projects/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
