package teamcity

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBuildConfigurationVcsRoot_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "VCS Project"
  id   = "vcs_project"
}

resource "teamcity_vcsroot" "v" {
  name       = "My VCS Root"
  id         = "my_vcs_root"
  project_id = teamcity_project.p.id
  git = {
    url    = "https://github.com/jetbrains/teamcity"
    branch = "refs/heads/master"
  }
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "VCS BC"
  id         = "vcs_bc"
}

resource "teamcity_build_configuration_vcs_root" "attach" {
  build_configuration_id = teamcity_build_configuration.bc.id
  vcs_root_id            = teamcity_vcsroot.v.id
  checkout_rules         = "+:.=>somewhere"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_vcs_root.attach", "vcs_root_id", "my_vcs_root"),
					resource.TestCheckResourceAttr("teamcity_build_configuration_vcs_root.attach", "checkout_rules", "+:.=>somewhere"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_project" "p" {
  name = "VCS Project"
  id   = "vcs_project"
}

resource "teamcity_vcsroot" "v" {
  name       = "My VCS Root"
  id         = "my_vcs_root"
  project_id = teamcity_project.p.id
  git = {
    url    = "https://github.com/jetbrains/teamcity"
    branch = "refs/heads/master"
  }
}

resource "teamcity_build_configuration" "bc" {
  project_id = teamcity_project.p.id
  name       = "VCS BC"
  id         = "vcs_bc"
}

resource "teamcity_build_configuration_vcs_root" "attach" {
  build_configuration_id = teamcity_build_configuration.bc.id
  vcs_root_id            = teamcity_vcsroot.v.id
  checkout_rules         = "+:lib=>libs"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_build_configuration_vcs_root.attach", "checkout_rules", "+:lib=>libs"),
				),
			},
		},
	})
}
