package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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
