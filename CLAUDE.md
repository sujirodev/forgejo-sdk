# Forgejo SDK - Claude Code instructions

## Where the project lives

The source of truth is GitHub: https://github.com/sujirodev/forgejo-sdk.
https://codeberg.org/sujirodev/forgejo-sdk is a read-only mirror of `main`,
`develop` and the `forgejo/v*` tags, kept in sync by `.github/workflows/mirror.yml`.
`upstream` (`mvdkleijn/forgejo-sdk` on Codeberg) is the project this was
forked from.

Use `gh` (or the `github` MCP server) for issues, pull requests, releases,
secrets and CI runs. Do not open issues or PRs on Codeberg: they are closed
there and the repository is not written to by hand (pushes to `main` and
`develop` there come only from the mirror workflow).

The Go module path is `github.com/sujirodev/forgejo-sdk/forgejo/v3`.

## Release train

Feature branches go into `develop` by pull request (merge commit only).
`develop` reaches `main` through a train PR; merging it runs
`.github/workflows/release.yml`, which tags `forgejo/vX.Y.Z`, commits the
changelog to `main` and fast-forwards `develop`. It needs the `RELEASE_TOKEN`
secret (a PAT of a repository admin, able to bypass the branch rulesets).

## CI

GitHub-hosted runners. Required checks on `main` and `develop`: `train-guard`,
`lint`, `testing (lts-11)`, `testing (lts-15)`, `testing (latest-stable)` and
`route-matrix`. `ROUTES.md` and the README route block are generated: when
`route-matrix` fails, take them from the `route-contract` artifact of that run.
