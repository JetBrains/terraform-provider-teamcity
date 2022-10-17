package client

import (
	"fmt"
	"io"
	"net/http"
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

func (c *Client) GetVersion() (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/version", c.HostURL), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", c.Token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return string(body), nil
}
