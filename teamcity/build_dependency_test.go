package teamcity

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSnapshotDependency_basic(t *testing.T) {
	projectName := "TestProjectSnap"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
                    resource "teamcity_project" "test" {
                        name = "%s"
                    }

                    resource "teamcity_build_configuration" "upstream" {
                        name       = "Upstream"
                        project_id = teamcity_project.test.id
                    }

                    resource "teamcity_build_configuration" "downstream" {
                        name       = "Downstream"
                        project_id = teamcity_project.test.id
                    }

                    resource "teamcity_build_configuration_snapshot_dependency" "test" {
                        build_configuration_id = teamcity_build_configuration.downstream.id
                        depends_on_id          = teamcity_build_configuration.upstream.id
                        properties = {
                            "run-build-on-the-same-agent" = "true"
                        }
                    }
                `, projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("teamcity_build_configuration_snapshot_dependency.test", "id"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_snapshot_dependency.test", "depends_on_id", "TestProjectSnap_Upstream"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_snapshot_dependency.test", "properties.run-build-on-the-same-agent", "true"),
				),
			},
			{
				ResourceName:            "teamcity_build_configuration_snapshot_dependency.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"properties"},
			},
		},
	})
}

func TestAccArtifactDependency_basic(t *testing.T) {
	projectName := "TestProjectArt"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
                    resource "teamcity_project" "test" {
                        name = "%s"
                    }

                    resource "teamcity_build_configuration" "upstream" {
                        name       = "Upstream"
                        project_id = teamcity_project.test.id
                    }

                    resource "teamcity_build_configuration" "downstream" {
                        name       = "Downstream"
                        project_id = teamcity_project.test.id
                    }

                    resource "teamcity_build_configuration_artifact_dependency" "test" {
                        build_configuration_id = teamcity_build_configuration.downstream.id
                        depends_on_id          = teamcity_build_configuration.upstream.id
                        properties = {
                            "pathRules"    = "*.jar => lib"
                            "revisionName" = "lastSuccessful"
                        }
                    }
                `, projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("teamcity_build_configuration_artifact_dependency.test", "id"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_artifact_dependency.test", "depends_on_id", "TestProjectArt_Upstream"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_artifact_dependency.test", "properties.pathRules", "*.jar => lib"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_artifact_dependency.test", "properties.revisionName", "lastSuccessful"),
				),
			},
			{
				ResourceName:            "teamcity_build_configuration_artifact_dependency.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"properties"},
			},
		},
	})
}
