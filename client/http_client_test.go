package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClient(t *testing.T) {
	clientTests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "test-verify-connection",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"someData": "data"}`))
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "")

				_, err := httpClient.VerifyConnection(context.Background())
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "test-GetRequest-when-not-found",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "")

				body := &struct{}{}
				err := httpClient.GetRequest("", "", body)
				if !errors.Is(err, ErrNotFound) {
					t.Fatal(fmt.Errorf("got wrong error: %w", err))
				}
			},
		},
	}

	for _, tc := range clientTests {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
