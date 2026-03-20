package teamcity

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccGroupResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read testing with description
			{
				Config: providerConfig + `
                    resource "teamcity_group" "test" {
                        name        = "test_group"
                        description = "initial description"
                    }
                `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_group.test", "name", "test_group"),
					resource.TestCheckResourceAttr("teamcity_group.test", "description", "initial description"),
				),
			},
			// Update description
			{
				Config: providerConfig + `
                    resource "teamcity_group" "test" {
                        name        = "test_group"
                        description = "updated description"
                    }
                `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_group.test", "name", "test_group"),
					resource.TestCheckResourceAttr("teamcity_group.test", "description", "updated description"),
				),
			},
			// Remove description
			{
				Config: providerConfig + `
                    resource "teamcity_group" "test" {
                        name        = "test_group"
                    }
                `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_group.test", "name", "test_group"),
					resource.TestCheckNoResourceAttr("teamcity_group.test", "description"),
				),
			},
		},
	})
}

func TestAccGroupDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
                    resource "teamcity_group" "test" {
                        name        = "test_group_ds"
                        description = "ds description"
                    }
                    data "teamcity_group" "test" {
                        key = teamcity_group.test.id
                    }
                `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.teamcity_group.test", "name", "test_group_ds"),
					resource.TestCheckResourceAttr("data.teamcity_group.test", "description", "ds description"),
				),
			},
		},
	})
}
