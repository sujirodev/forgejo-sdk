# 2. Modernizing the SDK

Date: 2026-03-04

## Status

Accepted (see the 2026-09-18 amendment below for how "new module" is
resolved in practice)

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

## Amendment (2026-09-18): what "new module" means in practice

This ADR originally said new code should use the new style "starting with
the introduction of Quota related functions", which reads as: any new
*function*, even one added to an existing legacy file, should get `ctx` and
a value `Response`. That is not what happened in practice, and a survey of
every endpoint added to an existing file after this ADR's date (2026-03-04)
found the legacy style used every time except one:

- `DeleteAccessTokenByID`/`DeleteAccessTokenByName` (`user_app.go`,
  2026-03-11) — new style, the one exception, a week after this ADR.
- `ListOrgs` (`org.go`, 2026-06-14) — legacy style.
- `ListOrgLabels`/`GetOrgLabel`/`CreateOrgLabel`/`EditOrgLabel`/`DeleteOrgLabel`
  (`org_label.go`, 2026-06-24) — legacy style.
- `DiffPatchFile` (`repo_file.go`, 2026-06-25) — legacy style.
- The mirror endpoints added to `repo_mirror.go` — legacy style, matching
  `PushMirrors`/`MirrorSync` already in that file and in `repo.go`.

The rule this codebase actually follows, and the one to keep following:

- **A new file/module** (no existing file to be consistent with) uses the
  new style: `ctx context.Context` as the first parameter, `Response`
  returned by value. Quota (`user_quota.go`) is the reference example.
- **An addition to an existing file** matches that file's existing style,
  legacy or new, even after this ADR. Mixing styles within one file (like
  `user_app.go` now has) makes the file harder to read for no benefit to
  users of that specific file.

This isn't a retreat from the ADR's goal — the dual-style SDK it predicted
is exactly what exists today, on purpose, and the migration path is still
"all legacy methods move to the new pattern in a future major version" (see
[ADR 0005](0005-roadmap-to-v4.md)). It only corrects which axis decides the
style for a *new* function: which file it's added to, not merely that the
SDK already has ctx-style code somewhere else.