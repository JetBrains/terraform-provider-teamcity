# teamcity_build_configuration_settings (Resource)

General settings for a build configuration, including build number counter, pattern, and artifact rules.

~> **Note:** Deleting this resource will not delete the settings from TeamCity, but will reset them to their default values (e.g., build counter to 1, pattern to `%build.counter%`, and empty artifact rules).

## Example Usage

```terraform
resource "teamcity_project" "example" {
  name = "Example Project"
}

resource "teamcity_build_configuration" "example" {
  project_id = teamcity_project.example.id
  name       = "Example Build Configuration"
}

resource "teamcity_build_configuration_settings" "example" {
  build_configuration_id = teamcity_build_configuration.example.id
  build_number_counter   = 100
  build_number_pattern   = "v%build.counter%"
  artifact_rules         = "+:target/*.jar"
}
```

## Schema

### Required

- `build_configuration_id` (String) The ID of the build configuration.

### Optional

- `artifact_rules` (String) Rules for artifacts produced by the build.
- `build_number_counter` (Number) The next build number to be used.
- `build_number_pattern` (String) The pattern for the build number.

### Computed

- `id` (String) Resource identifier (same as build_configuration_id).
