terraform {
  required_providers {
    teamcity = {
      source = "jetbrains/teamcity"
    }
  }
}
provider "teamcity" {
  host = "http://localhost:8111"
#  token = "" # env.TEAMCITY_TOKEN
}

data "teamcity_server" "buildserver" {}

output "server_version" {
  value = data.teamcity_server.server.version
}

resource "teamcity_cleanup" "cleanup" {
  enabled = true
  max_duration = 0

  daily = {
    hour = 2
    minute = 15
  }

#  cron = {
#    minute = 15
#    hour = 2
#    day = 2
#    month = "*"
#    day_week = "?"
#  }
}
