terraform {
  required_providers {
    teamcity = { source  = "jetbrains/teamcity" }
  }
}

resource "teamcity_project" "project" {
  name = var.project_name
}

resource "teamcity_ssh_key" "key" {
  project_id  = teamcity_project.project.id
  name        = "github"
  private_key = var.github_ssh_private_key
}

resource "teamcity_vcsroot" "root" {
  name       = "template"
  project_id = teamcity_project.project.id

  git = {
    url          = var.repo_url
    branch       = "main"
    auth_method  = "TEAMCITY_SSH_KEY"
    uploaded_key = teamcity_ssh_key.key.name
  }
}

resource "teamcity_versioned_settings" "settings" {
  project_id       = teamcity_project.project.id
  vcsroot_id       = teamcity_vcsroot.root.id
  settings         = "useFromVCS"
  allow_ui_editing = true
  show_changes     = false
}
