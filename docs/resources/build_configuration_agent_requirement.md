# teamcity_build_configuration_agent_requirement (Resource)

An agent requirement in a TeamCity build configuration.

## Example Usage

```terraform
resource "teamcity_project" "test" {
  name = "Test Project"
}

resource "teamcity_build_configuration" "test" {
  name       = "Test Build Conf"
  project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration_agent_requirement" "os_linux" {
  build_configuration_id = teamcity_build_configuration.test.id
  condition              = "equals"
  name                   = "os.name"
  value                  = "Linux"
}

resource "teamcity_build_configuration_agent_requirement" "docker_exists" {
  build_configuration_id = teamcity_build_configuration.test.id
  condition              = "exists"
  name                   = "docker.server.version"
}
```

## Schema

### Required

- **build_configuration_id** (String) ID of the build configuration to which this requirement belongs.
- **condition** (String) The condition of the agent requirement (e.g., `equals`, `exists`, `contains`, `matches`, `more-than`, `less-than`, `not-more-than`, `not-less-than`, `starts-with`, `ends-with`, `not-contains`, `ver-more-than`, `ver-not-more-than`, `ver-less-than`, `ver-not-less-than`).
- **name** (String) The name of the agent parameter to check.

### Optional

- **value** (String) The value to compare against (not required for all conditions like `exists`).

### Computed

- **id** (String) Resource identifier (Requirement ID).
