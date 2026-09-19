#!/usr/bin/env bash
# Fails when the Forgejo test-instance version differs between the places that pin it.
# Renovate updates all of them in one PR; this guards against a partial bump.
set -euo pipefail
cd "$(dirname "$0")/.."

# The testing job's version matrix has three image lines (lts-11, lts-15,
# latest-stable); only latest-stable is meant to match Makefile/README/
# main_test.go, so pull the line right after its "label: latest-stable"
# marker instead of just the first "forgejo/forgejo:" match in the file.
wf=$(awk '/label: latest-stable/{found=1; next} found && /forgejo\/forgejo:/{print; exit}' .forgejo/workflows/integration.yml \
  | grep -oE 'forgejo/forgejo:[0-9]+\.[0-9]+\.[0-9]+' | cut -d: -f2)
mk=$(grep -oE '^FORGEJO_VERSION := [0-9]+\.[0-9]+\.[0-9]+' Makefile | awk '{print $3}')
rd=$(grep -A1 'renovate: datasource=docker depName=codeberg.org/forgejo/forgejo' README.md | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
gt=$(grep -oE 'testForgejoVersion = "[0-9]+\.[0-9]+\.[0-9]+"' forgejo/main_test.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')

printf '%-38s %s\n' \
  '.forgejo/workflows/integration.yml' "${wf:-<missing>}" \
  'Makefile' "${mk:-<missing>}" \
  'README.md' "${rd:-<missing>}" \
  'forgejo/main_test.go' "${gt:-<missing>}"

if [ -z "$wf" ] || [ -z "$mk" ] || [ -z "$rd" ] || [ -z "$gt" ]; then
  echo "error: could not extract the Forgejo version from every file" >&2
  exit 1
fi
if [ "$wf" != "$mk" ] || [ "$wf" != "$rd" ] || [ "$wf" != "$gt" ]; then
  echo "error: Forgejo version drift between files" >&2
  exit 1
fi
echo "ok: every reference pins Forgejo $wf"
