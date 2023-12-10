terraform {
  required_providers {
    teamcity = { source = "jetbrains/teamcity" }
  }
}

provider "teamcity" {
  host  = var.teamcity_url
  token = var.teamcity_token
}

module "project" {
  source = "./project"
  project_name = "Demo 4"
  github_ssh_private_key = var.github_ssh_private_key
  repo_url = "git@github.com:teamcity-terraform-test/template.git"
}
module "project" {
  source = "./project"
  project_name = "Demo 4a"
  github_ssh_private_key = var.github_ssh_private_key
  repo_url = "git@github.com:teamcity-terraform-test/template.git"
}
module "project" {
  source = "./project"
  project_name = "Demo 4b"
  github_ssh_private_key = var.github_ssh_private_key
  repo_url = "git@github.com:teamcity-terraform-test/template.git"
}
