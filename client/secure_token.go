package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"terraform-provider-teamcity/models"
)

type SecureTokens struct {
	Tokens []models.Property `json:"versionedSettingsToken"`
}

func (c *Client) AddSecureToken(project, value string) (*string, error) {
	id := "credentialsJSON:" + uuid.New().String()
	body := SecureTokens{
		Tokens: []models.Property{
			{
				Name:  id,
				Value: value,
			},
		},
	}

	rb, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/projects/id:%s/versionedSettings/tokens", c.RestURL, project),
		bytes.NewReader(rb),
	)
	if err != nil {
		return nil, err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (c *Client) GetSecureTokens(project string) ([]string, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/projects/id:%s/versionedSettings/tokens", c.RestURL, project),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := SecureTokens{}
	err = json.Unmarshal(resp, &actual)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, i := range actual.Tokens {
		names = append(names, i.Name)
	}

	return names, nil
}

func (c *Client) DeleteSecureToken(project, id string) error {
	body := SecureTokens{
		Tokens: []models.Property{
			{
				Name: id,
			},
		},
	}

	rb, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/projects/id:%s/versionedSettings/tokens", c.RestURL, project),
		bytes.NewReader(rb),
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
