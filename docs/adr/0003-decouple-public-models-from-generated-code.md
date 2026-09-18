# 3. Decouple Public Models from Generated Code

Date: 2026-03-04

## Status

Superseded by [4. Stop Generating Quota Models from Swagger](0004-stop-generating-quota-models-from-swagger.md)

## Context

The SDK uses / will use go-swagger to generate Go structs from the Forgejo Swagger
specification. While automated generation is efficient for internal implementation,
the resulting code is sub-optimal for a public-facing SDK:

- High Noise-to-Signal Ratio: Generated files are dominated by boilerplate for
  validation (Validate, ContextValidate) and binary marshaling (MarshalBinary),
  obscuring the actual data structure for developers.

- Pointer Overuse: The generator uses pointers for nested structs (e.g., *QuotaUsed)
  to handle nullability, forcing SDK users into defensive nil-checks to avoid panics.

- Leaky Dependencies: Generated models depend on github.com/go-openapi/errors, strfmt,
  and swag. Exposing these directly forces SDK users to inherit these heavy,
  tool-specific dependencies.

- Non-Idiomatic Types: The generator tends to create custom collection types (e.g.,
  QuotaGroupList) instead of standard Go slices ([]QuotaGroup).

- Breaking Changes: Any minor change in the upstream Swagger spec (e.g., renaming
  a field or changing a type) results in a breaking change for the SDK's public API.

## Decision

We will implement a strict isolation layer between the generated Swagger models
and the public SDK models to ensure a clean, idiomatic Go experience.

- Internal Storage: All generated code will be restricted to the internal/generated/models
  package. This package is strictly internal and inaccessible to SDK users.

- Handwritten Public Models: The models package will contain handwritten, idiomatic
  Go structs. These will prioritize value types over pointers to improve ergonomics
  and safety.

- Explicit Mapping: The SDK will implement internal "mapper" functions (e.g., mapQuotaInfo)
  to translate data from the internal generated structs to the public models.

- Zero-Dependency Models: The public models package must remain free of any
  dependencies on the code generation toolchain or go-openapi libraries.

## Consequences

- Superior Ergonomics: SDK users interact with clean, well-documented Go structs
  that behave predictably with zero-values, eliminating "pointer hell."

- API Stability: The SDK can maintain a stable public API even if the underlying
  Forgejo API or Swagger spec changes, by adjusting the internal mapping logic.

- Reduced Dependency Footprint: Users of the SDK do not need to pull in the go-openapi
  stack, keeping their own dependency trees lean.

- Increased Maintenance Effort: Adding new features requires a manual mapping step.
  However, this is considered a worthwhile investment to ensure the SDK feels like
  a first-class Go library rather than a machine-generated wrapper.

- Memory Safety: By converting internal pointers to public value types during mapping,
  we eliminate the risk of nil-pointer panics for the end-user when accessing deeply
  nested data.
