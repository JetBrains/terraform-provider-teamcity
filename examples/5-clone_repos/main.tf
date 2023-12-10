terraform {
  required_providers {
    github   = { source = "integrations/github" }
    teamcity = { source = "jetbrains/teamcity" }
  }
}

provider "github" {
  owner = var.github_organization
  token = var.github_token
}

provider "teamcity" {
  host  = var.teamcity_url
  token = var.teamcity_token
}

module "project5" {
  source = "./project"

  organization = var.github_organization
  repo_name = "demo5"
  teamcity_project_name = "Demo 5"
}
module "project6" {
  source = "./project"

  organization = var.github_organization
  repo_name = "demo6"
  teamcity_project_name = "Demo 6"
}
