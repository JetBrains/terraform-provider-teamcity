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

resource "teamcity_global_settings" "global" {
  root_url = var.teamcity_url
#  artifact_directories               = "system/artifacts"
#  max_artifact_size                  = 4294967296
#  max_artifact_number                = 10
#  default_execution_timeout          = 10 * 60
#  default_vcs_check_interval         = 100
#  enforce_default_vcs_check_interval = true
#  default_quiet_period               = 50
#
#  encryption = {
#    key = var.encryption_key
#  }
#
#  artifacts_domain_isolation = {
#    artifacts_url = "https://artifacts"
#  }
}

resource "teamcity_auth" "test" {
  allow_guest             = false
  guest_username          = "guest"
  welcome_text            = ""
  collapse_login_form     = true
  per_project_permissions = true
  email_verification      = false

  modules = {
    token = {}

    built_in = {
      registration     = false
      change_passwords = false
    }

    github_app = {
      create_new_users = true
      organizations    = "teamcity-terraform-test"
    }
  }
}

resource "teamcity_connection" "github" {
  project_id = "_Root"
  github_app = {
    display_name   = "teamcity-terraform-test"
    owner_url      = "https://github.com/teamcity-terraform-test"
    app_id         = "391349"
    client_id      = "Iv1.22ae7b78f91c2eb1"
    client_secret  = var.github_client_secret
    private_key    = var.github_private_key
    webhook_secret = var.github_webhook_secret
  }
}

resource "teamcity_email_settings" "email" {
  enabled           = false
  host              = "mail"
  port              = 587
  from              = "TeamCity"
  login             = "teamcity"
  password          = "password"
  secure_connection = "STARTTLS"
}

resource "teamcity_license" "key" {
  key = "12345-67890-12345-67890"
}

resource "teamcity_cleanup" "cleanup" {
  enabled = true
  daily   = {
    hour   = 2
    minute = 0
  }
  max_duration = 0
}
