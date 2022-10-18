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

data "teamcity_server" "server" {}

output "server_version" {
  value = data.teamcity_server.server.version
}

resource "teamcity_cleanup" "cleanup" {
  enabled = true
  max_duration = 0
}
