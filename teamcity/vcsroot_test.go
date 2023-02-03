package teamcity

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccVcsRoot_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_vcsroot" "test" {
	name = "test"
	project_id = "_Root"
	git = {
		url = "git@github.com:mkuzmin/test.git"
		branch = "master"	
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "name", "test"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "id", "Root_Test"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "project_id", "_Root"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "git.url", "git@github.com:mkuzmin/test.git"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "git.branch", "master"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_vcsroot" "test" {
	name = "test"
	project_id = "_Root"
	git = {
		url = "git@github.com:mkuzmin/test.git"
		branch = "main"	
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "git.url", "git@github.com:mkuzmin/test.git"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "git.branch", "main"),
				),
			},
		},
	})
}
