package teamcity

import (
	"fmt"
	"os"
	"terraform-provider-teamcity/client"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func TestAccBuildConfigurationSettings_delete(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "Delete Settings Project"
  id   = "del_settings_project"
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "Delete Settings BC"
  id         = "del_settings_bc"
}

resource "teamcity_build_configuration_settings" "s" {
  build_configuration_id = teamcity_build_configuration.bc.id
  build_number_counter   = 789
  build_number_pattern   = "del-%build.counter%"
  artifact_rules         = "+:del/*"
}
`,
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "Delete Settings Project"
  id   = "del_settings_project"
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "Delete Settings BC"
  id         = "del_settings_bc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBuildConfigurationSettingsReset("del_settings_bc"),
				),
			},
		},
	})
}

func testAccCheckBuildConfigurationSettingsReset(buildTypeId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		host := os.Getenv("TEAMCITY_HOST")
		password := os.Getenv("TEAMCITY_PASSWORD")
		c := client.NewClient(host, "", "", password, 0)

		counter, err := c.GetBuildTypeSetting(buildTypeId, "buildNumberCounter")
		if err != nil {
			return err
		}
		if counter == nil || *counter != "1" {
			return fmt.Errorf("expected buildNumberCounter to be 1, got %v", counter)
		}

		pattern, err := c.GetBuildTypeSetting(buildTypeId, "buildNumberPattern")
		if err != nil {
			return err
		}
		if pattern == nil || *pattern != "%build.counter%" {
			return fmt.Errorf("expected buildNumberPattern to be %%build.counter%%, got %v", pattern)
		}

		rules, err := c.GetBuildTypeSetting(buildTypeId, "artifactRules")
		if err != nil {
			return err
		}
		if rules == nil || *rules != "" {
			return fmt.Errorf("expected artifactRules to be empty, got %v", rules)
		}

		return nil
	}
}

func TestAccBuildConfigurationSettings_deleteAfterParentRemoval(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "OOB Project"
  id   = "oob_project"
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "OOB BC"
  id         = "oob_bc"
}

resource "teamcity_build_configuration_settings" "s" {
  build_configuration_id = teamcity_build_configuration.bc.id
  build_number_counter   = 789
  build_number_pattern   = "v%build.counter%"
  artifact_rules         = ""
}
`,
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "OOB Project"
  id   = "oob_project"
}
`,
				PreConfig: func() {
					host := os.Getenv("TEAMCITY_HOST")
					password := os.Getenv("TEAMCITY_PASSWORD")
					c := client.NewClient(host, "", "", password, 0)
					_ = c.DeleteBuildType("oob_bc")
				},
			},
		},
	})
}
