package teamcity

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgentRequirement_basic(t *testing.T) {
	projectName := "TestProjectAR"
	buildConfName := "TestBuildConfAR"

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

                    resource "teamcity_build_configuration_agent_requirement" "test" {
                        build_configuration_id = teamcity_build_configuration.test.id
                        condition              = "equals"
                        name                   = "os.name"
                        value                  = "Linux"
                    }
                `, projectName, buildConfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("teamcity_build_configuration_agent_requirement.test", "id"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_agent_requirement.test", "condition", "equals"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_agent_requirement.test", "name", "os.name"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_agent_requirement.test", "value", "Linux"),
				),
			},
			{
				ResourceName:      "teamcity_build_configuration_agent_requirement.test",
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

                    resource "teamcity_build_configuration_agent_requirement" "test" {
                        build_configuration_id = teamcity_build_configuration.test.id
                        condition              = "exists"
                        name                   = "docker.server.version"
                    }
                `, projectName, buildConfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_agent_requirement.test", "condition", "exists"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_agent_requirement.test", "name", "docker.server.version"),
					resource.TestCheckNoResourceAttr("teamcity_build_configuration_agent_requirement.test", "value"),
				),
			},
		},
	})
}
