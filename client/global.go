package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GlobalSettings struct {
	ArtifactDirectories            string `json:"artifactDirectories"`
	RootUrl                        string `json:"rootUrl"`
	MaxArtifactSize                int64  `json:"maxArtifactSize"`
	MaxArtifactNumber              int64  `json:"maxArtifactsNumber"`
	DefaultExecutionTimeout        int64  `json:"defaultExecutionTimeout"`
	DefaultVCSCheckInterval        int64  `json:"defaultVCSCheckInterval"`
	EnforceDefaultVCSCheckInterval bool   `json:"enforceDefaultVCSCheckInterval"`
	DefaultQuietPeriod             int64  `json:"defaultQuietPeriod"`
	UseEncryption                  bool   `json:"useEncryption"`
	EncryptionKey                  string `json:"encryptionKey"`
	ArtifactsDomainIsolation       bool   `json:"artifactsDomainIsolation"`
	ArtifactsUrl                   string `json:"artifactsUrl"`
}

func (c *Client) GetGlobalSettings() (*GlobalSettings, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/globalSettings", c.RestURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := GlobalSettings{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}

func (c *Client) SetGlobalSettings(settings GlobalSettings) (*GlobalSettings, error) {
	rb, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/server/globalSettings", c.RestURL), bytes.NewReader(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	actual := GlobalSettings{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		return nil, err
	}

	return &actual, nil
}
