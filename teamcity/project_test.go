package teamcity

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccProject_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
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
				Config: `
resource "teamcity_project" "test" {
	name = "test2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_project.test", "name", "test2"),
				),
			},
			{
				Config: `
resource "teamcity_project" "test" {
	name = "test2"
	id = "new"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_project.test", "id", "new"),
				),
			},
		},
	})
}
