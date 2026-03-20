package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"terraform-provider-teamcity/models"
	"testing"
)

func TestGroup(t *testing.T) {
	groupTests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "test-get-group-with-description",
			test: func(t *testing.T) {
				testGroupJSON := `{"key":"TEST_GROUP","name":"Test Group","description":"A group for testing"}`

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(testGroupJSON))
					if r.URL.Path != "/app/rest/userGroups/TEST_GROUP" {
						t.Fatal(fmt.Errorf("wrong url: %s", r.URL.Path))
					}
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				group, err := httpClient.GetGroup("TEST_GROUP")
				if err != nil {
					t.Fatal(err)
				}

				if group.Description != "A group for testing" {
					t.Fatalf("expected description 'A group for testing', got '%s'", group.Description)
				}

				actualGroupBytes, err := json.Marshal(group)
				if err != nil {
					t.Fatal(err)
				}
				if testGroupJSON != string(actualGroupBytes) {
					t.Fatal(fmt.Errorf("returned group is not as expected: %s, expected: %s", string(actualGroupBytes), testGroupJSON))
				}
			},
		},
		{
			name: "test-new-group-with-description",
			test: func(t *testing.T) {
				testGroupJSON := `{"key":"NEW_GROUP","name":"New Group","description":"A new group"}`

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(testGroupJSON))
					if r.Method != "POST" {
						t.Fatal(fmt.Errorf("expected POST, got %s", r.Method))
					}
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				newGroup := models.GroupJson{
					Key:         "NEW_GROUP",
					Name:        "New Group",
					Description: "A new group",
				}

				actual, err := httpClient.NewGroup(newGroup)
				if err != nil {
					t.Fatal(err)
				}

				if actual.Description != "A new group" {
					t.Fatalf("expected description 'A new group', got '%s'", actual.Description)
				}
			},
		},
	}

	for _, tc := range groupTests {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
