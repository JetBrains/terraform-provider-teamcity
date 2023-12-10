terraform {
  required_providers {
    teamcity = { source = "jetbrains/teamcity" }
  }
}

provider "teamcity" {
  host  = var.teamcity_url
  token = var.teamcity_token
}

resource "teamcity_role" "developer" {
  name     = "Developer"
  included = [
    "PROJECT_DEVELOPER"
  ]
  permissions = [
    "connect_to_agent",
  ]
}

resource "teamcity_project" "demo" {
  name = "Demo 2"
}

resource "teamcity_group" "developers" {
  name  = "Developers"
  roles = [
    {
      id      = teamcity_role.developer.id
      project = teamcity_project.demo.id
    },
  ]
}

resource "teamcity_user" "mkuzmin" {
  username = "mkuzmin"
  github_username = "mkuzmin"
}

resource "teamcity_group_member" "developers_members" {
  group_id = teamcity_group.developers.id
  username = teamcity_user.mkuzmin.username
}
