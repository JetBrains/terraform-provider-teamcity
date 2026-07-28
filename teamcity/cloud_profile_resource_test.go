package teamcity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"terraform-provider-teamcity/models"
)

type fakeProjectFeatureServerState struct {
	mu       sync.Mutex
	features []models.ProjectFeatureJson
	nextID   int
}

func newFakeCloudProfileServer(t *testing.T) *httptest.Server {
	t.Helper()
	state := &fakeProjectFeatureServerState{nextID: 1}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const collection = "/app/rest/projects/id:CloudProject/projectFeatures"
		if r.Method == http.MethodGet && r.URL.Path == "/app/rest" {
			_, _ = w.Write([]byte("OK"))
			return
		}

		state.mu.Lock()
		defer state.mu.Unlock()
		if r.URL.Path == collection {
			switch r.Method {
			case http.MethodGet:
				writeCloudProfileJSON(w, models.ProjectFeaturesJson{ProjectFeature: redactedFeatures(state.features)})
			case http.MethodPost:
				var feature models.ProjectFeatureJson
				if err := json.NewDecoder(r.Body).Decode(&feature); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if feature.Type == "CloudProfile" {
					feature.Id = stringPointer("amazon-42")
				} else if feature.Type == "CloudImage" {
					feature.Id = stringPointer(fmt.Sprintf("PROJECT_EXT_%d", state.nextID))
					state.nextID++
				} else {
					http.Error(w, "unexpected project feature type", http.StatusBadRequest)
					return
				}
				state.features = append(state.features, feature)
				w.WriteHeader(http.StatusOK) // live TeamCity returns no JSON body for this route
			default:
				http.Error(w, "unsupported collection method", http.StatusMethodNotAllowed)
			}
			return
		}

		const itemPrefix = collection + "/id:"
		if !strings.HasPrefix(r.URL.Path, itemPrefix) {
			http.Error(w, "unexpected route", http.StatusNotFound)
			return
		}
		featureID := strings.TrimPrefix(r.URL.Path, itemPrefix)
		index := featureIndex(state.features, featureID)
		if index == -1 {
			http.Error(w, "feature not found", http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			writeCloudProfileJSON(w, redactFeature(state.features[index]))
		case http.MethodPut:
			var feature models.ProjectFeatureJson
			if err := json.NewDecoder(r.Body).Decode(&feature); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			feature.Id = state.features[index].Id
			state.features[index] = feature
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			state.features = append(state.features[:index], state.features[index+1:]...)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unsupported item method", http.StatusMethodNotAllowed)
		}
	}))
}

func stringPointer(value string) *string {
	return &value
}

func featureIndex(features []models.ProjectFeatureJson, id string) int {
	for index, feature := range features {
		if feature.Id != nil && *feature.Id == id {
			return index
		}
	}
	return -1
}

func redactedFeatures(features []models.ProjectFeatureJson) []models.ProjectFeatureJson {
	result := make([]models.ProjectFeatureJson, len(features))
	for index, feature := range features {
		result[index] = redactFeature(feature)
	}
	return result
}

func redactFeature(feature models.ProjectFeatureJson) models.ProjectFeatureJson {
	properties := append([]models.Property(nil), feature.Properties.Property...)
	for index := range properties {
		if strings.HasPrefix(properties[index].Name, "secure:") {
			properties[index].Value = ""
		}
	}
	feature.Properties = models.Properties{Property: properties}
	return feature
}

func writeCloudProfileJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		panic(err)
	}
}

func cloudProfileConfig(name, accessID, imageID string) string {
	return providerConfig + fmt.Sprintf(`
resource "teamcity_cloud_profile" "test" {
  name              = %q
  cloud_provider_id = "amazon"
  project_id        = "CloudProject"

  properties = {
    "secure:access-id" = %q
    "region"           = "eu-west-1"
  }

  image {
    name = "Ubuntu agent"
    properties = {
      "amazon-id" = %q
      "source-id" = "tcci-5370"
    }
    agent_pool_id = 1
  }
}
`, name, accessID, imageID)
}

func TestAccCloudProfileResourceLifecycle(t *testing.T) {
	server := newFakeCloudProfileServer(t)
	defer server.Close()
	t.Setenv("TEAMCITY_HOST", server.URL)
	t.Setenv("TEAMCITY_TOKEN", "test-token")
	t.Setenv("TF_CLI_CONFIG_FILE", "")

	initialConfig := cloudProfileConfig("AWS EC2 Profile", "initial-access-key", "ami-0123456789abcdef0")
	updatedConfig := cloudProfileConfig("Renamed AWS EC2 Profile", "updated-access-key", "ami-0fedcba9876543210")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "id", "amazon-42"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "properties.secure:access-id", "initial-access-key"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.0.id", "PROJECT_EXT_1"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.0.agent_pool_id", "1"),
				),
			},
			{
				Config:           initialConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "name", "Renamed AWS EC2 Profile"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "properties.secure:access-id", "updated-access-key"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.0.properties.amazon-id", "ami-0fedcba9876543210"),
				),
			},
			{
				Config:           updatedConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
			},
			{
				ResourceName:            "teamcity_cloud_profile.test",
				ImportStateId:           "CloudProject/amazon-42",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"properties.secure:access-id"},
			},
		},
	})
}
