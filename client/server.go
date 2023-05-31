package client

import (
	"fmt"
	"net/http"
)

func (c *Client) GetVersion() (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/version", c.RestURL), nil)
	if err != nil {
		return "", err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
