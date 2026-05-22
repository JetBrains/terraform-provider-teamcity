# teamcity_build_configuration_vcs_root (Resource)

Attaches a VCS root to a build configuration.

## Example Usage

```terraform
resource "teamcity_project" "example" {
  name = "Example Project"
}

resource "teamcity_vcsroot" "example" {
  name       = "Example VCS Root"
  project_id = teamcity_project.example.id
  git = {
    url    = "https://github.com/example/repo.git"
    branch = "refs/heads/main"
  }
}

resource "teamcity_build_configuration" "example" {
  project_id = teamcity_project.example.id
  name       = "Example Build Configuration"
}

resource "teamcity_build_configuration_vcs_root" "example" {
  build_configuration_id = teamcity_build_configuration.example.id
  vcs_root_id            = teamcity_vcsroot.example.id
  checkout_rules         = "+:.=>somewhere"
}
```

## Schema

### Required

- `build_configuration_id` (String) The ID of the build configuration.
- `vcs_root_id` (String) The ID of the VCS root to attach.

### Optional

- `checkout_rules` (String) Checkout rules for the VCS root.

### Computed

- `id` (String) Resource identifier in the form 'build_configuration_id/vcs_root_id'.
