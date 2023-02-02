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
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "included.*", "PROJECT_DEVELOPER"),
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "included.*", "AGENT_MANAGER"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_role" "test" {
	name = "Test Role"
	permissions = ["VIEW_ALL_USERS", "ASSIGN_INVESTIGATION"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "permissions.*", "VIEW_ALL_USERS"),
					resource.TestCheckTypeSetElemAttr("teamcity_role.test", "permissions.*", "ASSIGN_INVESTIGATION"),
				),
			},
		},
	})
}
