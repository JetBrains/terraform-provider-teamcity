package teamcity

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccProject_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_project.test", "name", "test"),
					resource.TestCheckResourceAttr("teamcity_project.test", "id", "Test"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_project.test", "name", "test2"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test2"
	id = "new"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_project.test", "id", "new"),
				),
			},

			//TW-88034
			{
				Config: providerConfig + `
resource "teamcity_project" "parent" {
	name = "parent"
	id = "parent_project"
}

resource "teamcity_project" "child" {
	name = "child"
	id = "child_project"
	parent_project_id = teamcity_project.parent.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_project.child", "id", "child_project"),
					resource.TestCheckResourceAttr("teamcity_project.child", "parent_project_id", "parent_project"),
					resource.TestCheckResourceAttr("teamcity_project.parent", "parent_project_id", "_Root"),
				),
			},
			//TW-88034 ^
		},
	})
}
