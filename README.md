# Forgejo SDK for Go

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
![release-badge](https://codeberg.org/MatheusAlves96/forgejo-sdk/badges/release.svg)
![status-badge](https://codeberg.org/MatheusAlves96/forgejo-sdk/badges/workflows/integration.yml/badge.svg)

![stars-badge](https://codeberg.org/MatheusAlves96/forgejo-sdk/badges/stars.svg)
![issues-badge](https://codeberg.org/MatheusAlves96/forgejo-sdk/badges/issues/open.svg)
![prs-badge](https://codeberg.org/MatheusAlves96/forgejo-sdk/badges/pulls/closed.svg)
[![Go Report Card](https://goreportcard.com/badge/codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3)](https://goreportcard.com/report/codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3)
[![GoDoc](https://godoc.org/codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3?status.svg)](https://godoc.org/codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3)

This project is a client SDK implementation written in Go to interact with the Forgejo API implementation. For further informations take a look at the current [documentation](https://pkg.go.dev/codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3).

Note: function arguments are escaped by the SDK.

## Relationship to upstream

This is an independently maintained base derived from [mvdkleijn/forgejo-sdk](https://codeberg.org/mvdkleijn/forgejo-sdk). We started here as a fork to send PRs upstream, and decided to keep developing on our own copy going forward: as of 2026-09-18, upstream's `main` hadn't moved since 2026-08-31, its latest tagged release ([forgejo/v3.0.0](https://codeberg.org/mvdkleijn/forgejo-sdk/releases/tag/forgejo/v3.0.0)) had been out since 2026-03-05 (over 6 months, despite further commits landing on `main` after it), and it had 8 open issues / 9 open PRs, some sitting without a maintainer response for months (e.g. [#159](https://codeberg.org/mvdkleijn/forgejo-sdk/issues/159), open since 2026-05-17). None of that is a knock on the maintainer — it's a one-person project — it's just why we didn't want our own usage blocked on that review queue.

We still track upstream (`git remote add upstream https://codeberg.org/mvdkleijn/forgejo-sdk.git`) and pull in changes when useful, and may still send PRs back when a fix is generally applicable.

## Use it

```go
import "codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3"
```

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
| 11.0.16 (LTS) | 341 | 6 | 10 | 357 |
| 15.0.9 (LTS) | 347 | 0 | 10 | 357 |
| 16.0.5 (latest stable) | 347 | 0 | 10 | 357 |

Generated from the integration suite on every pull request.
Per-route detail: [`ROUTES.md`](ROUTES.md).
Declared exceptions: [`route-exceptions.json`](route-exceptions.json).
<!-- route-matrix:end -->

## Contributing

Fork -> Patch -> Push -> Pull Request

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
