package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailSettings struct {
	Enabled          bool    `json:"enabled"`
	Host             string  `json:"host"`
	Port             int     `json:"port"`
	From             string  `json:"from"`
	Login            string  `json:"login"`
	Password         *string `json:"password"`
	SecureConnection string  `json:"secureConnection"`
}

func (c *Client) GetEmailSettings() (*EmailSettings, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/email/rest", c.AppURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := EmailSettings{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) SetEmailSettings(settings EmailSettings) (*EmailSettings, error) {
	rb, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/email/rest", c.AppURL), bytes.NewReader(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := EmailSettings{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}
