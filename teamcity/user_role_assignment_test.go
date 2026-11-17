package teamcity

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"terraform-provider-teamcity/client"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUserRoleAssignment_parallel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "proj_a" {
  id   = "AccRoleProjA"
  name = "Acc Role Project A"
}

resource "teamcity_project" "proj_b" {
  id   = "AccRoleProjB"
  name = "Acc Role Project B"
}

resource "teamcity_user" "racer" {
  username = "acc_racer_parallel"
  lifecycle {
    ignore_changes = [roles]
  }
}

resource "teamcity_user_role_assignment" "sys_admin_global" {
  user_id  = teamcity_user.racer.id
  username = teamcity_user.racer.username
  role_id  = "SYSTEM_ADMIN"
  scope    = "g"
}

resource "teamcity_user_role_assignment" "proj_a_viewer" {
  user_id  = teamcity_user.racer.id
  username = teamcity_user.racer.username
  role_id  = "PROJECT_VIEWER"
  scope    = "p:${teamcity_project.proj_a.id}"
}

resource "teamcity_user_role_assignment" "proj_a_dev" {
  user_id  = teamcity_user.racer.id
  username = teamcity_user.racer.username
  role_id  = "PROJECT_DEVELOPER"
  scope    = "p:${teamcity_project.proj_a.id}"
}

resource "teamcity_user_role_assignment" "proj_a_admin" {
  user_id  = teamcity_user.racer.id
  username = teamcity_user.racer.username
  role_id  = "PROJECT_ADMIN"
  scope    = "p:${teamcity_project.proj_a.id}"
}

resource "teamcity_user_role_assignment" "proj_b_viewer" {
  user_id  = teamcity_user.racer.id
  username = teamcity_user.racer.username
  role_id  = "PROJECT_VIEWER"
  scope    = "p:${teamcity_project.proj_b.id}"
}

resource "teamcity_user_role_assignment" "proj_b_agent_mgr" {
  user_id  = teamcity_user.racer.id
  username = teamcity_user.racer.username
  role_id  = "AGENT_MANAGER"
  scope    = "p:${teamcity_project.proj_b.id}"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.sys_admin_global", "role_id", "SYSTEM_ADMIN"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.sys_admin_global", "scope", "g"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_a_viewer", "role_id", "PROJECT_VIEWER"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_a_viewer", "scope", "p:AccRoleProjA"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_a_dev", "role_id", "PROJECT_DEVELOPER"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_a_dev", "scope", "p:AccRoleProjA"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_a_admin", "role_id", "PROJECT_ADMIN"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_a_admin", "scope", "p:AccRoleProjA"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_b_viewer", "role_id", "PROJECT_VIEWER"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_b_viewer", "scope", "p:AccRoleProjB"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_b_agent_mgr", "role_id", "AGENT_MANAGER"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_b_agent_mgr", "scope", "p:AccRoleProjB"),
					testAccCheckUserRoles("acc_racer_parallel", []string{
						"AGENT_MANAGER@p:AccRoleProjB",
						"PROJECT_ADMIN@p:AccRoleProjA",
						"PROJECT_DEVELOPER@p:AccRoleProjA",
						"PROJECT_VIEWER@p:AccRoleProjA",
						"PROJECT_VIEWER@p:AccRoleProjB",
						"SYSTEM_ADMIN@g",
					}),
				),
			},
			{
				ResourceName:      "teamcity_user_role_assignment.proj_a_admin",
				ImportState:       true,
				ImportStateId:     "acc_racer_parallel/PROJECT_ADMIN/p:AccRoleProjA",
				ImportStateVerify: true,
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "proj_a" {
  id   = "AccRoleProjA"
  name = "Acc Role Project A"
}

resource "teamcity_project" "proj_b" {
  id   = "AccRoleProjB"
  name = "Acc Role Project B"
}

resource "teamcity_user" "racer" {
  username = "acc_racer_parallel"
  lifecycle {
    ignore_changes = [roles]
  }
}

resource "teamcity_user_role_assignment" "sys_admin_global" {
  user_id  = teamcity_user.racer.id
  username = teamcity_user.racer.username
  role_id  = "SYSTEM_ADMIN"
  scope    = "g"
}

resource "teamcity_user_role_assignment" "proj_a_viewer" {
  user_id  = teamcity_user.racer.id
  username = teamcity_user.racer.username
  role_id  = "PROJECT_VIEWER"
  scope    = "p:${teamcity_project.proj_a.id}"
}

resource "teamcity_user_role_assignment" "proj_a_admin" {
  user_id  = teamcity_user.racer.id
  username = teamcity_user.racer.username
  role_id  = "PROJECT_ADMIN"
  scope    = "p:${teamcity_project.proj_a.id}"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.sys_admin_global", "role_id", "SYSTEM_ADMIN"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_a_viewer", "role_id", "PROJECT_VIEWER"),
					resource.TestCheckResourceAttr("teamcity_user_role_assignment.proj_a_admin", "role_id", "PROJECT_ADMIN"),
					testAccCheckUserRoles("acc_racer_parallel", []string{
						"PROJECT_ADMIN@p:AccRoleProjA",
						"PROJECT_VIEWER@p:AccRoleProjA",
						"SYSTEM_ADMIN@g",
					}),
				),
			},
		},
	})
}

func testAccCheckUserRoles(username string, want []string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.NewClient(
			os.Getenv("TEAMCITY_HOST"),
			os.Getenv("TEAMCITY_TOKEN"),
			os.Getenv("TEAMCITY_USER"),
			os.Getenv("TEAMCITY_PASSWORD"),
			0,
		)
		user, err := c.GetUserByName(username)
		if err != nil {
			return fmt.Errorf("failed to get user %q: %w", username, err)
		}
		if user == nil {
			return fmt.Errorf("user %q not found", username)
		}

		var got []string
		if user.Roles != nil {
			for _, r := range user.Roles.RoleAssignment {
				got = append(got, r.Id+"@"+r.Scope)
			}
		}
		sort.Strings(got)
		sortedWant := append([]string(nil), want...)
		sort.Strings(sortedWant)

		if !reflect.DeepEqual(got, sortedWant) {
			return fmt.Errorf("user %q roles = %v, want %v", username, got, sortedWant)
		}
		return nil
	}
}
