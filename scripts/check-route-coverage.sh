#!/usr/bin/env bash
# Fails when an exported *Client method (a "route": one HTTP call to the
# Forgejo API) has no test referencing it. Without --strict, any *_test.go
# call site `.Method(` counts. With --strict, the method must additionally
# be called from inside a TestUnit_* function AND from inside a non-unit
# (integration) Test* function.
set -euo pipefail
cd "$(dirname "$0")/../forgejo"

strict=0
if [ "${1:-}" = "--strict" ]; then
  strict=1
fi

# No exceptions: every exported *Client method has a test as of fase 7 of
# the route-coverage plan (SetHTTPClient/SetOTP/SetContext/SetUserAgent are
# tested as methods in client_plumbing_unit_test.go; SignRequest there too).
is_exception() {
  return 1
}

test_files=(*_test.go)

# Prints "1" if $2 is called (`.method(`) from inside a function whose name
# matches $1 (an ERE anchored to the function line), "0" otherwise.
called_from() {
  local fn_pattern=$1 method=$2
  awk -v fnpat="$fn_pattern" -v m="$method" '
    /^func / { infn = ($0 ~ fnpat) }
    infn && $0 ~ ("\\." m "\\(") { found = 1 }
    END { print found + 0 }
  ' "${test_files[@]}"
}

total=0
covered=0
fail=0

for f in *.go; do
  case "$f" in
  *_test.go) continue ;;
  esac
  methods=$(grep -oE '^func \(c \*Client\) [A-Z][A-Za-z0-9_]*\(' "$f" | sed -E 's/^func \(c \*Client\) //; s/\($//') || true
  [ -z "$methods" ] && continue

  file_total=0
  file_covered=0
  missing=""

  for m in $methods; do
    file_total=$((file_total + 1))
    total=$((total + 1))

    if is_exception "$m"; then
      file_covered=$((file_covered + 1))
      covered=$((covered + 1))
      continue
    fi

    if [ "$strict" -eq 1 ]; then
      in_unit=$(called_from '^func TestUnit_' "$m")
      in_integration=$(called_from '^func Test[A-Za-z0-9_]+' "$m")
      # A plain non-strict match on ^func Test... also matches TestUnit_, so
      # subtract: integration coverage requires a Test* function that is NOT
      # a TestUnit_* one.
      in_integration_only=$(awk -v m="$m" '
        /^func TestUnit_/ { infn = 0; next }
        /^func Test[A-Za-z0-9_]+/ { infn = 1; next }
        infn && $0 ~ ("\\." m "\\(") { found = 1 }
        END { print found + 0 }
      ' "${test_files[@]}")
      referenced=0
      [ "$in_unit" -ge 1 ] && [ "$in_integration_only" -ge 1 ] && referenced=1
    else
      referenced=0
      grep -qE "\.$m\(" "${test_files[@]}" && referenced=1
    fi

    if [ "$referenced" -eq 1 ]; then
      file_covered=$((file_covered + 1))
      covered=$((covered + 1))
    else
      missing="$missing $m"
      fail=1
    fi
  done

  printf '%-28s %3d/%3d%s\n' "$f" "$file_covered" "$file_total" "${missing:+  missing:$missing}"
done

echo "----"
echo "TOTAL $covered/$total"
if [ "$strict" -eq 1 ]; then
  echo "(strict: requires both a TestUnit_* and an integration Test* reference)"
fi

exit $fail
