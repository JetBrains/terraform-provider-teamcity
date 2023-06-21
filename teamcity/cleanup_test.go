package teamcity

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccCleanup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "teamcity_cleanup" "test" {
	enabled = true
	max_duration = 0
	daily = {
		hour = 2
		minute = 15
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "enabled", "true"),
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "max_duration", "0"),
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "daily.hour", "2"),
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "daily.minute", "15"),
				),
			},
			{
				Config: `
resource "teamcity_cleanup" "test" {
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
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "enabled", "true"),
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "max_duration", "0"),
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "cron.minute", "15"),
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "cron.hour", "2"),
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "cron.day", "2"),
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "cron.month", "*"),
					resource.TestCheckResourceAttr("teamcity_cleanup.test", "cron.day_week", "?"),
				),
			},
		},
	})
}
