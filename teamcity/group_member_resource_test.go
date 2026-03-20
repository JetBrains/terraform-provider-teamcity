package teamcity

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"
)

func TestAccGroupMemberResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create group and user first
			{
				Config: providerConfig + `
                    resource "teamcity_group" "test" {
                        name = "test_membership_group"
                    }

                    resource "teamcity_user" "test" {
                        username = "test_member_user"
                    }

                    resource "teamcity_group_member" "test" {
                        group_id = teamcity_group.test.id
                        username = teamcity_user.test.username
                    }
                `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("teamcity_group_member.test", "group_id", "teamcity_group.test", "id"),
					resource.TestCheckResourceAttr("teamcity_group_member.test", "username", "test_member_user"),
				),
			},
			// Import testing
			{
				ResourceName:      "teamcity_group_member.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["teamcity_group_member.test"]
					if !ok {
						return "", nil
					}
					return rs.Primary.Attributes["group_id"] + "/" + rs.Primary.Attributes["username"], nil
				},
			},
		},
	})
}
