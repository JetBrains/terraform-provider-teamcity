# teamcity_build_configuration (Resource)

A build configuration is a collection of settings used to start a build and group the sequence of the builds.

## Example Usage

```terraform
resource "teamcity_project" "test" {
  name = "Test Project"
}

resource "teamcity_build_configuration" "test" {
  name       = "Test Build Conf"
  project_id = teamcity_project.test.id
  description = "My test build configuration"
}
```

## Schema

### Required

- **name** (String) Name of the build configuration.
- **project_id** (String) ID of the project where the build configuration will be created. Changing this attribute will replace the build configuration.

### Optional

- **build_type** (String) Type of the build configuration. Possible values: `regular`, `composite`, `deployment`. Default: `regular`.
- **description** (String) Description of the build configuration.
- **id** (String) ID of the build configuration. If not provided, it will be generated from the name.
- **paused** (Bool) Whether the build configuration is paused.

### Computed

- **id** (String) Resource identifier (Build Configuration ID).
