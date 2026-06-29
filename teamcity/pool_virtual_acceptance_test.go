package teamcity

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"terraform-provider-teamcity/models"
)

const testPoolFieldsQuery = "fields=id,name,maxAgents,projects(project(id,name,virtual))"

type fakePoolServerState struct {
	mu              sync.Mutex
	poolID          int64
	poolName        string
	managedProjects []string
	virtualProjects []string
}

func newFakePoolServer(t *testing.T, poolName string, managedProjects, virtualProjects []string) *httptest.Server {
	t.Helper()

	state := &fakePoolServerState{
		poolID:          1,
		poolName:        poolName,
		managedProjects: append([]string{}, managedProjects...),
		virtualProjects: append([]string{}, virtualProjects...),
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/app/rest":
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
			return
		case r.Method == http.MethodPost && r.URL.Path == "/app/rest/agentPools":
			var req models.PoolJson
			if err := decodeJSONBody(r, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			state.mu.Lock()
			state.poolName = req.Name
			state.mu.Unlock()

			writeJSON(w, models.PoolJson{Name: req.Name, Id: &state.poolID, Size: req.Size})
			return
		case r.Method == http.MethodPut && r.URL.Path == fmt.Sprintf("/app/rest/agentPools/name:%s/projects", state.poolName):
			var req models.ProjectsJson
			if err := decodeJSONBody(r, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			managed := make([]string, 0, len(req.Project))
			for _, project := range req.Project {
				if project.Id != nil {
					managed = append(managed, *project.Id)
				}
			}

			state.mu.Lock()
			state.managedProjects = managed
			state.mu.Unlock()

			writeJSON(w, state.projectsJSON(false))
			return
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/app/rest/agentPools/name:%s", state.poolName):
			if r.URL.RawQuery != testPoolFieldsQuery {
				http.Error(w, fmt.Sprintf("unexpected pool query: %s", r.URL.RawQuery), http.StatusBadRequest)
				return
			}
			writeJSON(w, models.PoolJson{Name: state.poolName, Id: &state.poolID, Projects: state.projectsJSON(true)})
			return
		case r.Method == http.MethodDelete && r.URL.Path == fmt.Sprintf("/app/rest/agentPools/id:%d", state.poolID):
			w.WriteHeader(http.StatusOK)
			return
		default:
			http.Error(w, fmt.Sprintf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery), http.StatusNotFound)
			return
		}
	}))
}

func (s *fakePoolServerState) projectsJSON(includeVirtual bool) *models.ProjectsJson {
	s.mu.Lock()
	defer s.mu.Unlock()

	projects := make([]models.ProjectJson, 0, len(s.managedProjects)+len(s.virtualProjects))
	for _, id := range s.managedProjects {
		projectID := id
		projects = append(projects, models.ProjectJson{Name: id, Id: &projectID})
	}
	if includeVirtual {
		for _, id := range s.virtualProjects {
			projectID := id
			projects = append(projects, models.ProjectJson{Name: id, Id: &projectID, Virtual: true})
		}
	}

	return &models.ProjectsJson{Project: projects}
}

func decodeJSONBody(r *http.Request, target any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, target)
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func poolProviderConfig() string {
	return providerConfig
}

func testCheckProjects(resourceName string, want []string, unwanted []string) resource.TestCheckFunc {
	wantSet := make(map[string]struct{}, len(want))
	for _, item := range want {
		wantSet[item] = struct{}{}
	}
	unwantedSet := make(map[string]struct{}, len(unwanted))
	for _, item := range unwanted {
		unwantedSet[item] = struct{}{}
	}

	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		got := map[string]struct{}{}
		for key, value := range rs.Primary.Attributes {
			if strings.HasPrefix(key, "projects.") && key != "projects.#" {
				got[value] = struct{}{}
			}
		}

		if len(got) != len(wantSet) {
			return fmt.Errorf("unexpected projects length: got %d (%v), want %d (%v)", len(got), got, len(wantSet), want)
		}
		for item := range wantSet {
			if _, ok := got[item]; !ok {
				return fmt.Errorf("expected project %q missing from state: %v", item, got)
			}
		}
		for item := range unwantedSet {
			if _, ok := got[item]; ok {
				return fmt.Errorf("unexpected virtual project %q found in state: %v", item, got)
			}
		}

		return nil
	}
}

func TestAccPoolResource_virtualProjectsNoDrift(t *testing.T) {
	virtualProjects := []string{
		"Gradle_Master_Check_Quick_2_bucket1_virtual",
		"Gradle_Release7x_Check_Stage_PullRequestFeedback_X",
	}
	server := newFakePoolServer(t, "demo-pool", nil, virtualProjects)
	defer server.Close()
	t.Setenv("TEAMCITY_HOST", server.URL)
	t.Setenv("TEAMCITY_TOKEN", "test-token")

	cfg := poolProviderConfig() + `
resource "teamcity_pool" "test" {
  name     = "demo-pool"
  projects = ["Managed_A", "Managed_B"]
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_pool.test", "name", "demo-pool"),
					resource.TestCheckResourceAttr("teamcity_pool.test", "projects.#", "2"),
					testCheckProjects("teamcity_pool.test", []string{"Managed_A", "Managed_B"}, virtualProjects),
				),
			},
			{
				Config: cfg,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccPoolDataSource_virtualProjectsFiltered(t *testing.T) {
	virtualProjects := []string{
		"Gradle_Master_Check_Quick_2_bucket1_virtual",
		"Gradle_Release7x_Check_Stage_PullRequestFeedback_X",
	}
	server := newFakePoolServer(t, "demo-pool", []string{"Managed_A", "Managed_B"}, virtualProjects)
	defer server.Close()
	t.Setenv("TEAMCITY_HOST", server.URL)
	t.Setenv("TEAMCITY_TOKEN", "test-token")

	cfg := poolProviderConfig() + `
data "teamcity_pool" "test" {
  name = "demo-pool"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.teamcity_pool.test", "name", "demo-pool"),
					resource.TestCheckResourceAttr("data.teamcity_pool.test", "projects.#", "2"),
					testCheckProjects("data.teamcity_pool.test", []string{"Managed_A", "Managed_B"}, virtualProjects),
				),
			},
		},
	})
}
