package teamcity

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildConfigurationTrigger_basic(t *testing.T) {
	projectName := "TestProjectTrigger"
	buildConfName := "TestBuildConfTrigger"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
                    resource "teamcity_project" "test" {
                        name = "%s"
                    }

                    resource "teamcity_build_configuration" "test" {
                        name       = "%s"
                        project_id = teamcity_project.test.id
                    }

                    resource "teamcity_build_configuration_trigger" "test" {
                        build_configuration_id = teamcity_build_configuration.test.id
                        type                   = "vcsTrigger"
                        properties = {
                            "quietPeriodMode" = "DO_NOT_USE"
                        }
                    }
                `, projectName, buildConfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("teamcity_build_configuration_trigger.test", "id"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_trigger.test", "type", "vcsTrigger"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_trigger.test", "properties.quietPeriodMode", "DO_NOT_USE"),
				),
			},
			{
				ResourceName:      "teamcity_build_configuration_trigger.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: providerConfig + fmt.Sprintf(`
                    resource "teamcity_project" "test" {
                        name = "%s"
                    }

                    resource "teamcity_build_configuration" "test" {
                        name       = "%s"
                        project_id = teamcity_project.test.id
                    }

                    resource "teamcity_build_configuration_trigger" "test" {
                        build_configuration_id = teamcity_build_configuration.test.id
                        type                   = "vcsTrigger"
                        properties = {
                            "quietPeriodMode" = "USE_DEFAULT"
                        }
                    }
                `, projectName, buildConfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_trigger.test", "properties.quietPeriodMode", "USE_DEFAULT"),
				),
			},
		},
	})
}
