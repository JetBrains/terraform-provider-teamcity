package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Project struct {
	Name string  `json:"name"`
	Id   *string `json:"id"`
}

func (c *Client) NewProject(p Project) (Project, error) {
	rb, err := json.Marshal(p)
	if err != nil {
		return Project{}, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects", c.HostURL), bytes.NewReader(rb))
	if err != nil {
		return Project{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return Project{}, err
	}

	actual := Project{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return Project{}, err
	}

	return actual, nil
}

func (c *Client) GetProject(id string) (Project, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%s", c.HostURL, id), nil)
	if err != nil {
		return Project{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return Project{}, err
	}

	actual := Project{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return Project{}, err
	}

	return actual, nil
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

type ProjectLocator struct {
	Id string `json:"id"`
}
