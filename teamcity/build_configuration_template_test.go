package teamcity

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccBuildConfigurationTemplateAttachment_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "attach_test_project"
}

resource "teamcity_template" "test" {
	name       = "attach_test_template"
	project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration" "test" {
	name       = "attach_test_bc"
	project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration_template" "test" {
	build_configuration_id = teamcity_build_configuration.test.id
	template_id            = teamcity_template.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_template.test", "build_configuration_id", "AttachTestProject_AttachTestBc"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_template.test", "template_id", "AttachTestProject_AttachTestTemplate"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_template.test", "id", "AttachTestProject_AttachTestBc/AttachTestProject_AttachTestTemplate"),
				),
			},
			// Second plan with no changes should be empty
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "attach_test_project"
}

resource "teamcity_template" "test" {
	name       = "attach_test_template"
	project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration" "test" {
	name       = "attach_test_bc"
	project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration_template" "test" {
	build_configuration_id = teamcity_build_configuration.test.id
	template_id            = teamcity_template.test.id
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Import with composite id
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "attach_test_project"
}

resource "teamcity_template" "test" {
	name       = "attach_test_template"
	project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration" "test" {
	name       = "attach_test_bc"
	project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration_template" "test" {
	build_configuration_id = teamcity_build_configuration.test.id
	template_id            = teamcity_template.test.id
}
`,
				ResourceName:      "teamcity_build_configuration_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Switching to another template replaces the attachment
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "attach_test_project"
}

resource "teamcity_template" "test" {
	name       = "attach_test_template"
	project_id = teamcity_project.test.id
}

resource "teamcity_template" "second" {
	name       = "attach_second_template"
	project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration" "test" {
	name       = "attach_test_bc"
	project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration_template" "test" {
	build_configuration_id = teamcity_build_configuration.test.id
	template_id            = teamcity_template.second.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_template.test", "template_id", "AttachTestProject_AttachSecondTemplate"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_template.test", "id", "AttachTestProject_AttachTestBc/AttachTestProject_AttachSecondTemplate"),
				),
			},
			// Detaching leaves the build configuration and template in place
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "attach_test_project"
}

resource "teamcity_template" "test" {
	name       = "attach_test_template"
	project_id = teamcity_project.test.id
}

resource "teamcity_template" "second" {
	name       = "attach_second_template"
	project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration" "test" {
	name       = "attach_test_bc"
	project_id = teamcity_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration.test", "name", "attach_test_bc"),
					resource.TestCheckResourceAttr("teamcity_template.test", "name", "attach_test_template"),
				),
			},
		},
	})
}
