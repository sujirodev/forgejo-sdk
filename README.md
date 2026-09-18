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

This is an independently maintained base derived from [mvdkleijn/forgejo-sdk](https://codeberg.org/mvdkleijn/forgejo-sdk). We started here as a fork to send PRs upstream, and decided to keep developing on our own copy going forward: as of 2026-09-18, upstream's `main` hadn't moved since 2026-08-31 and had 8 open issues / 9 open PRs, some sitting without a maintainer response for months (e.g. [#159](https://codeberg.org/mvdkleijn/forgejo-sdk/issues/159), open since 2026-05-17). None of that is a knock on the maintainer — it's a one-person project — it's just why we didn't want our own usage blocked on that review queue.

We still track upstream (`git remote add upstream https://codeberg.org/mvdkleijn/forgejo-sdk.git`) and pull in changes when useful, and may still send PRs back when a fix is generally applicable.

## Use it

```go
import "codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3"
```

## Version Requirements
 * go >= 1.25
 * forgejo >= 11.0.10
 
 **Please note:** that the SDK might or might not work with lower versions of Forgejo
 depending on what part of the SDK you use, but it was tested against this one.

## Contributing

Fork -> Patch -> Push -> Pull Request

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
