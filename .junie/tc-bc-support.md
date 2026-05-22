# TeamCity Build Configurations Support Estimation & Phased Implementation Plan

## Executive Summary
Supporting TeamCity build configurations in our Terraform provider requires mapping a complex set of nested entities exposed by the TeamCity REST API (/app/rest/buildTypes). A build configuration is not just a single entity; it is composed of settings, parameters, VCS attachments, build steps, features, triggers, and dependencies.

To ensure a smooth, stable, and iterative delivery, the implementation should be split into **7 specific phases**. This approach introduces approximately **9 new Terraform resources** alongside their respective client/ and models/ bindings.

---

## Phased Implementation Plan

### Phase 1: Basic Build Configuration (teamcity_build_configuration)
**Goal:** Create the basic shell of a build configuration.
**Description:** Implement the core Terraform resource for creating, reading, updating, and deleting a build configuration.
- **REST API Endpoints:** POST /app/rest/buildTypes, GET /app/rest/buildTypes/{id}
- **Attributes:** id, name, project_id, description, build_type (regular, composite, deployment), and paused.
- **Additional Work:** Update the existing teamcity_build_configuration data source to use the new models and client/ helpers, aligning it with current guidelines.
- **Complexity:** Low. Follows the standard New(), Get(), Update(), Delete() abstraction.

### Phase 2: Build Configuration Parameters (teamcity_build_configuration_parameter)
**Goal:** Allow users to set system, environment, and configuration parameters.
**Description:** Build configurations have their own parameter context. This resource will be nearly identical in structure to the existing teamcity_project_parameter resource.
- **REST API Endpoints:** /app/rest/buildTypes/{id}/parameters
- **Attributes:** build_configuration_id, name, value, type.
- **Complexity:** Low. Can reuse logic/patterns from project parameters.

### Phase 3: General Settings & VCS Root Attachments
**Goal:** Configure basic execution settings and attach source code repositories.
**Description:**
1. **teamcity_build_configuration_settings**: Manage properties like buildNumberCounter, buildNumberPattern, and artifactRules via /app/rest/buildTypes/{id}/settings.
2. **teamcity_build_configuration_vcs_root**: Attach existing VCS roots to the build configuration.
- **REST API Endpoints:** /app/rest/buildTypes/{id}/vcs-root-entries
- **Attributes:** build_configuration_id, vcs_root_id, checkout_rules.
- **Complexity:** Medium. Requires careful handling of checkout rules strings.

### Phase 4: Build Steps (teamcity_build_configuration_step)
**Goal:** Define the actual execution steps (e.g., Maven, CommandLine, Docker).
**Description:** Because TeamCity relies heavily on plugins for runners, creating a strictly typed resource for every runner is unscalable. A generic approach using a properties map is recommended.
- **REST API Endpoints:** /app/rest/buildTypes/{id}/steps
- **Attributes:** build_configuration_id, type (e.g., simpleRunner, Maven2), properties (map of strings).
- **Complexity:** Medium. Will require good documentation on how to map TeamCity UI step properties to the map format.

### Phase 5: Build Features (teamcity_build_configuration_feature)
**Goal:** Enable features like Free Disk Space, Swabra, or XML Report Processing.
**Description:** Similar to build steps, features should be implemented as a generic resource taking a type and a map of properties.
- **REST API Endpoints:** /app/rest/buildTypes/{id}/features
- **Attributes:** build_configuration_id, type, properties (map of strings).
- **Complexity:** Medium. Follows the exact same pattern as Build Steps.

### Phase 6: Build Triggers (teamcity_build_configuration_trigger)
**Goal:** Automatically start builds based on VCS changes, schedules, or other events.
**Description:** A generic resource for triggers.
- **REST API Endpoints:** /app/rest/buildTypes/{id}/triggers
- **Attributes:** build_configuration_id, type (e.g., vcsTrigger, schedulingTrigger), properties (map of strings).
- **Complexity:** Medium. Follows the exact same pattern as Build Steps and Features.

### Phase 7: Dependencies & Agent Requirements
**Goal:** Achieve full feature parity by supporting build chains and agent routing.
**Description:** Implement three specific resources for advanced configuration:
1. **teamcity_build_configuration_snapshot_dependency**: Define pipeline dependencies (depends_on, run-on-same-agent options).
2. **teamcity_build_configuration_artifact_dependency**: Define artifact sharing rules (depends_on, artifact_rules, revision_rule).
3. **teamcity_build_configuration_agent_requirement**: Define agent compatibility conditions (parameter_name, condition, value).
- **REST API Endpoints:** /app/rest/buildTypes/{id}/snapshot-dependencies, /artifact-dependencies, /agent-requirements.
- **Complexity:** High. Requires robust state management, especially when dealing with complex artifact dependency rules.

---

## Conclusion
By splitting the build configuration support into 7 iterative phases, we can deliver value immediately starting from basic configuration provisioning. Using generic properties maps for Steps, Features, and Triggers will drastically reduce the maintenance burden while providing 100% parity with TeamCity's underlying flexibility.