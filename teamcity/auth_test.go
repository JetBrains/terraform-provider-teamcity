package teamcity

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccAuth_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_auth" "test" {
  allow_guest         = false
  guest_username      = "guest"
  welcome_text        = ""
  collapse_login_form = false
  two_factor_mode     = "OPTIONAL"
  project_permissions = false
  email_verification  = false

  modules = {
    token = {}
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_auth.test", "allow_guest", "false"),
				),
			},
		},
	})
}
