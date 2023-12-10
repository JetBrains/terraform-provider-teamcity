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

data "github_repositories" "projects" {
  query = "project in:name org:${var.github_organization} template:false"
}

output "repos" {
  value = data.github_repositories.projects.full_names
}

module "project" {
  for_each = toset(data.github_repositories.projects.full_names)
  source = "./project"
  github_token = var.github_token
  repo = each.key
}
