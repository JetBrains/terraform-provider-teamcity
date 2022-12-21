terraform {
  required_providers {
    teamcity = {
      source = "jetbrains/teamcity"
    }
  }
}
provider "teamcity" {
  host = "http://localhost:8111"
#  token = "" # env.TEAMCITY_TOKEN
}

data "teamcity_server" "buildserver" {}

output "version" {
  value = data.teamcity_server.buildserver.version
}

resource "teamcity_cleanup" "cleanup" {
  enabled = true
  max_duration = 0

  daily = {
    hour = 2
    minute = 15
  }
#
#  cron = {
#    minute = 15
#    hour = 2
#    day = 2
#    month = "*"
#    day_week = "?"
#  }
}

#resource "teamcity_project" "provider" {
#  name = "Terraform Provider2"
#}

#resource "teamcity_vcsroot" "site" {
#  name = "site5"
#  type = "jetbrains.git"
#
#  #  project_id = "TerraformProvider"
#  project_id = "_Root"
#
##  polling_interval = 100
#
#  git = {
#    url    = "git@github.com:mkuzmin/test.git"
##    url    = "https://github.com/mkuzmin/mkuzmin.github.io.git"
#    branch = "main"
##    push_url = "https://github.com/mkuzmin/mkuzmin.github.io.git"
##    branch_spec = "+:*"
##    tags_as_branches = false
##    username_style = "NAME"
##    submodules = "CHECKOUT"
##    username_for_tags = "aaa"
##    ignore_known_hosts = true
##    path_to_git = "/qwe"
##    convert_crlf = true
##    checkout_policy = "SHALLOW_CLONE"
##    clean_policy = "NEVER"
##    clean_files_policy = "IGNORED_ONLY"
#    auth_method = "PRIVATE_KEY_FILE"
##    auth_method = "PASSWORD"
#    username = "aaa"
##    password = "aaa"
##    uploaded_key = "gitkey"
#
#    auth_method = "ANONYMOUS" //default
#
#    auth_method = "PASSWORD"
#    username = "aaa"
#    password = "123"
#
#    auth_method = "TEAMCITY_SSH_KEY"
#    username = "git"
#    uploaded_key = "key1"
#
#    auth_method = "PRIVATE_KEY_DEFAULT"
#    username = "aaa"
#
#    auth_method = "PRIVATE_KEY_FILE"
#    username = "git"
#    private_key_path = "/etc/abc"
#    passphrase = "123"
#
#    auth_anonymous = {}
#    auth_password = {
#      username = "aaa"
#      password = "123"
#    }
#    auth_uploaded_key = {}
#    auth_default_key = {}
#    auth_custom_key = {}
#
#  }
#}
