package teamcity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	frameworkschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	emptyPropertiesConfig := providerConfig + `
resource "teamcity_cloud_profile" "test" {
  name              = "Renamed AWS EC2 Profile"
  cloud_provider_id = "amazon"
  project_id        = "CloudProject"
  properties        = {}

  image {
    name       = "Ubuntu agent"
    properties = {}
  }
}
`
	noImagesConfig := providerConfig + `
resource "teamcity_cloud_profile" "test" {
  name              = "Renamed AWS EC2 Profile"
  cloud_provider_id = "amazon"
  project_id        = "CloudProject"
  properties        = {}
}
`

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
				Config: emptyPropertiesConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "properties.%", "0"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.0.properties.%", "0"),
				),
			},
			{
				Config:           emptyPropertiesConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
			},
			{
				Config: noImagesConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.#", "0"),
				),
			},
			{
				Config:           noImagesConfig,
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

func TestCloudProfileSchemaPropertiesAreOptionalAndComputed(t *testing.T) {
	resourceUnderTest := &cloudProfileResource{}
	response := frameworkresource.SchemaResponse{}
	resourceUnderTest.Schema(context.Background(), frameworkresource.SchemaRequest{}, &response)

	profileProperties, ok := response.Schema.Attributes["properties"].(frameworkschema.MapAttribute)
	if !ok {
		t.Fatalf("profile properties attribute = %T, want schema.MapAttribute", response.Schema.Attributes["properties"])
	}
	if !profileProperties.Optional || !profileProperties.Computed {
		t.Fatalf("profile properties Optional/Computed = %t/%t, want true/true", profileProperties.Optional, profileProperties.Computed)
	}

	imageBlock, ok := response.Schema.Blocks["image"].(frameworkschema.ListNestedBlock)
	if !ok {
		t.Fatalf("image block = %T, want schema.ListNestedBlock", response.Schema.Blocks["image"])
	}
	imageProperties, ok := imageBlock.NestedObject.Attributes["properties"].(frameworkschema.MapAttribute)
	if !ok {
		t.Fatalf("image properties attribute = %T, want schema.MapAttribute", imageBlock.NestedObject.Attributes["properties"])
	}
	if !imageProperties.Optional || !imageProperties.Computed {
		t.Fatalf("image properties Optional/Computed = %t/%t, want true/true", imageProperties.Optional, imageProperties.Computed)
	}
}

func TestCloudProfileModelToJSONMatchesImageIDsByName(t *testing.T) {
	resourceUnderTest := &cloudProfileResource{}
	previous := models.CloudProfileDataModel{Images: []models.CloudImageDataModel{
		{Id: types.StringValue("PROJECT_EXT_A"), Name: types.StringValue("first")},
		{Id: types.StringValue("PROJECT_EXT_B"), Name: types.StringValue("second")},
	}}
	plan := models.CloudProfileDataModel{Images: []models.CloudImageDataModel{
		{Name: types.StringValue("second")},
		{Name: types.StringValue("first")},
	}}
	diagnostics := diag.Diagnostics{}

	profile := resourceUnderTest.modelToJSON(context.Background(), plan, &previous, &diagnostics)
	if diagnostics.HasError() {
		t.Fatalf("modelToJSON diagnostics: %#v", diagnostics)
	}
	if profile.Images == nil || len(profile.Images.CloudImage) != 2 {
		t.Fatalf("images = %#v, want two images", profile.Images)
	}
	if profile.Images.CloudImage[0].Id != "PROJECT_EXT_B" || profile.Images.CloudImage[1].Id != "PROJECT_EXT_A" {
		t.Fatalf("planned IDs = [%q, %q], want [PROJECT_EXT_B, PROJECT_EXT_A]", profile.Images.CloudImage[0].Id, profile.Images.CloudImage[1].Id)
	}
}

func TestCloudProfileJSONToModelPreservesConfiguredImageOrder(t *testing.T) {
	resourceUnderTest := &cloudProfileResource{}
	state := models.CloudProfileDataModel{Images: []models.CloudImageDataModel{
		{Id: types.StringValue("PROJECT_EXT_B"), Name: types.StringValue("second")},
		{Id: types.StringValue("PROJECT_EXT_A"), Name: types.StringValue("first")},
	}}
	profile := &models.CloudProfileJson{Images: &models.CloudImagesJson{CloudImage: []models.CloudImageJson{
		{Id: "PROJECT_EXT_A", Name: "first"},
		{Id: "PROJECT_EXT_B", Name: "second"},
	}}}
	diagnostics := diag.Diagnostics{}

	resourceUnderTest.jsonToModel(context.Background(), profile, &state, &diagnostics)
	if diagnostics.HasError() {
		t.Fatalf("jsonToModel diagnostics: %#v", diagnostics)
	}
	if len(state.Images) != 2 {
		t.Fatalf("image count = %d, want 2", len(state.Images))
	}
	if state.Images[0].Name.ValueString() != "second" || state.Images[0].Id.ValueString() != "PROJECT_EXT_B" ||
		state.Images[1].Name.ValueString() != "first" || state.Images[1].Id.ValueString() != "PROJECT_EXT_A" {
		t.Fatalf("state images = %#v, want configured order second/first", state.Images)
	}
}

func TestCloudProfileModelToJSONRejectsDuplicateImageNames(t *testing.T) {
	resourceUnderTest := &cloudProfileResource{}
	plan := models.CloudProfileDataModel{Images: []models.CloudImageDataModel{
		{Name: types.StringValue("duplicate")},
		{Name: types.StringValue("duplicate")},
	}}
	diagnostics := diag.Diagnostics{}

	resourceUnderTest.modelToJSON(context.Background(), plan, nil, &diagnostics)
	if !diagnostics.HasError() {
		t.Fatal("modelToJSON accepted duplicate cloud image names")
	}
}
