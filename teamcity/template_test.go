package teamcity

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTemplate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "template_test_project"
}

resource "teamcity_template" "test" {
	name       = "test_template"
	project_id = teamcity_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_template.test", "name", "test_template"),
					resource.TestCheckResourceAttr("teamcity_template.test", "project_id", "TemplateTestProject"),
					resource.TestCheckResourceAttr("teamcity_template.test", "description", ""),
					resource.TestCheckResourceAttrSet("teamcity_template.test", "id"),
				),
			},
			// Second plan with no changes should be empty
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "template_test_project"
}

resource "teamcity_template" "test" {
	name       = "test_template"
	project_id = teamcity_project.test.id
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// In-place update of name and description
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "template_test_project"
}

resource "teamcity_template" "test" {
	name        = "test_template_updated"
	project_id  = teamcity_project.test.id
	description = "updated description"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_template.test", "name", "test_template_updated"),
					resource.TestCheckResourceAttr("teamcity_template.test", "description", "updated description"),
				),
			},
			// Import
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "template_test_project"
}

resource "teamcity_template" "test" {
	name        = "test_template_updated"
	project_id  = teamcity_project.test.id
	description = "updated description"
}
`,
				ResourceName:      "teamcity_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Change ID (RequiresReplace)
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "template_test_project"
}

resource "teamcity_template" "id_change" {
	id         = "custom_template_id"
	name       = "template_name1"
	project_id = teamcity_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_template.id_change", "id", "custom_template_id"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "test" {
	name = "template_test_project"
}

resource "teamcity_template" "id_change" {
	id         = "new_custom_template_id"
	name       = "template_name1"
	project_id = teamcity_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_template.id_change", "id", "new_custom_template_id"),
				),
			},
			// Move between projects (RequiresReplace on project_id)
			{
				Config: providerConfig + `
resource "teamcity_project" "tp1" {
	name = "template_p1"
}

resource "teamcity_project" "tp2" {
	name = "template_p2"
}

resource "teamcity_template" "move" {
	name       = "move_template"
	project_id = teamcity_project.tp1.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_template.move", "project_id", "TemplateP1"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "tp1" {
	name = "template_p1"
}

resource "teamcity_project" "tp2" {
	name = "template_p2"
}

resource "teamcity_template" "move" {
	name       = "move_template"
	project_id = teamcity_project.tp2.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_template.move", "project_id", "TemplateP2"),
				),
			},
		},
	})
}
