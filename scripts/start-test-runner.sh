#!/usr/bin/env bash
# Starts a forgejo-runner container attached to a running test instance, so
# a dispatched workflow really executes (issue #77: the artifact and job-log
# routes can only answer 2xx once a job has run). Used by `make
# test-instance-docker` and by the CI workflow so both build the runner the
# same way.
#
# usage: start-test-runner.sh <forgejo-container> <instance-url> <network> [forgejo-cli args...]
#
#   forgejo-container  name of the running Forgejo test container
#   instance-url       the URL the runner (not the host) reaches the instance
#                      at; it has to equal the instance's ROOT_URL, because
#                      that is the address Forgejo hands to a job for the
#                      artifact service
#   network            passed to `docker run --network`: a network name, or
#                      container:<name> to share the instance's network
#                      namespace (then localhost:3000 is the instance)
#   forgejo-cli args   extra flags for the `forgejo forgejo-cli` call, e.g.
#                      `-c /tmp/conf/app.ini`
#
# Environment: FORGEJO_EXEC_USER (docker exec --user for the forgejo call),
# RUNNER_CONTAINER (default forgejo-test-runner), RUNNER_IMAGE.
#
# Jobs run in host mode (label "host"), inside the runner container: no
# Docker socket and no docker-in-docker. The image is Alpine and carries no
# node, which the upload-artifact action needs, so it is added at startup.
set -euo pipefail

forgejo_container=$1
instance_url=$2
network=$3
shift 3

runner_container=${RUNNER_CONTAINER:-forgejo-test-runner}
# renovate: datasource=docker depName=code.forgejo.org/forgejo/runner
runner_image=${RUNNER_IMAGE:-code.forgejo.org/forgejo/runner:12.13.2}

exec_user=()
if [ -n "${FORGEJO_EXEC_USER:-}" ]; then
  exec_user=(--user "$FORGEJO_EXEC_USER")
fi

token=$(docker exec ${exec_user[@]+"${exec_user[@]}"} "$forgejo_container" \
  forgejo forgejo-cli actions generate-runner-token "$@" | tail -n 1)
if [ -z "$token" ]; then
  echo "error: could not get a runner registration token from $forgejo_container" >&2
  exit 1
fi

docker rm -f "$runner_container" > /dev/null 2>&1 || true
docker run -d --name "$runner_container" --network "$network" --user 0 \
  --entrypoint sh "$runner_image" -c '
    apk add --no-cache nodejs > /dev/null &&
    forgejo-runner register --no-interactive --instance "$0" --token "$1" \
      --name sdk-test-runner --labels host:host &&
    exec forgejo-runner daemon
  ' "$instance_url" "$token" > /dev/null

# Registration and node take a few seconds; the tests poll for a terminal run
# state with their own deadline, so this only reports, it does not gate.
for _ in $(seq 1 60); do
  if docker logs "$runner_container" 2>&1 | grep -q 'fetching task\|declared successfully\|runner: .* started'; then
    echo "test runner is up"
    exit 0
  fi
  sleep 1
done
echo "warning: test runner not confirmed up; its log:" >&2
docker logs "$runner_container" >&2 || true
