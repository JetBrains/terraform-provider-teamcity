terraform {
  required_providers {
    teamcity = {
      source = "jetbrains/teamcity"
    }
  }
}

//you dont need provider configuration if you use GoLand run configuration "DebugProvider" with env vars

resource "teamcity_pool" "testing" {
  name  = "Test pool"
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

output "test_output" {
  value = teamcity_pool.unlimited
}
