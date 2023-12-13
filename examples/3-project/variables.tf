variable "teamcity_url" {
  type = string
}

variable "teamcity_token" {
  type      = string
  sensitive = true
}

variable "github_ssh_private_key" {
  type      = string
  sensitive = true
}
