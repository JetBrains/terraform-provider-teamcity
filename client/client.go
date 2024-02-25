package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
    "context"
	"strings"
	"time"
)

type Client struct {
	AppURL     string
	RestURL    string
	Token      string
	Username   string
	Password   string
	HTTPClient *http.Client
}

func NewClient(host, token, username, password string) Client {
	client := Client{
		AppURL:     host + "/app",
		RestURL:    host + "/app/rest",
		Token:      token,
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
	return client
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	return c.doRequestWithType(req, "application/json")
}

func (c *Client) doRequestWithType(req *http.Request, ct string) ([]byte, error) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	} else {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.Username+":"+c.Password)))
	}
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
	return c.requestWithType(req, "application/json")
}

func (c *Client) requestWithType(req *http.Request, ct string) (Response, error) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	} else {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.Username+":"+c.Password)))
	}
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Accept", ct)

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
		fmt.Sprintf("%s/%s/%s/%s", c.RestURL, resource, id, name),
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
		fmt.Sprintf("%s/%s/id:%s/%s", c.RestURL, resource, id, name),
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

// Calling http methods directly
// resp must be ready for json.Unmarshall
func (c *Client) GetRequest(ctx context.Context, endpoint, query string, resp any) (Response, error) {
    
    // Build full address and verify it
    addr, err := url.Parse(c.RestURL)
    if    err != nil {
		return Response{}, err
    }
    addr = addr.JoinPath(endpoint)

    // Adding queries
	_, err  = url.ParseQuery(query)
	if err != nil {
		return Response{}, err
	}
	addr.RawQuery = query

    // Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr.String(), nil)
	if err != nil {
		return Response{}, err
	}

    // Run request
    response, err := c.request(req)
    if err != nil {
        return Response{}, err
    }

    // Unmarshal the response
    err = json.Unmarshal(response.Body, resp)
    if err != nil {
        return Response{}, err
    }

    return response, nil 
}
