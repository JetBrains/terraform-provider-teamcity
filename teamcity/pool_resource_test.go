package teamcity

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"regexp"
	"testing"
)

func TestAccPoolResource_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Invalid create and read testing
			{
				Config: providerConfig + `
                    resource "teamcity_pool" "test" {
                        name = "test_pool"
                        size = -20
                    }
                    `,
				ExpectError: regexp.MustCompile("Attribute size value must be at least 0"),
			},
			// Create and read testing
			{
				Config: providerConfig + `
                    resource "teamcity_pool" "test" {
                        name = "test_pool"
                        size = 10
                    }
                `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_pool.test", "name", "test_pool"),
					resource.TestCheckResourceAttr("teamcity_pool.test", "size", "10"),
					resource.TestCheckResourceAttr("teamcity_pool.test", "projects.#", "0"),
				),
			},
			// Invalid update and read testing
			{
				Config: providerConfig + `
                    resource "teamcity_pool" "test" {
                        name = "test_pool_renamed"
                        size = -20
                    }
                    `,
				ExpectError: regexp.MustCompile("Attribute size value must be at least 0"),
			},
			// Update and read testing
			{
				Config: providerConfig + `
                    resource "teamcity_pool" "test" {
                        name = "test_pool_renamed"
                        size = 20
                    }
                    `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_pool.test", "name", "test_pool_renamed"),
					resource.TestCheckResourceAttr("teamcity_pool.test", "size", "20"),
				),
			},
			// Create pool with projects testing
			{
				Config: providerConfig + `
                    resource "teamcity_project" "demo" {
                        name = "Demo Project"
                    }
                    resource "teamcity_pool" "test" {
                        name = "Demo Pool"
                        projects = [ teamcity_project.demo.id ]
                    }
                `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_pool.test", "name", "Demo Pool"),
					resource.TestCheckResourceAttr("teamcity_pool.test", "projects.0", "DemoProject"),
				),
			},
			// Update pool with projects
			{
				Config: providerConfig + `
                    resource "teamcity_project" "demo" {
                        name = "Demo Project"
                    }
                    resource "teamcity_project" "demo2" {
                        name = "Demo Project 2"
                    }
                    resource "teamcity_pool" "test" {
                        name = "Demo Pool"
                        projects = [ teamcity_project.demo.id, teamcity_project.demo2.id ]
                    }
                `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_pool.test", "name", "Demo Pool"),
					resource.TestCheckResourceAttr("teamcity_pool.test", "projects.0", "DemoProject"),
					resource.TestCheckResourceAttr("teamcity_pool.test", "projects.1", "DemoProject2"),
				),
			},
		},
	})
}
