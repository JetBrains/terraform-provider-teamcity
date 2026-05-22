package teamcity

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildConfigurationSettings_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "Settings Project"
  id   = "settings_project"
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "Settings BC"
  id         = "settings_bc"
}

resource "teamcity_build_configuration_settings" "s" {
  build_configuration_id = teamcity_build_configuration.bc.id
  build_number_counter   = 123
  build_number_pattern   = "v%build.counter%"
  artifact_rules         = "+:target/*.jar"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_settings.s", "build_number_counter", "123"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_settings.s", "build_number_pattern", "v%build.counter%"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_settings.s", "artifact_rules", "+:target/*.jar"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "Settings Project"
  id   = "settings_project"
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "Settings BC"
  id         = "settings_bc"
}

resource "teamcity_build_configuration_settings" "s" {
  build_configuration_id = teamcity_build_configuration.bc.id
  build_number_counter   = 456
  build_number_pattern   = "release-%build.counter%"
  artifact_rules         = "+:dist/*.zip"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_settings.s", "build_number_counter", "456"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_settings.s", "build_number_pattern", "release-%build.counter%"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_settings.s", "artifact_rules", "+:dist/*.zip"),
				),
			},
		},
	})
}
