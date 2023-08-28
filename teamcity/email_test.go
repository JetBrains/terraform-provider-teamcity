package teamcity

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccEmail_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_email_settings" "test" {
  enabled = true
  host = "teamcity"
  port = 587
  from = "TeamCity"
  login = "teamcity"
  password = "password"
  secure_connection = "STARTTLS"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_email_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("teamcity_email_settings.test", "host", "teamcity"),
					resource.TestCheckResourceAttr("teamcity_email_settings.test", "port", "587"),
					resource.TestCheckResourceAttr("teamcity_email_settings.test", "from", "TeamCity"),
					resource.TestCheckResourceAttr("teamcity_email_settings.test", "login", "teamcity"),
					resource.TestCheckResourceAttr("teamcity_email_settings.test", "password", "password"),
					resource.TestCheckResourceAttr("teamcity_email_settings.test", "secure_connection", "STARTTLS"),
				),
			},
		},
	})
}
