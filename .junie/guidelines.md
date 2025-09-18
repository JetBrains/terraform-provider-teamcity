# TeamCity Terraform Provider – Project Guidelines

This document describes how to contribute code to this repository and align with the current architecture. It focuses on how to abstract raw HTTP calls into TeamCity resources, how to use the new HTTP client helpers, how to organize models.

## Overview
- `client/`: Go HTTP client for TeamCity REST API. Each file in this package focuses on a specific TeamCity entity (project, pool, vcsroot, etc.).
- `models/`: Shared models with two purposes:
  - DataModel – Terraform SDK structures
  - Json – JSON payload structures used by the HTTP client
- `teamcity/`: Terraform provider implementation (resources and data sources) built on top of `client/` and `models/`.
- You can use [teamcity/pool_resource.go](../teamcity/pool_resource.go) as, currently, the latest example of how new Terraform resources should be built.

## 1. HTTP Client Layer (client/)
- [client/http_client.go](../client/http_client.go) contains the canonical HTTP layer. New code should use the new helper methods:
  - GetRequest / GetRequestWithContext
  - PostRequest / PostRequestWithContext
  - PutRequest / PutRequestWithContext
  - DeleteRequest / DeleteRequestWithContext
- The older methods are deprecated and should not be used:
  - doRequest, doRequestWithType, request, requestWithType
- Error handling: use `errors.Is(err, client.ErrNotFound)` to handle 404-like conditions instead of checking HTTP status codes each time in the caller.
- Request/response bodies:
  - Marshal models.Json using encoding/json.
  - Pass bytes.NewReader(rb) as the body to PostRequest/PutRequest.
  - Provide a pointer for the response struct; the helper fills it when applicable.

Example (see [client/pool.go](../client/pool.go)):

```
func (c *Client) NewPool(p models.PoolJson) (*models.PoolJson, error) {
    var actual models.PoolJson
    rb, err := json.Marshal(p)
    if err != nil {
        return nil, err
    }
    if err := c.PostRequest("/agentPools", bytes.NewReader(rb), &actual); err != nil {
        return nil, err
    }
    return &actual, nil
}
```

### If operation requires retries
- There is `retryableRequest` with `retryPolicy`, this provider has a parameter for configuring retries: `max_retries`, each retry is 5 sec long, 12 by default.
- Example where retries are needed - setting properties on versioned settings after main configuration is applied `SetVersionedSettingsProperty()` in [project.go](../client/project.go), since
  we need to wait for the feature to be ready. 
- Example of retry policy implementation in [project.go](../client/project.go) - `retryPolicy()`, ignore 500 error from server.

## 2. Terraform Provider (teamcity/)
- The Terraform Plugin Framework is used. The current latest great example is [teamcity/pool_resource.go](../teamcity/pool_resource.go).
- General guidelines for implementing a new resource:
    - Implement resource.Resource with Metadata, Schema, Create, Read, Update, Delete, and Configure methods.
    - In Create/Update, map DataModel → Json and use `client/.go` functions (e.g., New, Update), then map the response back to DataModel and set the state.
    - In Read, call Get; if it returns nil (ErrNotFound), remove the resource from the state.
    - In Delete, call Delete and handle errors appropriately.
    - It is better to use validators, plan modifiers, and attribute conversions similar to [pool_resource.go](../teamcity/pool_resource.go) for robust UX (e.g., int64validator, setplanmodifier, etc.).
- Data sources should follow the same separation of concerns: read using logic from `client/`, map to DataModel, and expose via the Terraform schema.

# Additional information

## Resource-specific HTTP abstractions (client/.go)
- Each TeamCity resource should have a thin, typed abstraction over HTTP.
- Provide the canonical CRUD functions that work with models.Json:
  - New(payload models.Json) (*models.Json, error) → POST
  - Get(locator string) (*models.Json, error) → GET
    - Return (nil, nil) when ErrNotFound
  - Update(locator string, payload models.Json) (*models.Json, error) → POST or PUT
  - Delete(locator or id string) error → DELETE
- Build endpoints using TeamCity locators (e.g., id:<id>, name:<name>) consistently.
- Avoid using deprecated doRequest/request variants for new code.

## Models (models/)
- All new models should live under /models.
- Maintain two layers per resource:
  - DataModel – used by Terraform Plugin Framework types. Mirrors the Terraform schema types (types.String, types.Int64, types.Set, etc.). Example: `models.PoolDataModel`.
  - Json – the TeamCity API JSON representation used by the client. Example: `models.PoolJson`.
- Provide clear field names and json tags that match TeamCity REST API payloads.
- Keep conversion logic between DataModel and Json simple and explicit. Helper functions are welcome if they improve clarity.


## Deprecations
- doRequest/doRequestWithType/request/requestWithType are deprecated for new development.
- All new code should use GetRequest/PostRequest/PutRequest/DeleteRequest (and their WithContext variants).
- When refactoring old code, migrate to the new helpers opportunistically. New PRs should not introduce fresh usages of deprecated functions.

## Error Handling Guidelines
- Prefer returning (nil, nil) for not-found reads in client layer by detecting errors.Is(err, client.ErrNotFound).
- Surface descriptive errors from resource layer using resp.Diagnostics.AddError with actionable messages.
- Ensure Create/Update state population is complete and uses values returned by the server (IDs, computed fields, etc.).

## Testing
- Add/maintain unit tests under client/*_test.go for client behavior.
- Add/maintain tests under teamcity/*_test.go for resource behavior when feasible.
- Validate that ErrNotFound flows are handled correctly in both client and resource layers.

## Style and Naming
- File naming: teamcity/_resource.go for resources; teamcity/_data_source.go for data sources; client/.go for HTTP abstractions; models/.go for models.
- Types: DataModel and Json.
- Functions: New, Get, Update, Delete in client package. Use clear parameter names that hint at locator usage (id, name, locator).
