# 5. Roadmap to v4

Date: 2026-09-18

## Status

Proposed

## Context

[ADR 0002](0002-modernizing-the-sdk.md) accepted a permanent dual style
within `forgejo/v3`: new modules use `ctx context.Context` and a
by-value `Response`, existing modules keep their original signatures, and
the two only converge in "a future major version". That's the only
breaking-change bucket this SDK currently has, and nothing tracks what
else should ride in the same major bump. Without a place to write it
down, a compatible-looking PR is the only realistic way anything breaking
ever ships, one method at a time, instead of as one clearly-communicated
v4.

This ADR is a holding area, not a commitment to start the work now. It
gets updated as more breaking changes are identified; when the list is
judged worth shipping, a v4 effort begins from what's written here.

## Decision

Track candidate v4 breaking changes here. As of this writing:

- **Unify every method on the ADR 0002 style.** Every `(T, *Response,
  error)` method gains a leading `ctx context.Context` parameter and
  returns `Response` by value. This is the change ADR 0002 already
  committed to; v4 is where it actually happens across the board instead
  of only in new modules.
- **Remove `NewClientWithHTTP`.** Already marked `// Deprecated: use
  SetHTTPClient option` in `forgejo/client.go`; v4 is when a deprecated,
  already-replaced function actually goes away.
- **Raise the minimum supported Forgejo version to what the README calls
  "minimum supported" (currently 11.0.10)**, retiring the Gitea 1.x-era
  version constants and guards in `forgejo/version.go` that no supported
  server can be older than. Candidate for removal at that point:
  `version1_11_0` through `version1_17_0` and any guard that only exists
  for a server version below the new floor.
- **Module path**: `codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v4`,
  per Go's own major-version-in-import-path convention — not a new
  decision, just what "v4" means mechanically for a Go module.

Anything added to this list before v4 work starts should be a change that
genuinely needs a major version — i.e. it breaks existing callers — not a
staging area for ordinary features. Ordinary new endpoints, fields, and
non-breaking fixes keep shipping on `v3` as they do today; they don't wait
for this list.

## Consequences

- SDK users get one predictable major-version bump for the accumulated
  breaking changes, instead of a surprise across several `v3` patch/minor
  releases (which would violate semver, since none of these are meant to
  land as anything but a major).
- Until v4 actually starts, this ADR is a to-do list with no deadline. It
  should be revisited whenever a new candidate breaking change comes up
  elsewhere (e.g. while reviewing a PR that would otherwise break a
  legacy-style method's signature) — add it here instead of making it a
  one-off breaking change on `v3`.
- When v4 work does start, it should begin by re-reading this ADR and
  turning each bullet into its own tracked issue, the same way the
  post-fork cleanup this ADR was written alongside did.
