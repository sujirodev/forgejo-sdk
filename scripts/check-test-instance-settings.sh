#!/usr/bin/env bash
# Fails when the extra app.ini settings the route-coverage tests need
# (DISABLE_GIT_HOOKS=false, [migrations] ALLOW_LOCALNETWORKS=true) are
# missing from any of the three places that build a test instance's app.ini:
# Makefile (test-instance-native and test-instance-docker), the CI workflow's
# `forgejo` service, and main_test.go's runForgejo().
set -euo pipefail
cd "$(dirname "$0")/.."

check() {
  local label=$1 file=$2 pattern=$3
  if grep -qE "$pattern" "$file"; then
    printf '%-45s ok\n' "$label"
  else
    printf '%-45s MISSING\n' "$label"
    fail=1
  fi
}

fail=0

check 'Makefile (native): DISABLE_GIT_HOOKS' Makefile 'DISABLE_GIT_HOOKS = false'
check 'Makefile (docker): DISABLE_GIT_HOOKS' Makefile 'FORGEJO__security__DISABLE_GIT_HOOKS=false'
check 'integration.yml: DISABLE_GIT_HOOKS' .forgejo/workflows/integration.yml 'DISABLE_GIT_HOOKS = false'
check 'main_test.go: DISABLE_GIT_HOOKS' forgejo/main_test.go 'DISABLE_GIT_HOOKS = false'

check 'Makefile (native): ALLOW_LOCALNETWORKS' Makefile 'ALLOW_LOCALNETWORKS = true'
check 'Makefile (docker): ALLOW_LOCALNETWORKS' Makefile 'FORGEJO__migrations__ALLOW_LOCALNETWORKS=true'
check 'integration.yml: ALLOW_LOCALNETWORKS' .forgejo/workflows/integration.yml 'ALLOW_LOCALNETWORKS = true'
check 'main_test.go: ALLOW_LOCALNETWORKS' forgejo/main_test.go 'ALLOW_LOCALNETWORKS = true'

if [ "$fail" -ne 0 ]; then
  echo "error: test-instance settings drift between files" >&2
  echo "  DISABLE_GIT_HOOKS = false is needed by the git-hook routes' tests," >&2
  echo "  ALLOW_LOCALNETWORKS = true is needed by the mirror routes' tests." >&2
  exit 1
fi
echo "ok: DISABLE_GIT_HOOKS and ALLOW_LOCALNETWORKS are set consistently"
