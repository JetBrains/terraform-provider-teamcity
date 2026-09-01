package teamcity

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/models"
)

func TestGroupPayloadOmitsRolesWhenNotConfigured(t *testing.T) {
	plan := models.GroupDataModel{Name: types.StringValue("test_group")}
	payload := groupPayload(plan)

	assertRolesJSON(t, payload, false, "")
}

func TestGroupPayloadIncludesConfiguredRoles(t *testing.T) {
	plan := models.GroupDataModel{
		Name: types.StringValue("test_group"),
		Roles: []models.RoleAssignmentGroupDataModel{
			{Id: types.StringValue("PROJECT_ADMIN"), Global: types.BoolValue(true)},
		},
	}
	payload := groupPayload(plan)

	assertRolesJSON(t, payload, true, `{"role":[{"roleId":"PROJECT_ADMIN","scope":"g"}]}`)
}

func TestUserPayloadOmitsRolesWhenNotConfigured(t *testing.T) {
	resource := userResource{}
	payload := resource.update(userResourceModel{Username: types.StringValue("test_user")})

	assertRolesJSON(t, payload, false, "")
}

func TestUserPayloadIncludesConfiguredRoles(t *testing.T) {
	resource := userResource{}
	payload := resource.update(userResourceModel{
		Username: types.StringValue("test_user"),
		Roles: []roleAssignment{
			{Id: types.StringValue("PROJECT_ADMIN"), Project: types.StringValue("project-id")},
		},
	})

	assertRolesJSON(t, payload, true, `{"role":[{"roleId":"PROJECT_ADMIN","scope":"p:project-id"}]}`)
}

func assertRolesJSON(t *testing.T, payload any, wantPresent bool, wantRoles string) {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatal(err)
	}

	roles, present := object["roles"]
	if present != wantPresent {
		t.Fatalf("roles presence = %t, want %t; payload: %s", present, wantPresent, body)
	}
	if wantPresent && string(roles) != wantRoles {
		t.Fatalf("roles = %s, want %s; payload: %s", roles, wantRoles, body)
	}
}
