package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"terraform-provider-teamcity/models"
)

func TestBuildTypeTemplate(t *testing.T) {
	templateTests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "test-new-template-sets-template-flag-and-wraps-project",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/app/rest/buildTypes" {
						t.Errorf("wrong url path: %s", r.URL.Path)
					}
					body, _ := io.ReadAll(r.Body)
					var sent models.BuildTypeJson
					if err := json.Unmarshal(body, &sent); err != nil {
						t.Fatal(err)
					}
					if !sent.TemplateFlag {
						t.Error("templateFlag was not set in create payload")
					}
					if sent.ProjectID != "" {
						t.Errorf("projectId should be cleared in favor of project object, got: %s", sent.ProjectID)
					}
					if sent.Project == nil || sent.Project.Id == nil || *sent.Project.Id != "Proj" {
						t.Errorf("project object was not populated correctly: %+v", sent.Project)
					}
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"id":"Proj_Tmpl","name":"tmpl","templateFlag":true,"projectId":"Proj"}`))
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				actual, err := httpClient.NewBuildTypeTemplate(models.BuildTypeJson{
					Name:      "tmpl",
					ProjectID: "Proj",
				})
				if err != nil {
					t.Fatal(err)
				}
				if actual.ID != "Proj_Tmpl" {
					t.Errorf("unexpected template id: %s", actual.ID)
				}
			},
		},
		{
			name: "test-get-returns-nil-for-non-template-build-type",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"id":"Proj_Bc","name":"bc","projectId":"Proj"}`))
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				actual, err := httpClient.GetBuildTypeTemplate("Proj_Bc")
				if err != nil {
					t.Fatal(err)
				}
				if actual != nil {
					t.Errorf("expected nil for a regular build configuration, got: %+v", actual)
				}
			},
		},
		{
			name: "test-get-returns-nil-on-404",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				actual, err := httpClient.GetBuildTypeTemplate("Missing")
				if err != nil {
					t.Fatal(err)
				}
				if actual != nil {
					t.Errorf("expected nil for missing template, got: %+v", actual)
				}
			},
		},
		{
			name: "test-delete-is-idempotent-on-404",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				if err := httpClient.DeleteBuildTypeTemplate("Missing"); err != nil {
					t.Errorf("delete of a missing template should not fail, got: %s", err)
				}
			},
		},
	}

	for _, tt := range templateTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}
