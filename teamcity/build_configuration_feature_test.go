package teamcity

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildConfigurationFeature_basic(t *testing.T) {
	projectName := "TestProjectFeature"
	buildConfName := "TestBuildConfFeature"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource "teamcity_project" "test" {
                        name = "%s"
                    }

                    resource "teamcity_build_configuration" "test" {
                        name       = "%s"
                        project_id = teamcity_project.test.id
                    }

                    resource "teamcity_build_configuration_feature" "test" {
                        build_configuration_id = teamcity_build_configuration.test.id
                        type                   = "swabra"
                        properties = {
                            "swabra.enabled" = "true"
                            "swabra.strict"  = "true"
                        }
                    }
                `, projectName, buildConfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("teamcity_build_configuration_feature.test", "id"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_feature.test", "type", "swabra"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_feature.test", "properties.swabra.enabled", "true"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_feature.test", "properties.swabra.strict", "true"),
				),
			},
			{
				ResourceName:      "teamcity_build_configuration_feature.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
                    resource "teamcity_project" "test" {
                        name = "%s"
                    }

                    resource "teamcity_build_configuration" "test" {
                        name       = "%s"
                        project_id = teamcity_project.test.id
                    }

                    resource "teamcity_build_configuration_feature" "test" {
                        build_configuration_id = teamcity_build_configuration.test.id
                        type                   = "swabra"
                        properties = {
                            "swabra.enabled" = "false"
                        }
                    }
                `, projectName, buildConfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_feature.test", "properties.swabra.enabled", "false"),
					resource.TestCheckNoResourceAttr("teamcity_build_configuration_feature.test", "properties.swabra.strict"),
				),
			},
		},
	})
}
