package client

import (
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

func (c *Client) GetField(resource, id, name string) (*string, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", c.HostURL, resource, id, name),
		nil,
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

// TODO return value without pointer
func (c *Client) SetField(resource, id, name string, value *string) (*string, error) {
	var method, body string
	if value == nil {
		method = "DELETE"
		body = ""
	} else {
		method = "PUT"
		body = *value
	}

	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/%s/%s/%s", c.HostURL, resource, id, name),
		strings.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	result, err := c.doRequestWithType(req, "text/plain")
	if err != nil {
		return nil, err
	}

	val := string(result)
	return &val, nil
}

type Properties struct {
	Property []Property `json:"property"`
}

type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
