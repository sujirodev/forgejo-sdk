# 4. Stop Generating Quota Models from Swagger

Date: 2026-09-18

## Status

Accepted

## Context

ADR 0003 decided to isolate go-swagger-generated code in
`forgejo/internal/generated/models`, with handwritten public models in
`forgejo/models` and an explicit mapping layer translating between the two.
Only the Quota area (`forgejo/user_quota.go`) ever adopted this pattern.

In practice, after the mapping layer was written, the two sets of structs
ended up structurally identical: same fields, same JSON tags, no divergent
validation or transformation logic. The generated models added no behavior
that the public models didn't already have; they only added:

- ~230 lines of generated boilerplate (`Validate`, `ContextValidate`,
  `MarshalBinary`) that is never called anywhere in the SDK, since the SDK
  never validates responses it decodes.
- Three extra direct dependencies (`go-openapi/errors`, `go-openapi/strfmt`,
  `go-openapi/swag`) and roughly a dozen further indirect ones
  (`go-viper/mapstructure`, `google/uuid`, `oklog/ulid`, `golang.org/x/net`,
  `golang.org/x/text`, the `swag/*` submodules, etc.), all pulled into the
  `go.mod` of anyone who imports the SDK, purely to support a code
  generation step that runs at development time, not at runtime.
- ~130 lines of one-to-one mapping functions (`mapQuotaInfo`,
  `mapQuotaUsedSize`, ...) that just copy fields across two structs with the
  same shape.
- A `swagger.v1.json` file that has to be downloaded separately (it's not
  vendored, per `.gitignore`) to regenerate anything, and a `make swagger`
  target that depends on installing `go-swagger` v0.33.1.

## Decision

Drop the generated layer for Quota. `forgejo/user_quota.go` now decodes API
responses directly into the public `forgejo/models` structs with
`encoding/json`, the same way every other module in the SDK does.

- Delete `forgejo/internal/generated/models` entirely.
- Delete the `mapQuota*` functions; `forgejo/models/quota.go` keeps the same
  field names, types and JSON tags it already had.
- Remove the `swagger`, `swagger-install` and `swagger-generate-models`
  Makefile targets and the `go-swagger` dependency they installed.
- Drop `go-openapi/errors`, `go-openapi/strfmt`, `go-openapi/swag` and their
  transitive dependencies from `go.mod` (`go mod tidy`).

This does not reverse the *intent* of ADR 0003 — models handwritten in
`forgejo/models`, decoupled from any generator's own types, stay the
approach. What's dropped is the generator itself: for a spec this small and
this stable, a generation step earns its cost only if the generated and
public shapes are expected to diverge. They didn't.

If a future area of the SDK has response payloads too large or too
volatile to hand-write comfortably, go-swagger generation can be
reintroduced for that area specifically — this ADR only retires it for
Quota, and only because it turned out to be pure overhead there.

## Consequences

- Smaller dependency footprint for every consumer of the SDK: three fewer
  direct `go.mod` requires, about a dozen fewer indirect ones.
- Less code to maintain: no generated files, no mapping functions, no
  `swagger.v1.json` download step, no pinned `go-swagger` version.
- One thing is lost: if the Quota response schema evolves in a way
  go-swagger's validation would have caught (e.g. a field's type changing
  incompatibly), we'd now find out from a failed `json.Unmarshal` or a wrong
  zero value instead of a generated `Validate()` call. Given the SDK never
  invoked that validation, the practical loss is close to zero.
- Contributors adding a new Quota field now edit `forgejo/models/quota.go`
  by hand and add the field to the request path in `user_quota.go` directly,
  instead of regenerating and re-mapping.
