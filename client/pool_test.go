package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPool(t *testing.T) {
	poolTests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "test-get-pool-requests-virtual-project-fields",
			test: func(t *testing.T) {
				testPoolJSON := `{"name":"Default","id":1,"projects":{"project":[{"name":"Managed","id":"Managed_A"},{"name":"Virtual","id":"Virtual_Project","virtual":true}]}}`

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(testPoolJSON))
					if r.URL.Path != "/app/rest/agentPools/name:Default" {
						t.Fatal(fmt.Errorf("wrong url path: %s", r.URL.Path))
					}
					if r.URL.RawQuery != poolFieldsQuery {
						t.Fatal(fmt.Errorf("wrong query: %s, expected: %s", r.URL.RawQuery, poolFieldsQuery))
					}
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				pool, err := httpClient.GetPool("Default")
				if err != nil {
					t.Fatal(err)
				}

				actualPoolBytes, err := json.Marshal(pool)
				if err != nil {
					t.Fatal(err)
				}
				if testPoolJSON != string(actualPoolBytes) {
					t.Fatal(fmt.Errorf("returned pool is not as expected: %s, expected: %s", string(actualPoolBytes), testPoolJSON))
				}
			},
		},
	}

	for _, tc := range poolTests {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
