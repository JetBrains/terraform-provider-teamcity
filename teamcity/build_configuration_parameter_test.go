package teamcity

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Acceptance tests for build configuration parameters (text and password types)

func TestAccBuildConfigurationParameter_text_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create text parameter without explicit type; expect default to "text"
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "BC Param Project"
  id   = "bc_param_project"
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "BC Param Test"
  id         = "bc_param_test"
}

resource "teamcity_build_configuration_parameter" "param1" {
  build_configuration_id = teamcity_build_configuration.bc.id
  name  = "MY_PARAM"
  value = "v1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_parameter.param1", "name", "MY_PARAM"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_parameter.param1", "type", "text"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_parameter.param1", "value", "v1"),
				),
			},
			// Update text parameter value
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "BC Param Project"
  id   = "bc_param_project"
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "BC Param Test"
  id         = "bc_param_test"
}

resource "teamcity_build_configuration_parameter" "param1" {
  build_configuration_id = teamcity_build_configuration.bc.id
  name  = "MY_PARAM"
  value = "v2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_parameter.param1", "value", "v2"),
				),
			},
			// Import and verify state for text parameter
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "BC Param Project"
  id   = "bc_param_project"
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "BC Param Test"
  id         = "bc_param_test"
}

resource "teamcity_build_configuration_parameter" "param1" {
  build_configuration_id = teamcity_build_configuration.bc.id
  name  = "MY_PARAM"
  value = "v2"
}
`,
				ResourceName:      "teamcity_build_configuration_parameter.param1",
				ImportState:       true,
				ImportStateId:     "bc_param_test/MY_PARAM",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBuildConfigurationParameter_password_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create secure parameter; server does not echo value, provider keeps state value
			{
				Config: providerConfig + `
resource "teamcity_project" "p2" {
  name = "BC Param Project 2"
  id   = "bc_param_project2"
}

resource "teamcity_build_configuration" "bc2" {
  project_id = teamcity_project.p2.id
  name       = "BC Param Test 2"
  id         = "bc_param_test2"
}

resource "teamcity_build_configuration_parameter" "secret" {
  build_configuration_id = teamcity_build_configuration.bc2.id
  name  = "SECRET_TOKEN"
  value = "s3cr3t"
  type  = "password"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_parameter.secret", "name", "SECRET_TOKEN"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_parameter.secret", "type", "password"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_parameter.secret", "value", "s3cr3t"),
				),
			},
			// Update secure value; ensure state reflects new sensitive value
			{
				Config: providerConfig + `
resource "teamcity_project" "p2" {
  name = "BC Param Project 2"
  id   = "bc_param_project2"
}

resource "teamcity_build_configuration" "bc2" {
  project_id = teamcity_project.p2.id
  name       = "BC Param Test 2"
  id         = "bc_param_test2"
}

resource "teamcity_build_configuration_parameter" "secret" {
  build_configuration_id = teamcity_build_configuration.bc2.id
  name  = "SECRET_TOKEN"
  value = "n3w"
  type  = "password"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_parameter.secret", "type", "password"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_parameter.secret", "value", "n3w"),
				),
			},
		},
	})
}
