# Self-hosted Codeberg runner

Reviewed source of truth for the Portainer Swarm stack `codeberg-runner`
(Portainer environment id 3) that runs `.forgejo/workflows/integration.yml`
against codeberg.org for this repo.

- `stack.yml` — the Docker Swarm stack file, deployed manually through the
  Portainer UI (Stacks -> codeberg-runner -> Editor -> paste this file's
  content -> Update the stack). No CI workflow applies it automatically.
- `entrypoint.sh` — the container entrypoint, deployed as the Swarm config
  named in `stack.yml`'s `configs.codeberg_runner_entrypoint.name`
  (`codeberg_runner_entrypoint_v12` as of this file). Bumping the entrypoint
  requires creating a new Swarm config under a new version suffix and
  pointing `stack.yml` at it — Swarm configs are immutable once created.

## Why deployment is manual, not a workflow

A CI workflow with credentials to redeploy this stack would be able to
modify the very runner that executes it — a privilege-escalation path for
anyone who gets a workflow-file change approved to run. See the "Security of
Pull Requests" note in the PR template and CODEOWNERS. Keeping deploy manual
means a change here only takes effect after a human reads the diff and
applies it themselves.

## Secrets

Nothing sensitive is committed here or should ever be:

- The Forgejo connection token lives only in the Swarm secret
  `codeberg_runner_conn_token` (external, created out of band). A Swarm
  secret's value is never returned by the Docker API once set, unlike env
  vars.
- `FORGEJO_RUNNER_CONN_UUID` is a plain identifier (not a secret, per
  Forgejo's own docs) and is set as a stack env var.

## Caching

Go's build/module caches and the golangci-lint binary are bind-mounted from
`/data/toolcache/{go-build,go-mod,golangci-lint}` on the runner's own
persistent volume straight into every job container, via `container.options`
in `entrypoint.sh`. Two things are required for this to actually take
effect, not just be declared:

1. `container.valid_volumes` must list the exact host paths — act_runner
   denies any volume not in that allowlist (default: empty, i.e. deny all),
   silently ("is not a valid volume, will be ignored") regardless of whether
   the volume comes from a workflow's own `volumes:` or from
   `container.options`. Missing this in v11 meant the bind mounts were
   accepted syntactically but never actually applied.
2. The host paths must already have something in them for a job to benefit
   — the first run after adding a path is unavoidably cold.

Verified on the live runner (v12): a cold run showed the usual `go:
downloading ...` lines for every module; the next run showed none, and
golangci-lint went straight to `0 issues.` without reinstalling.

This replaced an earlier attempt (through v10) to route caching through the
runner's built-in cache proxy (`actions/cache`'s `ACTIONS_CACHE_URL`): that
proxy (port 3101, published in Swarm `mode: host`) turned out to time out
when reached from a per-job isolated bridge network on this host, confirmed
by direct testing. The workflow no longer uses `actions/cache` or
`setup-go`'s own `cache:` option for this reason.
