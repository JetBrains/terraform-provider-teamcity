terraform {
  required_providers {
    teamcity = {
      source = "jetbrains/teamcity"
    }
  }
}
provider "teamcity" {
  host = "http://localhost:8111"
  token = "123"
}

data "teamcity_server" "server" {}

output "server_version" {
  value = data.teamcity_server.server
}
