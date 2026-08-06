package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"terraform-provider-teamcity/models"
)

func TestBuildTypeTemplateEntry(t *testing.T) {
	entryTests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "test-attach-posts-template-id",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost {
						t.Errorf("expected POST, got: %s", r.Method)
					}
					if r.URL.Path != "/app/rest/buildTypes/id:Proj_Bc/templates" {
						t.Errorf("wrong url path: %s", r.URL.Path)
					}
					body, _ := io.ReadAll(r.Body)
					var sent models.BuildTypeJson
					if err := json.Unmarshal(body, &sent); err != nil {
						t.Fatal(err)
					}
					if sent.ID != "Proj_Tmpl" {
						t.Errorf("wrong template id in payload: %s", sent.ID)
					}
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"id":"Proj_Tmpl","name":"tmpl","templateFlag":true,"projectId":"Proj"}`))
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				actual, err := httpClient.AttachBuildTypeTemplate("Proj_Bc", "Proj_Tmpl")
				if err != nil {
					t.Fatal(err)
				}
				if actual.ID != "Proj_Tmpl" {
					t.Errorf("unexpected template id: %s", actual.ID)
				}
			},
		},
		{
			name: "test-get-entry-returns-nil-when-not-attached",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/app/rest/buildTypes/id:Proj_Bc/templates/id:Proj_Tmpl" {
						t.Errorf("wrong url path: %s", r.URL.Path)
					}
					w.WriteHeader(http.StatusNotFound)
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				actual, err := httpClient.GetBuildTypeTemplateEntry("Proj_Bc", "Proj_Tmpl")
				if err != nil {
					t.Fatal(err)
				}
				if actual != nil {
					t.Errorf("expected nil for detached template, got: %+v", actual)
				}
			},
		},
		{
			name: "test-detach-is-idempotent-on-404",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodDelete {
						t.Errorf("expected DELETE, got: %s", r.Method)
					}
					if r.URL.Path != "/app/rest/buildTypes/id:Proj_Bc/templates/id:Proj_Tmpl" {
						t.Errorf("wrong url path: %s", r.URL.Path)
					}
					w.WriteHeader(http.StatusNotFound)
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				if err := httpClient.DetachBuildTypeTemplate("Proj_Bc", "Proj_Tmpl"); err != nil {
					t.Errorf("detach of a missing attachment should not fail, got: %s", err)
				}
			},
		},
	}

	for _, tt := range entryTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}
