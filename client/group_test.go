package client

import (
	"encoding/json"
	"fmt"
	"io"
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
		{
			name: "test-new-group-generates-key-from-name",
			test: func(t *testing.T) {
				var sentBody []byte

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					sentBody, _ = io.ReadAll(r.Body)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"key":"MY_TESTGROUP_2","name":"My Test-Group 2!"}`))
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				// No key supplied, so it has to be derived from the name:
				// spaces become underscores, letters are upper-cased, digits are
				// kept and anything else is dropped.
				actual, err := httpClient.NewGroup(models.GroupJson{Name: "My Test-Group 2!"})
				if err != nil {
					t.Fatal(err)
				}

				var sent models.GroupJson
				if err := json.Unmarshal(sentBody, &sent); err != nil {
					t.Fatal(err)
				}
				if sent.Key != "MY_TESTGROUP_2" {
					t.Fatalf("expected generated key 'MY_TESTGROUP_2' to be sent, got '%s'", sent.Key)
				}
				if sent.Name != "My Test-Group 2!" {
					t.Fatalf("expected name to be sent unchanged, got '%s'", sent.Name)
				}
				if actual.Key != "MY_TESTGROUP_2" {
					t.Fatalf("expected returned key 'MY_TESTGROUP_2', got '%s'", actual.Key)
				}
			},
		},
		{
			name: "test-new-group-keeps-explicit-key",
			test: func(t *testing.T) {
				var sentBody []byte

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					sentBody, _ = io.ReadAll(r.Body)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"key":"custom_key","name":"My Test Group"}`))
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				_, err := httpClient.NewGroup(models.GroupJson{Key: "custom_key", Name: "My Test Group"})
				if err != nil {
					t.Fatal(err)
				}

				var sent models.GroupJson
				if err := json.Unmarshal(sentBody, &sent); err != nil {
					t.Fatal(err)
				}
				if sent.Key != "custom_key" {
					t.Fatalf("expected explicit key 'custom_key' to be preserved, got '%s'", sent.Key)
				}
			},
		},
		{
			name: "test-check-group-member",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/rest/users/username:testuser/groups/TEST_GROUP" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"key":"TEST_GROUP"}`))
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				// Test found
				ok, err := httpClient.CheckGroupMember("TEST_GROUP", "testuser")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !ok {
					t.Fatal("expected ok to be true")
				}

				// Test not found
				ok, err = httpClient.CheckGroupMember("NON_EXISTENT", "testuser")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if ok {
					t.Fatal("expected ok to be false")
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
