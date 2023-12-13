terraform {
  required_providers {
    github   = { source = "integrations/github" }
    tls      = { source = "hashicorp/tls" }
    local    = { source = "hashicorp/local" }
    teamcity = { source = "jetbrains/teamcity" }
  }
}

resource "github_repository" "test1" {
  name       = var.repo_name
  visibility = "private"

  template {
    owner      = var.organization
    repository = "template"
  }
}

resource "tls_private_key" "key1" {
  algorithm = "ED25519"
}

resource "github_repository_deploy_key" "key1" {
  repository = github_repository.test1.name
  title      = "teamcity"
  key        = tls_private_key.key1.public_key_openssh
  read_only  = false
}

////////////////////

resource "teamcity_project" "project" {
  name = var.teamcity_project_name
}

resource "teamcity_ssh_key" "key" {
  project_id  = teamcity_project.project.id
  name        = "github"
  private_key = tls_private_key.key1.private_key_openssh
}

resource "teamcity_vcsroot" "root" {
  name       = github_repository.test1.full_name
  project_id = teamcity_project.project.id

  git = {
    url          = github_repository.test1.ssh_clone_url
    branch       = github_repository.test1.default_branch
    auth_method  = "TEAMCITY_SSH_KEY"
    uploaded_key = teamcity_ssh_key.key.name
  }
}

resource "teamcity_versioned_settings" "settings1" {
  project_id       = teamcity_project.project.id
  vcsroot_id       = teamcity_vcsroot.root.id
  settings         = "useFromVCS"
  allow_ui_editing = true
  show_changes     = false
}
