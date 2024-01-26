package teamcity

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccCleanup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "teamcity_cleanup_settings" "test" {
	enabled = true
	max_duration = 0
	daily = {
		hour = 2
		minute = 15
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "max_duration", "0"),
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "daily.hour", "2"),
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "daily.minute", "15"),
				),
			},
			{
				Config: providerConfig + `
resource "teamcity_cleanup_settings" "test" {
	enabled = true
	max_duration = 0
	cron = {
		minute = 15
		hour = 2
		day = 2
		month = "*"
		day_week = "?"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "max_duration", "0"),
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "cron.minute", "15"),
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "cron.hour", "2"),
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "cron.day", "2"),
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "cron.month", "*"),
					resource.TestCheckResourceAttr("teamcity_cleanup_settings.test", "cron.day_week", "?"),
				),
			},
		},
	})
}
