# Forgejo SDK for Go

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
![release-badge](https://img.shields.io/github/v/tag/sujirodev/forgejo-sdk?filter=forgejo%2Fv*&label=release)
![status-badge](https://github.com/sujirodev/forgejo-sdk/actions/workflows/integration.yml/badge.svg)

![stars-badge](https://img.shields.io/github/stars/sujirodev/forgejo-sdk)
![issues-badge](https://img.shields.io/github/issues/sujirodev/forgejo-sdk)
![prs-badge](https://img.shields.io/github/issues-pr-closed/sujirodev/forgejo-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/sujirodev/forgejo-sdk/forgejo/v3)](https://goreportcard.com/report/github.com/sujirodev/forgejo-sdk/forgejo/v3)
[![GoDoc](https://godoc.org/github.com/sujirodev/forgejo-sdk/forgejo/v3?status.svg)](https://godoc.org/github.com/sujirodev/forgejo-sdk/forgejo/v3)

This project is a client SDK implementation written in Go to interact with the Forgejo API implementation. For further informations take a look at the current [documentation](https://pkg.go.dev/github.com/sujirodev/forgejo-sdk/forgejo/v3).

Note: function arguments are escaped by the SDK.

## Relationship to upstream

This is an independently maintained base derived from [mvdkleijn/forgejo-sdk](https://codeberg.org/mvdkleijn/forgejo-sdk). We started here as a fork to send PRs upstream, and decided to keep developing on our own copy going forward: as of 2026-09-18, upstream's `main` hadn't moved since 2026-08-31, its latest tagged release ([forgejo/v3.0.0](https://codeberg.org/mvdkleijn/forgejo-sdk/releases/tag/forgejo/v3.0.0)) had been out since 2026-03-05 (over 6 months, despite further commits landing on `main` after it), and it had 8 open issues / 9 open PRs, some sitting without a maintainer response for months (e.g. [#159](https://codeberg.org/mvdkleijn/forgejo-sdk/issues/159), open since 2026-05-17). None of that is a knock on the maintainer — it's a one-person project — it's just why we didn't want our own usage blocked on that review queue.

This repository lives on GitHub. [codeberg.org/sujirodev/forgejo-sdk](https://codeberg.org/sujirodev/forgejo-sdk) is a read-only mirror of `main`, `develop` and the release tags; issues and pull requests belong here.

We still track upstream (`git remote add upstream https://codeberg.org/mvdkleijn/forgejo-sdk.git`) and pull in changes when useful, and may still send PRs back when a fix is generally applicable.

## Use it

```sh
go get github.com/sujirodev/forgejo-sdk/forgejo/v3
```

```go
import "github.com/sujirodev/forgejo-sdk/forgejo/v3"
```

Reference docs: [pkg.go.dev](https://pkg.go.dev/github.com/sujirodev/forgejo-sdk/forgejo/v3).

Releases are tagged `forgejo/vX.Y.Z`, because the module lives in the
`forgejo/` subdirectory; for the Go tooling the version is `vX.Y.Z`, without
that prefix. Besides Codeberg's automatic source archives, every release
carries two attachments: a CycloneDX SBOM of the module and its dependencies,
and the [route contract](#route-coverage) pinned to that version.

## Version Requirements
 * go >= 1.25
 * forgejo >= 11.0.10 (minimum supported)
<!-- renovate: datasource=docker depName=codeberg.org/forgejo/forgejo -->
 * tested against forgejo 16.0.5 (the CI test instance, bumped automatically by Renovate)

 **Please note:** the SDK might or might not work with versions between the minimum and the tested one
 depending on what part of the SDK you use; CI only runs against the tested version.

## Route coverage

Every route the SDK exposes is one exported `*Client` method. The table below
is generated from what the integration suite actually did against each Forgejo
version -- not from reading the tests -- and CI regenerates it on every pull
request, so a stale table is a red build instead of a promise nobody checked.

<!-- route-matrix:start -->
| Forgejo | Routes ok | n/a (guard) | Declared exception | Total |
|---|---|---|---|---|
| 11.0.16 (LTS) | 459 | 45 | 16 | 520 |
| 15.0.9 (LTS) | 489 | 14 | 17 | 520 |
| 16.0.5 (latest stable) | 499 | 0 | 21 | 520 |

Generated from the integration suite on every pull request.
Per-route detail: [`ROUTES.md`](ROUTES.md).
Declared exceptions: [`route-exceptions.json`](route-exceptions.json).
<!-- route-matrix:end -->

## Contributing

Fork -> Patch -> Push -> Pull Request

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
