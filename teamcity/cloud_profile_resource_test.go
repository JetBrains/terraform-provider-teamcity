package teamcity

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"terraform-provider-teamcity/models"
)

const testCloudProfileFieldsQuery = "fields=id,name,cloudProviderId,project(id),properties(property(name,value)),images(image(id,name,agentPoolId,properties(property(name,value))))"

type fakeCloudProfileServerState struct {
	mu      sync.Mutex
	profile *models.CloudProfileJson
}

func newFakeCloudProfileServer(t *testing.T) *httptest.Server {
	t.Helper()

	state := &fakeCloudProfileServerState{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const profilePath = "/app/rest/cloud/profiles/id:aws-profile"

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/app/rest":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("OK"))
			return
		case r.Method == http.MethodPost && r.URL.Path == "/app/rest/cloud/profiles":
			var request models.CloudProfileJson
			if err := decodeJSONBody(r, &request); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			state.mu.Lock()
			state.profile = completeCloudProfileIDs(request)
			response := *state.profile
			state.mu.Unlock()
			writeJSON(w, response)
			return
		case r.Method == http.MethodGet && r.URL.Path == profilePath:
			if r.URL.RawQuery != testCloudProfileFieldsQuery {
				http.Error(w, fmt.Sprintf("unexpected cloud profile query: %s", r.URL.RawQuery), http.StatusBadRequest)
				return
			}
			state.mu.Lock()
			profile := state.profile
			state.mu.Unlock()
			if profile == nil {
				http.Error(w, "cloud profile not found", http.StatusNotFound)
				return
			}
			writeJSON(w, *profile)
			return
		case r.Method == http.MethodPut && r.URL.Path == profilePath:
			var request models.CloudProfileJson
			if err := decodeJSONBody(r, &request); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if request.Images != nil && len(request.Images.CloudImage) > 0 && request.Images.CloudImage[0].Id != "aws-image-1" {
				http.Error(w, "cloud image update must preserve the existing image ID", http.StatusBadRequest)
				return
			}
			state.mu.Lock()
			if state.profile == nil {
				state.mu.Unlock()
				http.Error(w, "cloud profile not found", http.StatusNotFound)
				return
			}
			request.Id = state.profile.Id
			state.profile = completeCloudProfileIDs(request)
			response := *state.profile
			state.mu.Unlock()
			writeJSON(w, response)
			return
		case r.Method == http.MethodDelete && r.URL.Path == profilePath:
			state.mu.Lock()
			state.profile = nil
			state.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, fmt.Sprintf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery), http.StatusNotFound)
		}
	}))
}

func completeCloudProfileIDs(profile models.CloudProfileJson) *models.CloudProfileJson {
	profile.Id = "aws-profile"
	if profile.Images != nil {
		for index := range profile.Images.CloudImage {
			profile.Images.CloudImage[index].Id = fmt.Sprintf("aws-image-%d", index+1)
		}
	}
	return &profile
}

func cloudProfileConfig(name, accessID, imageID string) string {
	return providerConfig + fmt.Sprintf(`
resource "teamcity_cloud_profile" "test" {
  name              = %q
  cloud_provider_id = "amazon"
  project_id        = "CloudProject"

  properties = {
    "access-id" = %q
    "region"    = "eu-west-1"
  }

  image {
    name = "Ubuntu agent"
    properties = {
      "image-id" = %q
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

	initialConfig := cloudProfileConfig("AWS EC2 Profile", "initial-access-key", "ami-0123456789abcdef0")
	updatedConfig := cloudProfileConfig("Renamed AWS EC2 Profile", "updated-access-key", "ami-0fedcba9876543210")
	emptyConfig := providerConfig + `
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
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "id", "aws-profile"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "name", "AWS EC2 Profile"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "cloud_provider_id", "amazon"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "project_id", "CloudProject"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "properties.access-id", "initial-access-key"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.#", "1"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.0.id", "aws-image-1"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.0.agent_pool_id", "1"),
				),
			},
			{
				Config: initialConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "name", "Renamed AWS EC2 Profile"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "properties.access-id", "updated-access-key"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.0.properties.image-id", "ami-0fedcba9876543210"),
				),
			},
			{
				Config: emptyConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "properties.%", "0"),
					resource.TestCheckResourceAttr("teamcity_cloud_profile.test", "image.#", "0"),
				),
			},
			{
				Config: emptyConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ResourceName:      "teamcity_cloud_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
