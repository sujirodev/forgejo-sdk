# 2. Modernizing the SDK

Date: 2026-03-04

## Status

Proposed

## Context

The existing SDK architecture was designed around older Go patterns (pre-context
and pointer-heavy returns). Specifically, the Client maintains a single internal
context.Context, and most methods return *Response pointers. This leads to several
sub-optimal behaviors:

- Lack of per-call timeout and cancellation control.

- Risk of nil-pointer panics when accessing response metadata (e.g., resp.StatusCode)
  on failed requests.

- Increased heap allocations and GC pressure due to pointer returns for small
  metadata structs.

- Inconsistency with modern Go (1.21+) idioms where context.Context is explicitly
  passed and value semantics are preferred for resource handles.
    
As a practical example, old style functions would be along the lines of:

```go
// CheckMyQuota checks if the authenticated user is over quota.
func (c *Client) CheckMyQuota(subject QuotaSubject) (bool, *Response, error) { ... }
```

## Decision

A newer style will be used for the SDK starting with the introduction of Quota related functions.
This will essentially mean:

- Pass context.Context as the first argument to every method.

- Return by value (Response instead of *Response for example)

A newer style will be used for the SDK starting with the introduction of Quota-related
functions. This will serve as the blueprint for future expansion and incremental
refactoring of existing modules.

The core changes include:

- Explicit Context: Every public method must accept context.Context as its first
  argument. Internal helpers (e.g., doRequest) will be updated to doRequestWithContext
  to propagate this context to the underlying http.Request.

- Value Returns: Methods will return Response by value instead of *Response. This
  ensures that the "Zero Value" of a response is safe to query (e.g., resp.StatusCode()
  returns 0 instead of panicking).

## Consequences

- Improved Robustness: Callers can safely check response metadata without defensive
  nil-checks, reducing the surface area for runtime crashes.

- Granular Control: Users can now use context.WithTimeout or context.WithCancel
  on a per-request basis, which is critical for high-availability applications.

- Performance: Returning small structs by value allows the Go compiler to utilize
  stack allocation more effectively, reducing garbage collection overhead.

- Incremental Breaking Changes: While new modules (Quota) will follow this pattern,
  existing modules will remain on the legacy pattern to maintain semver compatibility.
  This creates a temporary "dual-style" SDK, which will be resolved in a future
  major version by migrating all legacy methods to the new pattern.

- Testability: Tests can now utilize t.Context() to ensure network resources are
  cleaned up immediately when a test times out or fails.

Example of new style:

```go
// CheckMyQuota checks if the authenticated user is over quota.
func (c *Client) CheckMyQuota(ctx context.Context, subject QuotaSubject) (bool, Response, error) { ... }
```