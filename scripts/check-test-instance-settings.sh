#!/usr/bin/env bash
# Fails when the extra app.ini settings the route-coverage tests need are
# missing from any of the four places that build a test instance's app.ini:
# Makefile (test-instance-native and test-instance-docker), the CI workflow's
# `forgejo` service, and main_test.go's runForgejo(). See TESTING.md, "Server
# configuration": a setting missing from one of the four is what turned six
# repo-flag routes and nine ActivityPub routes into "needs-config" exceptions
# with no test proving they work (issues #62, #58).
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
check 'integration.yml: DISABLE_GIT_HOOKS' .github/workflows/integration.yml 'DISABLE_GIT_HOOKS = false'
check 'main_test.go: DISABLE_GIT_HOOKS' forgejo/main_test.go 'DISABLE_GIT_HOOKS = false'

check 'Makefile (native): ALLOW_LOCALNETWORKS' Makefile 'ALLOW_LOCALNETWORKS = true'
check 'Makefile (docker): ALLOW_LOCALNETWORKS' Makefile 'FORGEJO__migrations__ALLOW_LOCALNETWORKS=true'
check 'integration.yml: ALLOW_LOCALNETWORKS' .github/workflows/integration.yml 'ALLOW_LOCALNETWORKS = true'
check 'main_test.go: ALLOW_LOCALNETWORKS' forgejo/main_test.go 'ALLOW_LOCALNETWORKS = true'

check 'Makefile (native): ENABLE_FLAGS' Makefile 'ENABLE_FLAGS = true'
check 'Makefile (docker): ENABLE_FLAGS' Makefile 'FORGEJO__repository__ENABLE_FLAGS=true'
check 'integration.yml: ENABLE_FLAGS' .github/workflows/integration.yml 'ENABLE_FLAGS = true'
check 'main_test.go: ENABLE_FLAGS' forgejo/main_test.go 'ENABLE_FLAGS = true'

check 'Makefile (native): quota ENABLED' Makefile '\[quota\]'
check 'Makefile (docker): quota ENABLED' Makefile 'FORGEJO__quota__ENABLED=true'
check 'integration.yml: quota ENABLED' .github/workflows/integration.yml '\[quota\]'
check 'main_test.go: quota ENABLED' forgejo/main_test.go '\[quota\]'

check 'Makefile (native): federation ENABLED' Makefile '\[federation\]'
check 'Makefile (docker): federation ENABLED' Makefile 'FORGEJO__federation__ENABLED=true'
check 'integration.yml: federation ENABLED' .github/workflows/integration.yml '\[federation\]'
check 'main_test.go: federation ENABLED' forgejo/main_test.go '\[federation\]'

# The ActivityPub actor sub-routes authenticate their caller with an HTTP
# Signature whose key ID Forgejo resolves by fetching it over HTTP. From
# Forgejo 16.0.0 that fetch refuses private and loopback addresses unless
# INSECURE_ALLOW_INVALID_HOSTS is set -- and a test instance's peer is always
# on one. Without it the suite cannot present a resolvable key and the nine
# actor routes fall back to "needs-config" (issue #58). Test instances only:
# it disables an SSRF guard.
check 'Makefile (native): INSECURE_ALLOW_INVALID_HOSTS' Makefile 'INSECURE_ALLOW_INVALID_HOSTS = true'
check 'Makefile (docker): INSECURE_ALLOW_INVALID_HOSTS' Makefile 'FORGEJO__federation__INSECURE_ALLOW_INVALID_HOSTS=true'
check 'integration.yml: INSECURE_ALLOW_INVALID_HOSTS' .github/workflows/integration.yml 'INSECURE_ALLOW_INVALID_HOSTS = true'
check 'main_test.go: INSECURE_ALLOW_INVALID_HOSTS' forgejo/main_test.go 'INSECURE_ALLOW_INVALID_HOSTS = true'

if [ "$fail" -ne 0 ]; then
  echo "error: test-instance settings drift between files" >&2
  echo "  DISABLE_GIT_HOOKS = false is needed by the git-hook routes' tests," >&2
  echo "  ALLOW_LOCALNETWORKS = true is needed by the mirror routes' tests," >&2
  echo "  ENABLE_FLAGS = true is needed by the repo-flags routes' tests," >&2
  echo "  [quota]/[federation] ENABLED = true are needed by the quota and" >&2
  echo "  ActivityPub routes' tests, and [federation]" >&2
  echo "  INSECURE_ALLOW_INVALID_HOSTS = true lets Forgejo 16+ fetch the" >&2
  echo "  suite's signing key from a peer on a private address." >&2
  exit 1
fi
echo "ok: test-instance settings are set consistently across all four places"
