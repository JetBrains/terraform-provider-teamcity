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

var objId = "Test"
var objTcEndpoint = "/app/rest/projects"
var testObjJSON = fmt.Sprintf(`{"name":"Test","id":"%s"}`, objId)

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

				httpClient := NewClient(server.URL, "token", "", "", 12)

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

				httpClient := NewClient(server.URL, "token", "", "", 12)

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
			name: "test-get-project-404",
			test: func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					if r.URL.Path != objTcEndpoint+"/id:"+objId {
						t.Fatal(fmt.Errorf("wrong url: %s", r.URL.Path))
					}
				}))
				defer server.Close()

				httpClient := NewClient(server.URL, "token", "", "", 12)

				obj, err := httpClient.GetProject(objId)
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				if obj != nil {
					t.Fatalf("expected nil object on 404, got: %#v", obj)
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

				httpClient := NewClient(server.URL, "token", "", "", 12)

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

func TestProjectFeaturesAndVersionedSettings(t *testing.T) {
	const featureId = "Feat1"

	// Prepare common feature payload
	featurePayload := models.ProjectFeatureJson{
		Type:       "myFeature",
		Properties: models.Properties{Property: []models.Property{{Name: "k", Value: "v"}}},
	}
	featurePayloadJSON, _ := json.Marshal(featurePayload)

	// Versioned settings payloads
	vcsId := "VCS1"
	format := "Kotlin"
	allow := true
	storeOutside := false
	mode := "USE_PROJECT"
	showChangesTrue := true
	showChangesFalse := false
	importDecision := "APPLY"

	vsRequested := models.VersionedSettingsJson{
		SynchronizationMode:         "enabled",
		VcsRootId:                   &vcsId,
		Format:                      &format,
		AllowUIEditing:              &allow,
		StoreSecureValuesOutsideVcs: &storeOutside,
		BuildSettingsMode:           &mode,
		ShowSettingsChanges:         &showChangesTrue,
		ImportDecision:              &importDecision,
	}
	vsRequestedJSON, _ := json.Marshal(vsRequested)

	vsReturnedNoCorrection := vsRequested // identical
	vsReturnedNoCorrectionJSON, _ := json.Marshal(vsReturnedNoCorrection)

	vsReturnedNeedsCorrection := vsRequested
	vsReturnedNeedsCorrection.ShowSettingsChanges = &showChangesFalse
	vsReturnedNeedsCorrectionJSON, _ := json.Marshal(vsReturnedNeedsCorrection)

	t.Run("project-features: create success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/projectFeatures" && r.Method == http.MethodPost {
				defer r.Body.Close()
				b, _ := io.ReadAll(r.Body)
				if string(b) != string(featurePayloadJSON) {
					t.Fatalf("unexpected body: %s", string(b))
				}
				w.WriteHeader(http.StatusOK)
				w.Write(b)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		actual, err := httpClient.NewProjectFeature(objId, featurePayload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, _ := json.Marshal(actual)
		if string(got) != string(featurePayloadJSON) {
			t.Fatalf("unexpected result: %s", string(got))
		}
	})

	t.Run("project-features: create server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("boom"))
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		_, err := httpClient.NewProjectFeature(objId, featurePayload)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("project-features: get success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/projectFeatures/id:"+featureId && r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				w.Write(featurePayloadJSON)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		actual, err := httpClient.GetProjectFeature(objId, featureId)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if actual == nil {
			t.Fatalf("expected feature, got nil")
		}
		got, _ := json.Marshal(actual)
		if string(got) != string(featurePayloadJSON) {
			t.Fatalf("unexpected result: %s", string(got))
		}
	})

	t.Run("project-features: get 404 returns nil", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		actual, err := httpClient.GetProjectFeature(objId, featureId)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if actual != nil {
			t.Fatalf("expected nil on 404, got: %#v", actual)
		}
	})

	t.Run("project-features: delete success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/projectFeatures/id:"+featureId && r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		if err := httpClient.DeleteProjectFeature(objId, featureId); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("project-features: delete server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("boom"))
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		if err := httpClient.DeleteProjectFeature(objId, featureId); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("versioned settings: get success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/versionedSettings/config" && r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				w.Write(vsReturnedNoCorrectionJSON)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		actual, err := httpClient.GetVersionedSettings(objId)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if actual == nil {
			t.Fatalf("expected non-nil settings")
		}
		got, _ := json.Marshal(actual)
		if string(got) != string(vsReturnedNoCorrectionJSON) {
			t.Fatalf("unexpected result: %s", string(got))
		}
	})

	t.Run("versioned settings: get 404 returns nil", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		actual, err := httpClient.GetVersionedSettings(objId)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if actual != nil {
			t.Fatalf("expected nil, got: %#v", actual)
		}
	})

	t.Run("versioned settings: set success (may call correction path)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/versionedSettings/config" && r.Method == http.MethodPut {
				defer r.Body.Close()
				b, _ := io.ReadAll(r.Body)
				if string(b) != string(vsRequestedJSON) {
					t.Fatalf("unexpected PUT body: %s", string(b))
				}
				w.WriteHeader(http.StatusOK)
				w.Write(vsReturnedNoCorrectionJSON)
				return
			}
			// Current implementation compares pointers, so it may still touch the property endpoint
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/versionedSettings/config/parameters/showSettingsChanges" && r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("true"))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		actual, err := httpClient.SetVersionedSettings(objId, vsRequested)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, _ := json.Marshal(actual)
		if string(got) != string(vsReturnedNoCorrectionJSON) {
			t.Fatalf("unexpected result: %s", string(got))
		}
	})

	t.Run("versioned settings: set triggers correction path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First, PUT config returns payload with ShowSettingsChanges=false
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/versionedSettings/config" && r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				w.Write(vsReturnedNeedsCorrectionJSON)
				return
			}
			// Then, property correction is called
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/versionedSettings/config/parameters/showSettingsChanges" && r.Method == http.MethodPut {
				defer r.Body.Close()
				b, _ := io.ReadAll(r.Body)
				if string(b) != "true" {
					t.Fatalf("unexpected property body: %s", string(b))
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("true"))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		actual, err := httpClient.SetVersionedSettings(objId, vsRequested)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if actual.ShowSettingsChanges == nil || *actual.ShowSettingsChanges != true {
			t.Fatalf("expected ShowSettingsChanges corrected to true, got: %#v", actual.ShowSettingsChanges)
		}
	})

	t.Run("versioned settings: set correction malformed property response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/versionedSettings/config" && r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				w.Write(vsReturnedNeedsCorrectionJSON)
				return
			}
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/versionedSettings/config/parameters/showSettingsChanges" && r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not_boolean"))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		_, err := httpClient.SetVersionedSettings(objId, vsRequested)
		if err == nil || err.Error() == "" {
			t.Fatalf("expected error due to malformed property response, got: %v", err)
		}
	})

	t.Run("versioned settings property: success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == objTcEndpoint+"/id:"+objId+"/versionedSettings/config/parameters/showSettingsChanges" && r.Method == http.MethodPut {
				defer r.Body.Close()
				b, _ := io.ReadAll(r.Body)
				if string(b) != "true" {
					t.Fatalf("unexpected body: %s", string(b))
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("true"))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		httpClient := NewClient(server.URL, "token", "", "", 12)
		b, err := httpClient.SetVersionedSettingsProperty(objId, "showSettingsChanges", "true")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(b) != "true" {
			t.Fatalf("unexpected value: %s", string(b))
		}
	})

	t.Run("versioned settings property: server 500 error without retry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("boom"))
		}))
		defer server.Close()

		// Use 0 retries to avoid sleep
		httpClient := NewClient(server.URL, "token", "", "", 0)
		_, err := httpClient.SetVersionedSettingsProperty(objId, "showSettingsChanges", "true")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("retryPolicy behavior", func(t *testing.T) {
		resp500 := &http.Response{StatusCode: http.StatusInternalServerError}
		should, err := retryPolicy(nil, resp500, nil)
		if err != nil || !should {
			t.Fatalf("expected retry for 500, got should=%v, err=%v", should, err)
		}

		resp200 := &http.Response{StatusCode: http.StatusOK}
		should, err = retryPolicy(nil, resp200, nil)
		if err != nil || should {
			t.Fatalf("expected no retry for 200, got should=%v, err=%v", should, err)
		}

		resp404 := &http.Response{StatusCode: http.StatusNotFound}
		should, err = retryPolicy(nil, resp404, nil)
		if err != nil || should {
			t.Fatalf("expected no retry for 404, got should=%v, err=%v", should, err)
		}
	})
}
