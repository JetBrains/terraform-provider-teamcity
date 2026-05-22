# teamcity_build_configuration_step (Resource)

A build step in a TeamCity build configuration.

## Example Usage

```terraform
resource "teamcity_project" "test" {
  name = "Test Project"
}

resource "teamcity_build_configuration" "test" {
  name       = "Test Build Conf"
  project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration_step" "test" {
  build_configuration_id = teamcity_build_configuration.test.id
  name                   = "Run script"
  type                   = "simpleRunner"
  properties = {
    "script.content"    = "echo Hello World"
    "use.custom.script" = "true"
  }
}
```

## Schema

### Required

- **build_configuration_id** (String) ID of the build configuration to which this step belongs.
- **type** (String) The type of the build runner (e.g., `simpleRunner`, `Maven2`, `Ant`, `docker.runner`).

### Optional

- **name** (String) Name of the build step.
- **properties** (Map of String) Properties for the build runner. These correspond to the settings available for the specific runner in the TeamCity UI.

### Computed

- **id** (String) Resource identifier (Step ID).
