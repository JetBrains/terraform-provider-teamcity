terraform {
  required_providers {
    teamcity = { source = "jetbrains/teamcity" }
  }
}

provider "teamcity" {
  host  = var.teamcity_url
  token = var.teamcity_token
}
