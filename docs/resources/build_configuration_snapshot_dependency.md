# teamcity_build_configuration_snapshot_dependency (Resource)

A snapshot dependency in a TeamCity build configuration.

## Example Usage

```terraform
resource "teamcity_project" "test" {
  name = "Test Project"
}

resource "teamcity_build_configuration" "upstream" {
  name       = "Upstream"
  project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration" "downstream" {
  name       = "Downstream"
  project_id = teamcity_project.test.id
}

resource "teamcity_build_configuration_snapshot_dependency" "test" {
  build_configuration_id = teamcity_build_configuration.downstream.id
  depends_on_id          = teamcity_build_configuration.upstream.id
  properties = {
    "run-build-on-the-same-agent" = "true"
    "take-successful-builds-only" = "true"
  }
}
```

## Schema

### Required

- **build_configuration_id** (String) ID of the build configuration to which this dependency belongs.
- **depends_on_id** (String) ID of the build configuration on which this one depends.

### Optional

- **properties** (Map of String) Properties for the snapshot dependency.
  - `run-build-if-dependency-failed`: `MAKE_FAILED_TO_START`, `RUN`, `CANCEL`, `MARK_AS_FAILED`
  - `run-build-if-dependency-failed-to-start`: `MAKE_FAILED_TO_START`, `RUN`, `CANCEL`, `MARK_AS_FAILED`
  - `run-build-on-the-same-agent`: `true`, `false`
  - `sync-revisions`: `true`, `false`
  - `take-started-build-with-same-revisions`: `true`, `false`
  - `take-successful-builds-only`: `true`, `false`

### Computed

- **id** (String) Resource identifier.
