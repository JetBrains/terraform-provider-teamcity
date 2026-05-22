package teamcity

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildConfiguration_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test_project"
}

resource "teamcity_build_configuration" "test" {
	name       = "test_bc"
	project_id = teamcity_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration.test", "name", "test_bc"),
					resource.TestCheckResourceAttr("teamcity_build_configuration.test", "project_id", "TestProject"),
					resource.TestCheckResourceAttr("teamcity_build_configuration.test", "build_type", "regular"),
					resource.TestCheckResourceAttr("teamcity_build_configuration.test", "paused", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test_project"
}

resource "teamcity_build_configuration" "test" {
	name        = "test_bc_updated"
	project_id  = teamcity_project.test.id
	description = "updated description"
	paused      = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration.test", "name", "test_bc_updated"),
					resource.TestCheckResourceAttr("teamcity_build_configuration.test", "description", "updated description"),
					resource.TestCheckResourceAttr("teamcity_build_configuration.test", "paused", "true"),
				),
			},
			// Import
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test_project"
}

resource "teamcity_build_configuration" "test" {
	name        = "test_bc_updated"
	project_id  = teamcity_project.test.id
	description = "updated description"
	paused      = true
}
`,
				ResourceName:      "teamcity_build_configuration.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Composite build type
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test_project"
}

resource "teamcity_build_configuration" "composite" {
	name       = "composite_bc"
	project_id = teamcity_project.test.id
	build_type = "composite"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration.composite", "build_type", "composite"),
				),
			},
			// Move between projects
			{
				Config: providerConfig + `
resource "teamcity_project" "p1" {
	name = "p1"
}

resource "teamcity_project" "p2" {
	name = "p2"
}

resource "teamcity_build_configuration" "move" {
	name       = "move_bc"
	project_id = teamcity_project.p1.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration.move", "project_id", "P1"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "p1" {
	name = "p1"
}

resource "teamcity_project" "p2" {
	name = "p2"
}

resource "teamcity_build_configuration" "move" {
	name       = "move_bc"
	project_id = teamcity_project.p2.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration.move", "project_id", "P2"),
				),
			},
			// Change ID (RequiresReplace)
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test_project"
}

resource "teamcity_build_configuration" "id_change" {
	id         = "custom_id"
	name       = "name1"
	project_id = teamcity_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration.id_change", "id", "custom_id"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test_project"
}

resource "teamcity_build_configuration" "id_change" {
	id         = "new_custom_id"
	name       = "name1"
	project_id = teamcity_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration.id_change", "id", "new_custom_id"),
				),
			},
		},
	})
}

func TestAccBuildConfigurationDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "test_project_ds"
}

resource "teamcity_build_configuration" "test" {
	name       = "test_bc_ds"
	project_id = teamcity_project.test.id
}

data "teamcity_build_configuration" "test" {
	id = teamcity_build_configuration.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.teamcity_build_configuration.test", "name", "test_bc_ds"),
					resource.TestCheckResourceAttr("data.teamcity_build_configuration.test", "project_id", "TestProjectDs"),
					resource.TestCheckResourceAttr("data.teamcity_build_configuration.test", "build_type", "regular"),
				),
			},
		},
	})
}
