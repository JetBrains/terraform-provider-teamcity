variable "teamcity_url" {
  type = string
}

variable "teamcity_token" {
  type      = string
  sensitive = true
}

variable "github_organization" {
  type = string
}

variable "github_token" {
  type = string
  sensitive = true
}
