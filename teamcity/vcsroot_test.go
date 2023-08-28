package teamcity

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	name = "test1"
	id = "test1"
	project_id = "_Root"
	git = {
		url = "git@github.com:mkuzmin/test1.git"
		branch = "main"	
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "name", "test1"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "id", "test1"),
					//resource.TestCheckResourceAttr("teamcity_vcsroot.test", "project_id", "_Root"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "git.url", "git@github.com:mkuzmin/test1.git"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "git.branch", "main"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_vcsroot" "test" {
	name = "test1"
	project_id = "_Root"
	git = {
		url = "git@github.com:mkuzmin/test1.git"
		branch = "main"
		
		auth_method = "PASSWORD"
	    username    = "git"
    	password    = "1234"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "git.auth_method", "PASSWORD"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "git.username", "git"),
					resource.TestCheckResourceAttr("teamcity_vcsroot.test", "git.password", "1234"),
				),
			},
		},
	})
}
