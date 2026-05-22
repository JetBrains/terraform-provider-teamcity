terraform {
  required_providers {
    teamcity = {
      source  = "jetbrains/teamcity"
      version = "0.0.90"
    }
  }
}

provider "teamcity" {
  host     = var.teamcity_url
  password = var.teamcity_token
}

############################################################
# Supporting resources: project + VCS root
############################################################

resource "teamcity_project" "demo" {
  name = "BC Demo Project"
}

resource "teamcity_vcsroot" "demo_repo" {
  name       = "Demo VCS Root"
  project_id = teamcity_project.demo.id

  git = {
    url         = "https://github.com/JetBrains/teamcity-rest-api-sample.git"
    branch      = "refs/heads/main"
    auth_method = "ANONYMOUS"
  }
}

############################################################
# PHASE 1: teamcity_build_configuration
############################################################

# Regular build configuration
resource "teamcity_build_configuration" "build" {
  name        = "Build"
  project_id  = teamcity_project.demo.id
  description = "Main build configuration"
  build_type  = "regular"
  paused      = false
}

# Composite build configuration (build chain root)
resource "teamcity_build_configuration" "composite" {
  name        = "Release Chain"
  project_id  = teamcity_project.demo.id
  description = "Composite build that aggregates the build chain"
  build_type  = "composite"
}

# Deployment build configuration
resource "teamcity_build_configuration" "deploy" {
  name        = "Deploy to Staging"
  project_id  = teamcity_project.demo.id
  description = "Deployment build configuration"
  build_type  = "deployment"
  paused      = true
}

############################################################
# PHASE 2: teamcity_build_configuration_parameter
############################################################

# Plain text configuration parameter
resource "teamcity_build_configuration_parameter" "version" {
  build_configuration_id = teamcity_build_configuration.build.id
  name                   = "app.version"
  value                  = "1.0.0"
  type                   = "text"
}

# Default-type parameter (omitting `type` should default to "text")
resource "teamcity_build_configuration_parameter" "env_name" {
  build_configuration_id = teamcity_build_configuration.build.id
  name                   = "env.NAME"
  value                  = "staging"
}

# Secure (password) parameter
resource "teamcity_build_configuration_parameter" "api_key" {
  build_configuration_id = teamcity_build_configuration.build.id
  name                   = "secret.api_key"
  value                  = "super-secret-value"
  type                   = "password"
}

############################################################
# PHASE 3a: teamcity_build_configuration_settings
############################################################

resource "teamcity_build_configuration_settings" "build_settings" {
  build_configuration_id = teamcity_build_configuration.build.id
  build_number_counter   = 42
  build_number_pattern   = "1.0.%build.counter%"
  artifact_rules         = "target/*.jar => artifacts"
}

############################################################
# PHASE 3b: teamcity_build_configuration_vcs_root
############################################################

resource "teamcity_build_configuration_vcs_root" "build_vcs" {
  build_configuration_id = teamcity_build_configuration.build.id
  vcs_root_id            = teamcity_vcsroot.demo_repo.id
  checkout_rules         = "+:src=>source"
}

############################################################
# PHASE 4: teamcity_build_configuration_step
############################################################

# Command line / simple runner step
resource "teamcity_build_configuration_step" "cmd_step" {
  build_configuration_id = teamcity_build_configuration.build.id
  name                   = "Print hello"
  type                   = "simpleRunner"
  properties = {
    "script.content"     = "echo Hello from TeamCity!\necho Building version %app.version%"
    "teamcity.step.mode" = "default"
    "use.custom.script"  = "true"
  }
}

# A second step in the same build configuration
resource "teamcity_build_configuration_step" "compile_step" {
  build_configuration_id = teamcity_build_configuration.build.id
  name                   = "Fake compile"
  type                   = "simpleRunner"
  properties = {
    "script.content"          = "echo Compiling..."
    "teamcity.step.mode"      = "default"
    "use.custom.script"       = "true"
  }
}

############################################################
# PHASE 5: teamcity_build_configuration_feature
############################################################

# Free Disk Space build feature
resource "teamcity_build_configuration_feature" "free_disk" {
  build_configuration_id = teamcity_build_configuration.build.id
  type                   = "jetbrains.agent.free.space"
  properties = {
    "free-space-work" = "3gb"
  }
}

# Performance monitor feature (no required properties)
resource "teamcity_build_configuration_feature" "perfmon" {
  build_configuration_id = teamcity_build_configuration.build.id
  type                   = "perfmon"
  properties             = {}
}

############################################################
# PHASE 6: teamcity_build_configuration_trigger
############################################################

# VCS trigger - run on every VCS change
resource "teamcity_build_configuration_trigger" "vcs_trigger" {
  build_configuration_id = teamcity_build_configuration.build.id
  type                   = "vcsTrigger"
  properties = {
    "quietPeriodMode"     = "DO_NOT_USE"
    "branchFilter"        = "+:*"
    "groupCheckkinsByCommitter" = "true"
  }
}

# Schedule trigger - daily at midnight
resource "teamcity_build_configuration_trigger" "schedule_trigger" {
  build_configuration_id = teamcity_build_configuration.build.id
  type                   = "schedulingTrigger"
  properties = {
    "schedulingPolicy"    = "daily"
    "hour"                = "0"
    "minute"              = "0"
    "timezone"            = "SERVER"
    "triggerBuildWithPendingChangesOnly" = "true"
  }
}

############################################################
# PHASE 7a: teamcity_build_configuration_snapshot_dependency
############################################################

# Composite depends on the build (build chain)
resource "teamcity_build_configuration_snapshot_dependency" "chain" {
  build_configuration_id = teamcity_build_configuration.composite.id
  depends_on_id          = teamcity_build_configuration.build.id
  properties = {
    "run-build-if-dependency-failed"     = "MAKE_FAILED_TO_START"
    "run-build-if-dependency-failed-to-start" = "MAKE_FAILED_TO_START"
    "run-build-on-the-same-agent"        = "false"
    "take-started-build-with-same-revisions" = "true"
    "take-successful-builds-only"        = "true"
  }
}

############################################################
# PHASE 7b: teamcity_build_configuration_artifact_dependency
############################################################

# Deploy uses artifacts from build
resource "teamcity_build_configuration_artifact_dependency" "deploy_artifact" {
  build_configuration_id = teamcity_build_configuration.deploy.id
  depends_on_id          = teamcity_build_configuration.build.id
  properties = {
    "pathRules"           = "artifacts/**=>downloaded"
    "revisionName"        = "lastSuccessful"
    "revisionValue"       = "latest.lastSuccessful"
    "cleanDestinationDirectory" = "true"
  }
}

############################################################
# PHASE 7c: teamcity_build_configuration_agent_requirement
############################################################

# Equals condition with value
resource "teamcity_build_configuration_agent_requirement" "os_req" {
  build_configuration_id = teamcity_build_configuration.build.id
  condition              = "equals"
  name                   = "teamcity.agent.jvm.os.name"
  value                  = "Linux"
}

# 'exists' condition with no value
resource "teamcity_build_configuration_agent_requirement" "docker_req" {
  build_configuration_id = teamcity_build_configuration.build.id
  condition              = "exists"
  name                   = "docker.version"
}

############################################################
# Data source: teamcity_build_configuration
############################################################

data "teamcity_build_configuration" "build_lookup" {
  id = teamcity_build_configuration.build.id

  depends_on = [
    teamcity_build_configuration_parameter.version,
    teamcity_build_configuration_settings.build_settings,
    teamcity_build_configuration_step.cmd_step,
  ]
}

############################################################
# Outputs - useful for verification
############################################################

output "build_id" {
  value = teamcity_build_configuration.build.id
}

output "composite_id" {
  value = teamcity_build_configuration.composite.id
}

output "deploy_id" {
  value = teamcity_build_configuration.deploy.id
}

output "data_lookup" {
  value = data.teamcity_build_configuration.build_lookup
}

output "param_ids" {
  value = {
    text     = teamcity_build_configuration_parameter.version.id
    default  = teamcity_build_configuration_parameter.env_name.id
    password = teamcity_build_configuration_parameter.api_key.id
  }
}

output "vcs_attachment_id" {
  value = teamcity_build_configuration_vcs_root.build_vcs.id
}

output "step_ids" {
  value = [
    teamcity_build_configuration_step.cmd_step.id,
    teamcity_build_configuration_step.compile_step.id,
  ]
}

output "feature_ids" {
  value = [
    teamcity_build_configuration_feature.free_disk.id,
    teamcity_build_configuration_feature.perfmon.id,
  ]
}

output "trigger_ids" {
  value = [
    teamcity_build_configuration_trigger.vcs_trigger.id,
    teamcity_build_configuration_trigger.schedule_trigger.id,
  ]
}

output "snapshot_dep_id" {
  value = teamcity_build_configuration_snapshot_dependency.chain.id
}

output "artifact_dep_id" {
  value = teamcity_build_configuration_artifact_dependency.deploy_artifact.id
}

output "agent_req_ids" {
  value = [
    teamcity_build_configuration_agent_requirement.os_req.id,
    teamcity_build_configuration_agent_requirement.docker_req.id,
  ]
}
