terraform {
  required_providers {
    teamcity = {
      source = "jetbrains/teamcity"
    }
  }
}

provider "teamcity" {
  host  = var.teamcity_url
  token = var.teamcity_token
}

resource "teamcity_pool" "testing" {
  name  = "test"
  size  = 30
}

resource "teamcity_pool" "unlimited" {
  name  = "unlimited"
}

data "teamcity_pool" "default" {
  name = "Default"
}

output "test_default" {
  value = data.teamcity_pool.default
}
