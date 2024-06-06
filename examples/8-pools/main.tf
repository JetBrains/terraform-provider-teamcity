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

resource "teamcity_project" "demo" {
  name  = "Demo Test Project"
}

resource "teamcity_project" "demo2" {
  name  = "Demo Test Project 2"
}

resource "teamcity_pool" "testing" {
  name  = "test"
  size  = 30
  projects = [
        teamcity_project.demo.id,
        teamcity_project.demo2.id
  ]
}

data "teamcity_pool" "default" {
  name = "Default"
}

output "default_pool" {
  value = data.teamcity_pool.default
}
