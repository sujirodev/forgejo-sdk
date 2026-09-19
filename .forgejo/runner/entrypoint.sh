#!/bin/sh
set -eu

log() {
  echo "[codeberg-runner-entrypoint] $*"
}

TOKEN_FILE="/run/secrets/forgejo_runner_conn_token"

# Forgejo runner v13+ deprecated `register --token`; Codeberg's "Create new
# Runner" UI now shows a static connections block instead (server.connections.
# <name> with url/uuid/token). Per Forgejo's own docs the uuid is a plain
# identifier, only the token is the confidential secret -- so url/uuid come
# from plain env vars, and the token comes from a mounted Docker secret
# (never an env var: a Swarm secret's value is never returned by the API
# again, unlike env vars, which `docker service inspect` shows in cleartext).
# Labels are declared separately under top-level `runner.labels`, one YAML
# list item per comma-separated entry in FORGEJO_RUNNER_LABELS -- omitting
# this section leaves the runner registered with zero labels, matching no
# workflow's `runs-on`.
#
# v7: added `cache` + `container.network`. Job containers are spawned via the
# mounted docker.sock as standalone siblings, NOT inside this service's
# Swarm overlay -- act_runner's auto-detected cache address (an IP on
# docker_gwbridge) is unreachable from them (ETIMEDOUT, dropped by Docker's
# inter-network isolation), so every actions/cache save was silently failing.
#
# v8: added `runner.capacity` from FORGEJO_RUNNER_CAPACITY (default 2) so the
# forgejo-sdk `lint` and `testing` jobs of one workflow run side by side
# instead of queueing. The host has 4 CPU / 8 GB; keep it at 2.
#
# v9: the v7 fix (every job on one shared overlay) broke with capacity 2:
# two concurrent `testing` jobs both got a service container aliased
# `forgejo` on the same network, and one job's tests hit the other job's
# instance ("repository already exists"). Now `container.network` is left
# empty so act_runner creates an isolated network per job (services only
# resolve inside their own job).
#
# v10: cache.dir moved to /data/cache (the default, $HOME/.cache/actcache,
# is not on the codeberg_runner_data volume, so every stack redeploy threw
# the whole actions/cache store away). Turned out moot: the cache proxy
# (port 3101, Swarm `mode: host` publish) times out when reached from a
# per-job isolated bridge network on this host -- confirmed by direct
# testing (curl to the proxy times out even via the bridge's own gateway
# IP, while the same port answers instantly from the host's own network
# namespace). Root cause not fully diagnosed; looks like Swarm's host-mode
# publish not hairpinning back through other bridge networks here.
#
# v11: stopped depending on the cache proxy for Go's own caches. Instead,
# `container.options` bind-mounts persistent host directories straight into
# every job container -- no network path needed at all. Turned out to be a
# no-op: act_runner has a `container.valid_volumes` allowlist that defaults
# to empty (deny all), applied to every volume regardless of whether it
# comes from workflow-declared `volumes:` or operator-set `container.options`
# -- the bind mounts were silently dropped ("is not a valid volume, will be
# ignored"), confirmed in a run's log, so caching was still a no-op.
#
# v12: added `container.valid_volumes` listing the exact host paths used by
# `container.options`, so the bind mounts actually take effect. Verified:
# two back-to-back workflow_dispatch runs, first one cold (v11 never wrote
# anything to these host paths), second one with zero `go: downloading`
# lines and no golangci-lint reinstall.
if [ -n "${FORGEJO_RUNNER_CONN_URL:-}" ] && [ -n "${FORGEJO_RUNNER_CONN_UUID:-}" ] && [ -f "$TOKEN_FILE" ]; then
  log "Connection fields + token secret found. Starting daemon directly (no register step)."
  mkdir -p /data/toolcache/go-build /data/toolcache/go-mod /data/toolcache/golangci-lint
  cat > /data/config.yml <<EOF
server:
  connections:
    forgejo:
      url: ${FORGEJO_RUNNER_CONN_URL}
      uuid: ${FORGEJO_RUNNER_CONN_UUID}
      token: $(cat "$TOKEN_FILE")
runner:
  capacity: ${FORGEJO_RUNNER_CAPACITY:-2}
  labels:
$(printf '%s' "${FORGEJO_RUNNER_LABELS:-}" | tr ',' '\n' | sed '/^$/d; s/^/    - /')
container:
  network: "${FORGEJO_RUNNER_JOB_NETWORK:-}"
  valid_volumes:
    - /data/toolcache/go-build
    - /data/toolcache/go-mod
    - /data/toolcache/golangci-lint
  options: >-
    --volume /data/toolcache/go-build:/root/.cache/go-build
    --volume /data/toolcache/go-mod:/root/go/pkg/mod
    --volume /data/toolcache/golangci-lint:/root/go/bin
EOF
  exec forgejo-runner daemon --config /data/config.yml
fi

log "Static connection (URL/UUID env + token secret) not fully set; falling back to the legacy register flow."

# --- legacy fallback: token-based register loop (deprecated upstream) ---
if [ ! -f /data/config.yml ]; then
  forgejo-runner generate-config > /data/config.yml
fi

while true; do
  if [ -f /data/.runner ]; then
    log "Runner registrado encontrado. Iniciando daemon..."
    if forgejo-runner daemon --config /data/config.yml; then
      exit 0
    fi
    log "Daemon encerrou com erro. Limpando estado do runner para novo registro."
    rm -f /data/.runner
    sleep 5
    continue
  fi

  if [ -z "${FORGEJO_RUNNER_REGISTRATION_TOKEN:-}" ]; then
    log "Nenhuma credencial definida (conexao estatica ou token de registro legado). Aguardando 30s..."
    sleep 30
    continue
  fi

  log "Tentando registrar runner (fluxo legado) em ${FORGEJO_RUNNER_INSTANCE_URL}..."
  if forgejo-runner register \
    --no-interactive \
    --instance "${FORGEJO_RUNNER_INSTANCE_URL}" \
    --token "${FORGEJO_RUNNER_REGISTRATION_TOKEN}" \
    --name "${FORGEJO_RUNNER_NAME}" \
    --labels "${FORGEJO_RUNNER_LABELS}"; then
    log "Registro concluido com sucesso."
    continue
  fi

  log "Falha no registro do runner (token invalido/expirado ou instancia nao pronta). Nova tentativa em 30s."
  sleep 30
done
