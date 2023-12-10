variable "project_name" {
  type = string
}

variable "github_ssh_private_key" {
  type      = string
  sensitive = true
}

variable "repo_url" {
  type      = string
}
