terraform {
  required_providers {
    github   = { source = "integrations/github" }
    teamcity = { source = "jetbrains/teamcity" }
  }
}

data "github_repository" "repo" {
  full_name = var.repo
}
resource "teamcity_project" "project" {
  name = data.github_repository.repo.name
}

resource "teamcity_vcsroot" "root" {
  name       = data.github_repository.repo.name
  project_id = teamcity_project.project.id

  git = {
    url         = data.github_repository.repo.http_clone_url
    branch      = data.github_repository.repo.default_branch
    auth_method = "PASSWORD"
    password    = var.github_token
  }
}

resource "teamcity_versioned_settings" "settings1" {
  project_id       = teamcity_project.project.id
  vcsroot_id       = teamcity_vcsroot.root.id
  settings         = "useFromVCS"
  allow_ui_editing = true
  show_changes     = false
}
