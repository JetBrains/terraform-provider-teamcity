terraform {
  required_providers {
    teamcity = {
      source = "jetbrains/teamcity"
    }
  }
}

variable "aws_access_key_id" {
  type      = string
  sensitive = true
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}

variable "aws_ami_id" {
  type = string
}

resource "teamcity_project" "cloud_agents" {
  id   = "CloudAgents"
  name = "Cloud Agents"
}

resource "teamcity_cloud_profile" "aws" {
  name              = "AWS EC2 Build Agents"
  cloud_provider_id = "amazon"
  project_id        = teamcity_project.cloud_agents.id

  properties = {
    "secure:access-id"  = var.aws_access_key_id
    "secure:secret-key" = var.aws_secret_access_key
    "region"            = "eu-west-1"
  }

  image {
    name = "Ubuntu build agent"
    properties = {
      "amazon-id"     = var.aws_ami_id
      "instance-type" = "t3.medium"
    }
  }
}
