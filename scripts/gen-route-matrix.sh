#!/usr/bin/env bash
# Generates ROUTES.md and the route-matrix block of README.md from the route
# reports the integration suite wrote (see docs/PLANO-CONTRATO-ROTAS.md).
#
#   scripts/gen-route-matrix.sh [--check] [report.json...]
#
# With no report arguments it uses forgejo/route-report.json, which is what a
# local `make test` leaves behind. CI passes one report per matrix leg.
# --check writes nothing and fails when the committed files are stale.
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)

args=()
if [ "${1:-}" = "--check" ]; then
  args+=(-check)
  shift
fi

reports=()
for r in "$@"; do
  # The generator runs from scripts/routematrix, so relative report paths
  # given by the caller have to be resolved here.
  case "$r" in
  /* | [A-Za-z]:?*) reports+=("$r") ;;
  *) reports+=("$(pwd)/$r") ;;
  esac
done

cd "$root/scripts/routematrix"
exec go run . -root "$root" "${args[@]}" ${reports[@]+"${reports[@]}"}
