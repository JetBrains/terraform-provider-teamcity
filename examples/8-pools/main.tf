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

data "teamcity_pool" "test" {
  name = "Default"
}

output "test_output" {
  value = data.teamcity_pool.test
}
