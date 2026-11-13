package teamcity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"terraform-provider-teamcity/client"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fakeUserRoles stands in for TeamCity's endpoints for one user, "racer" with ID 7, keeping the
// user's roles as TeamCity does. It holds each read of the user until a second read arrives, or for
// a second at most, so that two resources working on the user in parallel both read it before
// either writes, as they can against a real server.
type fakeUserRoles struct {
	t *testing.T

	mu       sync.Mutex
	roles    map[string]bool // "ROLE_ID@scope"
	reads    int
	bothRead chan struct{}
}

func newFakeUserRoles(t *testing.T, roles ...string) (*fakeUserRoles, *httptest.Server) {
	f := &fakeUserRoles{t: t, roles: map[string]bool{}, bothRead: make(chan struct{})}
	for _, role := range roles {
		f.roles[role] = true
	}
	server := httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(server.Close)
	return f, server
}

func (f *fakeUserRoles) handle(w http.ResponseWriter, r *http.Request) {
	const rolesPath = "/app/rest/users/id:7/roles"
	switch {
	case r.Method == http.MethodGet && (r.URL.Path == "/app/rest/users/username:racer" || r.URL.Path == "/app/rest/users/id:7"):
		f.waitForSecondRead()
		f.writeUser(w)
	case r.Method == http.MethodPut && r.URL.Path == "/app/rest/users/id:7":
		var user client.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			f.t.Errorf("invalid user: %v", err)
		}
		f.mu.Lock()
		f.roles = map[string]bool{}
		if user.Roles != nil {
			for _, role := range user.Roles.RoleAssignment {
				f.roles[role.Id+"@"+role.Scope] = true
			}
		}
		f.mu.Unlock()
		f.writeUser(w)
	case r.Method == http.MethodPost && r.URL.Path == rolesPath:
		var role client.RoleAssignment
		if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
			f.t.Errorf("invalid role: %v", err)
		}
		f.mu.Lock()
		f.roles[role.Id+"@"+role.Scope] = true
		f.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, rolesPath+"/"):
		roleId, scope, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, rolesPath+"/"), "/")
		f.mu.Lock()
		delete(f.roles, roleId+"@"+scope)
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		f.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (f *fakeUserRoles) waitForSecondRead() {
	f.mu.Lock()
	f.reads++
	if f.reads == 2 {
		close(f.bothRead)
	}
	f.mu.Unlock()
	select {
	case <-f.bothRead:
	case <-time.After(time.Second):
	}
}

func (f *fakeUserRoles) writeUser(w http.ResponseWriter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := int64(7)
	user := client.User{Id: &id, Username: "racer", Roles: &client.RoleAssignments{RoleAssignment: []client.RoleAssignment{}}}
	for role := range f.roles {
		roleId, scope, _ := strings.Cut(role, "@")
		user.Roles.RoleAssignment = append(user.Roles.RoleAssignment, client.RoleAssignment{Id: roleId, Scope: scope})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func (f *fakeUserRoles) roleList() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var roles []string
	for role := range f.roles {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return roles
}

func newTestUserRoleAssignmentResource(serverURL string) (*userRoleAssignmentResource, resource.SchemaResponse) {
	c := client.NewClient(serverURL, "token", "", "", 0)
	r := &userRoleAssignmentResource{client: &c}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	return r, schemaResp
}

// Two roles assigned to one user in parallel, as Terraform does by default, must both end up on the
// user.
func TestUserRoleAssignmentCreate_ParallelAssignmentsToOneUser(t *testing.T) {
	fake, server := newFakeUserRoles(t)
	r, schemaResp := newTestUserRoleAssignmentResource(server.URL)

	var wg sync.WaitGroup
	for _, roleId := range []string{"AGENT_MANAGER", "PROJECT_VIEWER"} {
		plan := tfsdk.Plan{Schema: schemaResp.Schema}
		if diags := plan.Set(context.Background(), userRoleAssignmentResourceModel{
			Id:       types.StringUnknown(),
			UserId:   types.StringUnknown(),
			Username: types.StringValue("racer"),
			RoleId:   types.StringValue(roleId),
			Scope:    types.StringValue("p:_Root"),
		}); diags.HasError() {
			t.Fatalf("failed to set plan: %v", diags)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("unexpected diagnostics: %v", resp.Diagnostics)
			}
		}()
	}
	wg.Wait()

	want := []string{"AGENT_MANAGER@p:_Root", "PROJECT_VIEWER@p:_Root"}
	if got := fake.roleList(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("user has roles %v, want %v", got, want)
	}
}

// Two of a user's roles removed in parallel must both go, and the user's other roles must stay.
func TestUserRoleAssignmentDelete_ParallelRemovalsFromOneUser(t *testing.T) {
	fake, server := newFakeUserRoles(t, "AGENT_MANAGER@p:_Root", "PROJECT_VIEWER@p:_Root", "SYSTEM_ADMIN@g")
	r, schemaResp := newTestUserRoleAssignmentResource(server.URL)

	var wg sync.WaitGroup
	for _, roleId := range []string{"AGENT_MANAGER", "PROJECT_VIEWER"} {
		state := tfsdk.State{Schema: schemaResp.Schema}
		if diags := state.Set(context.Background(), userRoleAssignmentResourceModel{
			Id:       types.StringValue("7_" + roleId + "_p:_Root"),
			UserId:   types.StringValue("7"),
			Username: types.StringValue("racer"),
			RoleId:   types.StringValue(roleId),
			Scope:    types.StringValue("p:_Root"),
		}); diags.HasError() {
			t.Fatalf("failed to set state: %v", diags)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := resource.DeleteResponse{State: state}
			r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("unexpected diagnostics: %v", resp.Diagnostics)
			}
		}()
	}
	wg.Wait()

	want := []string{"SYSTEM_ADMIN@g"}
	if got := fake.roleList(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("user has roles %v, want %v", got, want)
	}
}

// A user deleted outside Terraform takes its roles with it, so removing one of them succeeds.
func TestUserRoleAssignmentDelete_UserNoLongerExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	r, schemaResp := newTestUserRoleAssignmentResource(server.URL)

	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(context.Background(), userRoleAssignmentResourceModel{
		Id:       types.StringValue("7_AGENT_MANAGER_p:_Root"),
		UserId:   types.StringValue("7"),
		Username: types.StringValue("racer"),
		RoleId:   types.StringValue("AGENT_MANAGER"),
		Scope:    types.StringValue("p:_Root"),
	}); diags.HasError() {
		t.Fatalf("failed to set state: %v", diags)
	}
	resp := resource.DeleteResponse{State: state}

	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}
