terraform {
  required_providers {
    teamcity = { source = "jetbrains/teamcity" }
  }
}

provider "teamcity" {
  host  = var.teamcity_url
  token = var.teamcity_token
}

resource "teamcity_project" "demo3" {
  name = "Demo 3"
}

resource "teamcity_ssh_key" "demo3" {
  project_id  = teamcity_project.demo3.id
  name        = "github"
  private_key = var.github_ssh_private_key
}

resource "teamcity_vcsroot" "template" {
  name       = "template"
  project_id = teamcity_project.demo3.id

  git = {
    url          = "git@github.com:teamcity-terraform-test/template.git"
    branch       = "main"
    auth_method  = "TEAMCITY_SSH_KEY"
    uploaded_key = teamcity_ssh_key.demo3.name
  }
}

resource "teamcity_versioned_settings" "demo3" {
  project_id       = teamcity_project.demo3.id
  vcsroot_id       = teamcity_vcsroot.template.id
  settings         = "useFromVCS"
  allow_ui_editing = true
  show_changes     = false

  provisioner "local-exec" {
    command     = "./healthcheck.sh '${var.teamcity_url}/app/rest/buildTypes/id:${teamcity_project.demo3.id}_Build'"
    environment = {
      TIMEOUT = 10*60 # 10 minutes
      TOKEN   = var.teamcity_token
    }
  }
}
