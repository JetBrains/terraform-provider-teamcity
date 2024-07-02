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

func TestProject(t *testing.T) {
	poolTests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "test-create-project",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)

					defer r.Body.Close()

					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatal(err)
					}
					if testObjJSON != string(body) {
						t.Fatal(fmt.Errorf("received object is not as expected: %s, expected: %s", string(body), testObjJSON))
					}

					w.Write(body)
					if r.URL.Path != objTcEndpoint {
						t.Fatal(fmt.Errorf("wrong url: %s", r.URL.Path))
					}
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "")

				testObj := &models.ProjectJson{}
				err := json.Unmarshal([]byte(testObjJSON), testObj)
				if err != nil {
					t.Fatal(err)
				}
				obj, err := httpClient.NewProject(*testObj)
				if err != nil {
					t.Fatal(err)
				}

				actualObjBytes, err := json.Marshal(obj)
				if err != nil {
					t.Fatal(err)
				}
				if testObjJSON != string(actualObjBytes) {
					t.Fatal(fmt.Errorf("returned object is not as expected: %s, expected: %s", string(actualObjBytes), testObjJSON))
				}
			},
		},
		{
			name: "test-get-project",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(testObjJSON))
					if r.URL.Path != objTcEndpoint+"/id:"+objId {
						t.Fatal(fmt.Errorf("wrong url: %s", r.URL.Path))
					}
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "")

				obj, err := httpClient.GetProject(objId)
				if err != nil {
					t.Fatal(err)
				}

				actualObjBytes, err := json.Marshal(obj)
				if err != nil {
					t.Fatal(err)
				}
				if testObjJSON != string(actualObjBytes) {
					t.Fatal(fmt.Errorf("returned object is not as expected: %s, expected: %s", string(actualObjBytes), testObjJSON))
				}
			},
		},
		{
			name: "test-delete-project",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if r.URL.Path != objTcEndpoint+"/id:"+objId {
						t.Fatal(fmt.Errorf("wrong url: %s", r.URL.Path))
					}
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "")

				err := httpClient.DeleteProject(objId)
				if err != nil {
					t.Fatal(err)
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

var objId = "Test"
var objTcEndpoint = "/app/rest/projects"
var testObjJSON = fmt.Sprintf(`{"name":"Test","id":"%s"}`, objId)
