variable "teamcity_url" {
  type = string
  default = "http://localhost:8111"
}

variable "teamcity_token" {
  type      = string
  sensitive = true
}
