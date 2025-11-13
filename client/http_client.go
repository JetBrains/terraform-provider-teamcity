package client

import (
	"bytes"
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

	"github.com/hashicorp/go-retryablehttp"
)

// ErrNotFound for special cases instead of always returning http statusCode.
var ErrNotFound = errors.New("not found")

const requestsTimeoutSec = 30

type Client struct {
	AppURL     string
	RestURL    string
	Token      string
	Username   string
	Password   string
	HTTPClient *http.Client
	MaxRetries int
}

type Response struct {
	StatusCode int
	Body       []byte
}

func NewClient(host, token, username, password string, maxRetries int) Client {
	client := Client{
		AppURL:     host + "/app",
		RestURL:    host + "/app/rest",
		Token:      token,
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: requestsTimeoutSec * time.Second},
		MaxRetries: maxRetries,
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
	c.setHeaders(req, ct)

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
			nil
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return Response{}, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return Response{
		StatusCode: res.StatusCode,
		Body:       body,
	}, nil
}

// retryableRequest performs an HTTP request with retry logic using the provided retry policy and request object.
//
// Parameters:
//   - req: The HTTP request to be executed with retry logic.
//   - retryPolicy: A function defining the retry policy to be applied for the request.
//
// Returns:
//   - Response: The response object containing the status code and body of the final request.
//   - error: An error object if the request fails after retrying or other errors occur.
func (c *Client) retryableRequest(req *http.Request, retryPolicy retryablehttp.CheckRetry) (Response, error) {
	return c.retryableRequestWithType(req, "application/json", retryPolicy)
}

func (c *Client) retryableRequestWithType(req *http.Request, ct string, retryPolicy retryablehttp.CheckRetry) (Response, error) {
	c.setHeaders(req, ct)

	rclient := retryablehttp.NewClient()
	rclient.RetryWaitMin = 5 * time.Second
	rclient.RetryWaitMax = 5 * time.Second
	rclient.RetryMax = c.MaxRetries
	rclient.CheckRetry = retryPolicy

	// Convert http.Request to retryablehttp request
	retryReq, err := retryablehttp.NewRequest(req.Method, req.URL.String(), req.Body)
	retryReq.Header = req.Header

	res, err := rclient.Do(retryReq)
	if err != nil {
		return Response{}, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read response failed: %w", err)
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

	result, err := c.retryableRequestWithType(req, "text/plain", retryPolicy)
	if err != nil {
		return "", err
	}

	return string(result.Body), nil
}

func (c *Client) SetFieldJson(resource, id, name string, value interface{}) (string, error) {
	var method string
	var body []byte
	var err error

	if value == nil {
		method = "DELETE"
		body = make([]byte, 0)
	} else {
		method = "PUT"
		body, err = json.Marshal(value)
		if err != nil {
			return "", err
		}
	}

	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/%s/id:%s/%s", c.RestURL, resource, id, name),
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}

	result, err := c.requestWithType(req, "application/json")
	if err != nil {
		return "", err
	}

	return string(result.Body), nil
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

// Uses default Background context. Use GetRequestWithContext for custom context. resp must be ready for json.Unmarshall
func (c *Client) GetRequest(endpoint, query string, resp any) error {
	ctx := context.Background()
	return c.GetRequestWithContext(ctx, endpoint, query, resp)
}

// Calling http methods directly. resp must be ready for json.Unmarshall
func (c *Client) GetRequestWithContext(ctx context.Context, endpoint, query string, resp any) error {
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
	if response.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	// Unmarshal the response
	err = json.Unmarshal(response.Body, resp)
	if err != nil {
		return err
	}

	return nil
}

// Uses default Background context. Use GetTextRequestWithContext for custom context. Returns body as string.
func (c *Client) GetTextRequest(endpoint, query string) (string, error) {
	ctx := context.Background()
	return c.GetTextRequestWithContext(ctx, endpoint, query)
}

// Calling http methods directly for text/plain endpoints. Returns body as string.
func (c *Client) GetTextRequestWithContext(ctx context.Context, endpoint, query string) (string, error) {
	addr, err := c.verifyRequestAddr(endpoint)
	if err != nil {
		return "", err
	}

	// Adding queries
	_, err = url.ParseQuery(query)
	if err != nil {
		return "", err
	}
	addr.RawQuery = query

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr.String(), nil)
	if err != nil {
		return "", err
	}

	// Run request as text/plain
	response, err := c.requestWithType(req, "text/plain")
	if err != nil {
		return "", err
	}
	if response.StatusCode == http.StatusNotFound {
		return "", ErrNotFound
	}

	return string(response.Body), nil
}

// Uses default Background context. Use PostRequestWithContext for custom context. resp must be ready for json.Unmarshall
func (c *Client) PostRequest(endpoint string, body io.Reader, resp any) error {
	ctx := context.Background()
	return c.PostRequestWithContext(ctx, endpoint, body, resp)
}

// Calling http methods directly. resp must be ready for json.Unmarshall if the post request returns body
func (c *Client) PostRequestWithContext(ctx context.Context, endpoint string, body io.Reader, resp any) error {
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

// Uses default Background context. Use DeleteRequestWithContext for custom context. resp must be ready for json.Unmarshall
func (c *Client) DeleteRequest(endpoint string) error {
	ctx := context.Background()
	return c.DeleteRequestWithContext(ctx, endpoint)
}

// Calling http methods directly
func (c *Client) DeleteRequestWithContext(ctx context.Context, endpoint string) error {
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
	_, err = c.request(req)
	if err != nil {
		return err
	}

	return nil
}

// Uses default Background context. Use PutRequestWithContext for custom context. resp must be ready for json.Unmarshall
func (c *Client) PutRequest(endpoint string, body io.Reader, resp any) error {
	ctx := context.Background()
	return c.PutRequestWithContext(ctx, endpoint, body, resp)
}

// Calling http methods directly. resp must be ready for json.Unmarshall if the put request returns body
func (c *Client) PutRequestWithContext(ctx context.Context, endpoint string, body io.Reader, resp any) error {
	addr, err := c.verifyRequestAddr(endpoint)
	if err != nil {
		return err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, addr.String(), body)
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

func (c *Client) verifyRequestAddr(endpoint string) (*url.URL, error) {
	// Build full address and verify it
	addr, err := url.Parse(c.RestURL)
	if err != nil {
		return nil, err
	}
	addr = addr.JoinPath(endpoint)
	return addr, nil
}

func (c *Client) setHeaders(req *http.Request, ct string) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	} else {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.Username+":"+c.Password)))
	}
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Accept", ct)
}
