package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	HostURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(host, token *string) (*Client, error) {
	c := Client{
		HostURL:    *host + "/app/rest",
		Token:      *token,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	return &c, nil
}

type CleanupSettings struct {
	Enabled     bool          `json:"enabled"`
	MaxDuration int           `json:"maxCleanupDuration"`
	Daily       *CleanupDaily `json:"daily"`
	Cron        *CleanupCron  `json:"cron"`
}

type CleanupDaily struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

type CleanupCron struct {
	Minute  string `json:"minute"`
	Hour    string `json:"hour"`
	Day     string `json:"day"`
	Month   string `json:"month"`
	DayWeek string `json:"dayWeek"`
}

func (c *Client) GetVersion() (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/version", c.HostURL), nil)
	if err != nil {
		return "", err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Client) SetCleanup(settings CleanupSettings) (*CleanupSettings, error) {
	rb, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/server/cleanup", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := CleanupSettings{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) GetCleanup() (*CleanupSettings, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/cleanup", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := CleanupSettings{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	return c.doRequestWithType(req, "application/json")
}

func (c *Client) doRequestWithType(req *http.Request, ct string) ([]byte, error) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Accept", "application/json, text/plain")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

func (c *Client) SetParameter(resource, id, name, value string) (*string, error) {
	var method string
	if value == "" {
		method = "DELETE"
	} else {
		method = "PUT"
	}

	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/%s/%s/%s", c.HostURL, resource, id, name),
		strings.NewReader(value),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithType(req, "text/plain")
	if err != nil {
		return nil, err
	}

	result := string(body)

	return &result, nil
}
