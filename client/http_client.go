package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrNotFound for special cases instead of always returning http statusCode.
var ErrNotFound = errors.New("not found")

type Client struct {
	AppURL     string
	RestURL    string
	Token      string
	Username   string
	Password   string
	HTTPClient *http.Client
}

type Response struct {
	StatusCode int
	Body       []byte
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

// Deprecated: Use request instead. Deprecated since version v0.0.69
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	return c.doRequestWithType(req, "application/json")
}

// Deprecated: Use requestWithType instead. Deprecated since version v0.0.69
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
		return Response{}, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read response failed: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return Response{
				StatusCode: res.StatusCode,
				Body:       body,
			},
			ErrNotFound
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

// Verify authethication and status of the REST endpoint
func (c *Client) VerifyConnection(ctx context.Context) (Response, error) {
	addr, err := c.verifyRequestAddr("")
	if err != nil {
		return Response{}, err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr.String(), nil)
	if err != nil {
		return Response{}, err
	}

	// Run text/plain request
	response, err := c.requestWithType(req, "text/plain")
	if err != nil {
		return Response{}, err
	}

	// Verify response: defense against a non caught error when calling requestWithType
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return response, fmt.Errorf("Got status %d when trying connection to the server", response.StatusCode)
	}

	return response, nil
}

// Calling http methods directly. resp must be ready for json.Unmarshall
func (c *Client) GetRequest(ctx context.Context, endpoint, query string, resp any) error {
	addr, err := c.verifyRequestAddr(endpoint)
	if err != nil {
		return err
	}

	// Adding queries
	_, err = url.ParseQuery(query)
	if err != nil {
		return err
	}
	addr.RawQuery = query

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr.String(), nil)
	if err != nil {
		return err
	}

	// Run request
	response, err := c.request(req)
	if err != nil {
		return err
	}

	// Unmarshal the response
	err = json.Unmarshal(response.Body, resp)
	if err != nil {
		return err
	}

	return nil
}

// Calling http methods directly. resp must be ready for json.Unmarshall if the post request returns body
func (c *Client) PostRequest(ctx context.Context, endpoint string, body io.Reader, resp any) error {
	addr, err := c.verifyRequestAddr(endpoint)
	if err != nil {
		return err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, addr.String(), body)
	if err != nil {
		return err
	}

	// Run request
	response, err := c.request(req)
	if err != nil {
		return err
	}

	// Unmarshal the response if it is not empty
	if len(response.Body) != 0 {
		err = json.Unmarshal(response.Body, resp)
		if err != nil {
			return err
		}
	}

	return nil
}

// Calling http methods directly
func (c *Client) DeleteRequest(ctx context.Context, endpoint string) error {
	addr, err := c.verifyRequestAddr(endpoint)
	if err != nil {
		return err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, addr.String(), nil)
	if err != nil {
		return err
	}

	// Run request
	_, err  = c.request(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) verifyRequestAddr(endpoint string) (*url.URL, error) {
	// Build full address and verify it
	addr, err := url.Parse(c.RestURL)
	if err != nil {
		return nil, err
	}
	addr = addr.JoinPath(endpoint)
	return addr, nil
}
