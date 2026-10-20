package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildTypeSettings(t *testing.T) {
	t.Run("SetBuildTypeSetting_puts_text_value", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			if r.URL.Path != "/app/rest/buildTypes/id:Test_Bc/settings/artifactRules" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			if string(body) != "+:target/*.jar" {
				t.Errorf("unexpected body: %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		}))
		defer server.Close()

		c := NewClient(server.URL, "token", "", "", 0)
		if err := c.SetBuildTypeSetting("Test_Bc", "artifactRules", "+:target/*.jar"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("GetBuildTypeSetting_returns_value", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/app/rest/buildTypes/id:Test_Bc/settings/buildNumberPattern" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("%build.counter%"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "token", "", "", 0)
		val, err := c.GetBuildTypeSetting("Test_Bc", "buildNumberPattern")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val == nil || *val != "%build.counter%" {
			t.Errorf("expected '%%build.counter%%', got %v", val)
		}
	})

	t.Run("GetBuildTypeSetting_returns_nil_on_404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		c := NewClient(server.URL, "token", "", "", 0)
		val, err := c.GetBuildTypeSetting("Missing_Bc", "buildNumberPattern")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil on 404, got %v", *val)
		}
	})
}
