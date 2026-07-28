package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"terraform-provider-teamcity/models"
)

func TestCreateCloudProfilePostsProfileAndImages(t *testing.T) {
	expected := models.CloudProfileJson{
		Name:            "AWS EC2 Profile",
		CloudProviderId: "amazon",
		Project:         &models.CloudProfileProjectJson{Id: stringPointer("CloudProject")},
		Properties: &models.Properties{Property: []models.Property{
			{Name: "access-id", Value: "access-key"},
		}},
		Images: &models.CloudImagesJson{CloudImage: []models.CloudImageJson{
			{
				Name: "Ubuntu agent",
				Properties: &models.Properties{Property: []models.Property{
					{Name: "image-id", Value: "ami-0123456789abcdef0"},
				}},
			},
		}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/app/rest/cloud/profiles" {
			t.Fatalf("path = %s, want /app/rest/cloud/profiles", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		var projectPayload map[string]json.RawMessage
		if err := json.Unmarshal(payload["project"], &projectPayload); err != nil {
			t.Fatal(err)
		}
		if _, hasName := projectPayload["name"]; hasName {
			t.Fatalf("project payload = %s, must not send an empty project name", payload["project"])
		}

		var actual models.CloudProfileJson
		if err := json.Unmarshal(body, &actual); err != nil {
			t.Fatal(err)
		}
		if actual.Name != expected.Name || actual.CloudProviderId != expected.CloudProviderId {
			t.Fatalf("profile = %#v, want name=%q cloudProviderId=%q", actual, expected.Name, expected.CloudProviderId)
		}
		if actual.Project == nil || actual.Project.Id == nil || *actual.Project.Id != "CloudProject" {
			t.Fatalf("project = %#v, want id CloudProject", actual.Project)
		}
		if actual.Properties == nil || len(actual.Properties.Property) != 1 || actual.Properties.Property[0] != (models.Property{Name: "access-id", Value: "access-key"}) {
			t.Fatalf("profile properties = %#v, want access-id", actual.Properties)
		}
		if actual.Images == nil || len(actual.Images.CloudImage) != 1 || actual.Images.CloudImage[0].Name != "Ubuntu agent" {
			t.Fatalf("images = %#v, want one Ubuntu agent image", actual.Images)
		}
		if actual.Images.CloudImage[0].Properties == nil || len(actual.Images.CloudImage[0].Properties.Property) != 1 || actual.Images.CloudImage[0].Properties.Property[0] != (models.Property{Name: "image-id", Value: "ami-0123456789abcdef0"}) {
			t.Fatalf("image properties = %#v, want image-id", actual.Images.CloudImage[0].Properties)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"aws-profile","name":"AWS EC2 Profile","cloudProviderId":"amazon"}`))
	}))
	defer server.Close()

	cloudClient := NewClient(server.URL, "token", "", "", 0)
	created, err := cloudClient.CreateCloudProfile(expected)
	if err != nil {
		t.Fatal(err)
	}
	if created.Id != "aws-profile" {
		t.Fatalf("created ID = %q, want aws-profile", created.Id)
	}
}

func TestGetCloudProfileRequestsAllStateFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/app/rest/cloud/profiles/id:aws-profile" {
			t.Fatalf("path = %s, want profile locator", r.URL.Path)
		}
		if r.URL.RawQuery != cloudProfileFieldsQuery {
			t.Fatalf("query = %q, want %q", r.URL.RawQuery, cloudProfileFieldsQuery)
		}
		_, _ = w.Write([]byte(`{
			"id":"aws-profile",
			"name":"AWS EC2 Profile",
			"cloudProviderId":"amazon",
			"project":{"id":"CloudProject"},
			"images":{"image":[{"id":"aws-image","name":"Ubuntu agent","agentPoolId":1}]}
		}`))
	}))
	defer server.Close()

	cloudClient := NewClient(server.URL, "token", "", "", 0)
	profile, err := cloudClient.GetCloudProfile("aws-profile")
	if err != nil {
		t.Fatal(err)
	}
	if profile == nil || profile.Id != "aws-profile" {
		t.Fatalf("profile = %#v, want aws-profile", profile)
	}
	if profile.Images == nil || len(profile.Images.CloudImage) != 1 || profile.Images.CloudImage[0].AgentPoolId == nil || *profile.Images.CloudImage[0].AgentPoolId != 1 {
		t.Fatalf("images = %#v, want image with agentPoolId 1", profile.Images)
	}
}

func TestGetCloudProfileReturnsNilForMissingProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cloudClient := NewClient(server.URL, "token", "", "", 0)
	profile, err := cloudClient.GetCloudProfile("missing-profile")
	if err != nil {
		t.Fatal(err)
	}
	if profile != nil {
		t.Fatalf("profile = %#v, want nil", profile)
	}
}

func TestUpdateAndDeleteCloudProfileUseProfileLocator(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, fmt.Sprintf("%s %s", r.Method, r.URL.Path))
		switch r.Method {
		case http.MethodPut:
			_, _ = w.Write([]byte(`{"id":"aws-profile","name":"Renamed AWS EC2 Profile","cloudProviderId":"amazon"}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	cloudClient := NewClient(server.URL, "token", "", "", 0)
	updated, err := cloudClient.UpdateCloudProfile("aws-profile", models.CloudProfileJson{
		Name:            "Renamed AWS EC2 Profile",
		CloudProviderId: "amazon",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Id != "aws-profile" {
		t.Fatalf("updated ID = %q, want aws-profile", updated.Id)
	}
	if err := cloudClient.DeleteCloudProfile("aws-profile"); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"PUT /app/rest/cloud/profiles/id:aws-profile",
		"DELETE /app/rest/cloud/profiles/id:aws-profile",
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request %d = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestDeleteCloudProfileIsIdempotentWhenProfileIsMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cloudClient := NewClient(server.URL, "token", "", "", 0)
	if err := cloudClient.DeleteCloudProfile("missing-profile"); err != nil {
		t.Fatalf("DeleteCloudProfile returned %v, want nil for a missing profile", err)
	}
}

func stringPointer(value string) *string {
	return &value
}
