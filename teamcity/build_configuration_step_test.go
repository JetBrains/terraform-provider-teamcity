package teamcity

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildConfigurationStep_basic(t *testing.T) {
	projectName := "TestProjectStep"
	buildConfName := "TestBuildConfStep"
	stepName := "TestStep"

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

                    resource "teamcity_build_configuration_step" "test" {
                        build_configuration_id = teamcity_build_configuration.test.id
                        name                   = "%s"
                        type                   = "simpleRunner"
                        properties = {
                            "script.content"    = "echo Hello World"
                            "use.custom.script" = "true"
                        }
                    }
                `, projectName, buildConfName, stepName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_step.test", "name", stepName),
					resource.TestCheckResourceAttr("teamcity_build_configuration_step.test", "type", "simpleRunner"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_step.test", "properties.script.content", "echo Hello World"),
					resource.TestCheckResourceAttrSet("teamcity_build_configuration_step.test", "id"),
				),
			},
			{
				ResourceName:      "teamcity_build_configuration_step.test",
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

                    resource "teamcity_build_configuration_step" "test" {
                        build_configuration_id = teamcity_build_configuration.test.id
                        name                   = "%s_updated"
                        type                   = "simpleRunner"
                        properties = {
                            "script.content"    = "echo Hello Updated"
                            "use.custom.script" = "true"
                        }
                    }
                `, projectName, buildConfName, stepName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_step.test", "name", stepName+"_updated"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_step.test", "properties.script.content", "echo Hello Updated"),
				),
			},
		},
	})
}
