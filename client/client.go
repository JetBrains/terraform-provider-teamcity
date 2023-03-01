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

func NewClient(host, token string) Client {
	client := Client{
		HostURL:    host + "/app/rest",
		Token:      token,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	return client
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

type Response struct {
	StatusCode int
	Body       []byte
}

// TODO replace other methods
func (c *Client) request(req *http.Request) (Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, err
	}

	if res.StatusCode == http.StatusNotFound {
		return Response{
			StatusCode: res.StatusCode,
			Body:       body,
		}, nil
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return Response{}, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return Response{
		StatusCode: res.StatusCode,
		Body:       body,
	}, nil
}

func (c *Client) GetField(resource, id, name string) (string, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/%s/%s/%s", c.HostURL, resource, id, name),
		nil,
	)
	if err != nil {
		return "", err
	}

	body, err := c.doRequestWithType(req, "text/plain")
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Client) SetField(resource, id, name string, value *string) (string, error) {
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
		fmt.Sprintf("%s/%s/id:%s/%s", c.HostURL, resource, id, name),
		strings.NewReader(body),
	)
	if err != nil {
		return "", err
	}

	result, err := c.doRequestWithType(req, "text/plain")
	if err != nil {
		return "", err
	}

	return string(result), nil
}

type Properties struct {
	Property []Property `json:"property"`
}

type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
