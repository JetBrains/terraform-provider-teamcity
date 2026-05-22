# teamcity_build_configuration_artifact_dependency (Resource)

An artifact dependency in a TeamCity build configuration.

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

resource "teamcity_build_configuration_artifact_dependency" "test" {
  build_configuration_id = teamcity_build_configuration.downstream.id
  depends_on_id          = teamcity_build_configuration.upstream.id
  properties = {
    "pathRules"      = "*.jar => lib"
    "revisionName"   = "lastSuccessful"
    "revisionValue"  = "latest.lastSuccessful"
    "cleanDestinationDirectory" = "true"
  }
}
```

## Schema

### Required

- **build_configuration_id** (String) ID of the build configuration to which this dependency belongs.
- **depends_on_id** (String) ID of the build configuration on which this one depends.

### Optional

- **properties** (Map of String) Properties for the artifact dependency.
  - `pathRules`: Rules to select artifacts and their destination.
  - `revisionName`: `lastSuccessful`, `lastPinned`, `lastFinished`, `buildTag`, `buildNumber`.
  - `revisionValue`: Value for the chosen revision name.
  - `cleanDestinationDirectory`: `true`, `false`.

### Computed

- **id** (String) Resource identifier.
