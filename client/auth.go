package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-teamcity/models"
)

type AuthSettings struct {
	AllowGuest            bool    `json:"allowGuest"`
	GuestUsername         string  `json:"guestUsername"`
	WelcomeText           string  `json:"welcomeText"`
	CollapseLoginForm     bool    `json:"collapseLoginForm"`
	TwoFactorMode         string  `json:"twoFactorMode"`
	PerProjectPermissions bool    `json:"perProjectPermissions"`
	EmailVerification     bool    `json:"emailVerification"`
	Modules               Modules `json:"modules"`
}

type Modules struct {
	Module []Module `json:"module"`
}

type Module struct {
	Name       string             `json:"name"`
	Properties *models.Properties `json:"properties,omitempty"`
}

func (c *Client) GetAuthSettings() (AuthSettings, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/authSettings", c.RestURL), nil)
	if err != nil {
		return AuthSettings{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return AuthSettings{}, err
	}

	actual := AuthSettings{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return AuthSettings{}, err
	}

	return actual, nil
}

func (c *Client) SetAuthSettings(settings AuthSettings) (AuthSettings, error) {
	rb, err := json.Marshal(settings)
	if err != nil {
		return AuthSettings{}, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/server/authSettings", c.RestURL), bytes.NewReader(rb))
	if err != nil {
		return AuthSettings{}, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return AuthSettings{}, err
	}

	actual := AuthSettings{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return AuthSettings{}, err
	}

	return actual, nil
}
