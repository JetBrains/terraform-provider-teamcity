variable "teamcity_url" {
  type = string
}

variable "teamcity_token" {
  type      = string
  sensitive = true
}

variable "github_client_secret" {
  type      = string
  sensitive = true
}

variable "github_private_key" {
  type      = string
  sensitive = true
}

variable "github_webhook_secret" {
  type      = string
  sensitive = true
}
