# Forgejo SDK for Go

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
![release-badge](https://codeberg.org/mvdkleijn/forgejo-sdk/badges/release.svg)
![status-badge](https://codeberg.org/mvdkleijn/forgejo-sdk/badges/workflows/integration.yml/badge.svg)

![stars-badge](https://codeberg.org/mvdkleijn/forgejo-sdk/badges/stars.svg)
![issues-badge](https://codeberg.org/mvdkleijn/forgejo-sdk/badges/issues/open.svg)
![prs-badge](https://codeberg.org/mvdkleijn/forgejo-sdk/badges/pulls/closed.svg)
[![Go Report Card](https://goreportcard.com/badge/codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2)](https://goreportcard.com/report/codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2)
[![GoDoc](https://godoc.org/codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2?status.svg)](https://godoc.org/codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2)

This project acts as a client SDK implementation written in Go to interact with the Forgejo API implementation. For further informations take a look at the current [documentation](https://pkg.go.dev/codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2).

Note: function arguments are escaped by the SDK.

## Use it

```go
import "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
```

## Version Requirements
 * go >= 1.25
 * forgejo >= 11.0.10
 
 **Please note:** that the SDK might or might not work with lower versions of Forgejo depending on what part of the SDK you use, but it was tested against this one. Starting with release v2 of the SDK, Forgejo versions below 7.x.x will not be supported.

## Contributing

Fork -> Patch -> Push -> Pull Request

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
