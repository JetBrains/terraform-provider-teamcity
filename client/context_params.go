package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ContextParams struct {
	Params []Property `json:"versionedSettingsContextParameter"`
}

func (c *Client) SetContextParams(project string, params map[string]string) (map[string]string, error) {
	body := ContextParams{}
	body.Params = make([]Property, 0)
	for k, v := range params {
		body.Params = append(body.Params, Property{
			Name:  k,
			Value: v,
		})
	}
	rb, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/projects/id:%s/versionedSettings/contextParameters", c.RestURL, project),
		bytes.NewReader(rb),
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := ContextParams{}
	err = json.Unmarshal(resp, &actual)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, param := range actual.Params {
		m[param.Name] = param.Value
	}

	return m, nil
}

func (c *Client) GetContextParams(project string) (map[string]string, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/projects/id:%s/versionedSettings/contextParameters", c.RestURL, project),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := ContextParams{}
	err = json.Unmarshal(resp, &actual)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, param := range actual.Params {
		m[param.Name] = param.Value
	}

	return m, nil
}
