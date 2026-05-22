# teamcity_build_configuration_feature (Resource)

A build feature in a TeamCity build configuration.

## Example Usage

```terraform
resource "teamcity_project" "test" {
  name = "Test Project"
}

resource "teamcity_build_configuration" "test" {
  name       = "Test Build Conf"
  project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration_feature" "swabra" {
  build_configuration_id = teamcity_build_configuration.test.id
  type                   = "swabra"
  properties = {
    "swabra.enabled" = "true"
    "swabra.strict"  = "true"
  }
}
```

## Schema

### Required

- **build_configuration_id** (String) ID of the build configuration to which this feature belongs.
- **type** (String) The type of the build feature (e.g., `swabra`, `freeDiskSpace`, `xml-report-plugin`).

### Optional

- **properties** (Map of String) Properties for the build feature. These correspond to the settings available for the specific feature in the TeamCity UI.

### Computed

- **id** (String) Resource identifier (Feature ID).
