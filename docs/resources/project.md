---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "teamcity_project Resource - terraform-provider-teamcity"
subcategory: ""
description: |-
  A project in TeamCity is a collection of build configurations. More info here https://www.jetbrains.com/help/teamcity/project.html
---

# teamcity_project (Resource)

A project in TeamCity is a collection of build configurations. More info [here](https://www.jetbrains.com/help/teamcity/project.html)

## Example Usage

```terraform
resource "teamcity_project" "provider" {
  name = "Project 1"
}
```

## Schema

### Required

- `name` (String)

### Optional

- `id` (String) Project ID. Autogenerated by default.

## Import

```terraform
import {
  to = teamcity_project.project1
  id = "Project1"
}
```
