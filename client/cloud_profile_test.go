package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"terraform-provider-teamcity/models"
)

func TestCreateCloudProfileUsesProjectFeaturesAndReadsGeneratedIDs(t *testing.T) {
	var features []models.ProjectFeatureJson
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const featuresPath = "/app/rest/projects/id:CloudProject/projectFeatures"
		if r.Method == http.MethodGet && r.URL.Path == featuresPath+"/id:amazon-42" {
			if len(features) == 0 {
				http.Error(w, "profile not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(features[0]); err != nil {
				t.Fatal(err)
			}
			return
		}
		if r.URL.Path != featuresPath {
			t.Fatalf("path = %q, want %q", r.URL.Path, featuresPath)
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(models.ProjectFeaturesJson{ProjectFeature: features}); err != nil {
				t.Fatal(err)
			}
		case http.MethodPost:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			var feature models.ProjectFeatureJson
			if err := json.Unmarshal(body, &feature); err != nil {
				t.Fatal(err)
			}
			switch feature.Type {
			case "CloudProfile":
				if featurePropertyValue(feature.Properties, "cloud-code") != "amazon" || featurePropertyValue(feature.Properties, "name") != "AWS EC2 Profile" {
					t.Fatalf("profile properties = %#v, want cloud-code and name", feature.Properties)
				}
				feature.Id = stringPointer("amazon-42")
			case "CloudImage":
				if featurePropertyValue(feature.Properties, "profileId") != "amazon-42" {
					t.Fatalf("image profileId = %q, want generated profile id", featurePropertyValue(feature.Properties, "profileId"))
				}
				if featurePropertyValue(feature.Properties, "image-name-prefix") != "Ubuntu agent" {
					t.Fatalf("image properties = %#v, want image-name-prefix", feature.Properties)
				}
				feature.Id = stringPointer("PROJECT_EXT_42")
			default:
				t.Fatalf("unexpected project feature type %q", feature.Type)
			}
			features = append(features, feature)
			w.WriteHeader(http.StatusOK) // TeamCity feature writes can return an empty body.
		default:
			t.Fatalf("method = %s, want GET or POST", r.Method)
		}
	}))
	defer server.Close()

	cloudClient := NewClient(server.URL, "token", "", "", 0)
	created, err := cloudClient.CreateCloudProfile("CloudProject", models.CloudProfileJson{
		Name:            "AWS EC2 Profile",
		CloudProviderId: "amazon",
		Properties: &models.Properties{Property: []models.Property{
			{Name: "region", Value: "eu-west-1"},
			{Name: "secure:access-id", Value: "access-key"},
		}},
		Images: &models.CloudImagesJson{CloudImage: []models.CloudImageJson{{
			Name: "Ubuntu agent",
			Properties: &models.Properties{Property: []models.Property{
				{Name: "amazon-id", Value: "ami-0123456789abcdef0"},
			}},
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Id != "amazon-42" {
		t.Fatalf("profile ID = %q, want generated profile feature ID", created.Id)
	}
	if created.Images == nil || len(created.Images.CloudImage) != 1 || created.Images.CloudImage[0].Id != "PROJECT_EXT_42" {
		t.Fatalf("images = %#v, want generated image feature ID", created.Images)
	}
}

func featurePropertyValue(properties models.Properties, name string) string {
	for _, property := range properties.Property {
		if property.Name == name {
			return property.Value
		}
	}
	return ""
}

func stringPointer(value string) *string {
	return &value
}
