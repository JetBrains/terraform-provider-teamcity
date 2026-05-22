# teamcity_build_configuration_trigger (Resource)

A build trigger in a TeamCity build configuration.

## Example Usage

```terraform
resource "teamcity_project" "test" {
  name = "Test Project"
}

resource "teamcity_build_configuration" "test" {
  name       = "Test Build Conf"
  project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration_trigger" "vcs" {
  build_configuration_id = teamcity_build_configuration.test.id
  type                   = "vcsTrigger"
  properties = {
    "quietPeriodMode" = "DO_NOT_USE"
  }
}
```

## Schema

### Required

- **build_configuration_id** (String) ID of the build configuration to which this trigger belongs.
- **type** (String) The type of the build trigger (e.g., `vcsTrigger`, `schedulingTrigger`).

### Optional

- **properties** (Map of String) Properties for the build trigger. These correspond to the settings available for the specific trigger in the TeamCity UI.

### Computed

- **id** (String) Resource identifier (Trigger ID).
