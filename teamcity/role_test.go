package teamcity

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccRole_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_role" "test" {
	name = "Test Role"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_role.test", "name", "Test Role"),
					resource.TestCheckResourceAttr("teamcity_role.test", "id", "TEST_ROLE"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_role" "test" {
	name = "Test Role"
	included = ["PROJECT_DEVELOPER", "AGENT_MANAGER"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_role.test", "included.#", "2"),
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "included.*", "PROJECT_DEVELOPER"),
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "included.*", "AGENT_MANAGER"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_role" "test" {
	name = "Test Role"
	permissions = ["view_all_users", "assign_investigation"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_role.test", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "permissions.*", "view_all_users"),
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "permissions.*", "assign_investigation"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_role" "test" {
	name = "Test Role"
	permissions = ["view_all_users", "view_project"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_role.test", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "permissions.*", "view_all_users"),
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "permissions.*", "view_project"),
				),
			},
		},
	})
}
