# Self-hosted Codeberg runner

Reviewed source of truth for the Portainer Swarm stack `codeberg-runner`
(Portainer environment id 3) that runs `.forgejo/workflows/integration.yml`
against codeberg.org for this repo.

- `stack.yml` — the Docker Swarm stack file, deployed manually through the
  Portainer UI (Stacks -> codeberg-runner -> Editor -> paste this file's
  content -> Update the stack). No CI workflow applies it automatically.
- `entrypoint.sh` — the container entrypoint, deployed as the Swarm config
  named in `stack.yml`'s `configs.codeberg_runner_entrypoint.name`
  (`codeberg_runner_entrypoint_v10` as of this file). Bumping the entrypoint
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

## Known limitation

See the "Known limitation as of v10" comment in `entrypoint.sh`: the cache
proxy is currently unreachable from the per-job isolated networks this setup
uses, so `actions/cache` steps miss every run. Planned fix: bind-mount
persistent host directories (GOMODCACHE, GOCACHE, golangci-lint) into job
containers via `container.options` instead of relying on the proxy.
