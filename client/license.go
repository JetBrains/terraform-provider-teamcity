package client

import (
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) NewLicense(key string) error {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/server/licensingData/licenseKeys", c.RestURL),
		strings.NewReader(key),
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithType(req, "text/plain")
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CheckLicense(key string) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/licensingData/licenseKeys/%s", c.RestURL, key), nil)
	if err != nil {
		return false, err
	}

	resp, err := c.request(req)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return true, nil
}

func (c *Client) DeleteLicense(key string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/server/licensingData/licenseKeys/%s", c.RestURL, key), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
