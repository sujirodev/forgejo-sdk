#!/usr/bin/env bash
# Helper for .forgejo/workflows/release.yml: decides the semver bump for a
# merged PR and computes/renders the resulting release metadata. Kept as a
# standalone script (rather than inline YAML) so it can be run and tested
# locally without pushing a workflow change.
set -euo pipefail

usage() {
  echo "usage: $0 {bump-level|next-tag <level>|changelog-group|render-entry|changelog-insert <file> <tag> <group> <entry> <repo-full-name>}" >&2
  exit 1
}

command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }

# Prints one label name per line from PR_LABELS_JSON. A jq failure (e.g.
# malformed JSON) must abort the script rather than be swallowed into an
# empty list, which would silently misclassify a breaking change as patch.
pr_label_names() {
  jq -r '.[].name' <<<"${PR_LABELS_JSON:-[]}"
}

# kind/* label -> CHANGELOG.md group, mirrors .changelog.yml.
changelog_group_for_labels() {
  local labels="$1"
  if grep -qx 'kind/feature' <<<"$labels"; then echo "FEATURES"; return; fi
  if grep -qx 'kind/bug' <<<"$labels"; then echo "BUGFIXES"; return; fi
  if grep -qxE 'kind/(enhancement|refactor|ui)' <<<"$labels"; then echo "ENHANCEMENTS"; return; fi
  if grep -qx 'kind/security' <<<"$labels"; then echo "SECURITY"; return; fi
  if grep -qx 'kind/testing' <<<"$labels"; then echo "TESTING"; return; fi
  if grep -qx 'kind/translation' <<<"$labels"; then echo "TRANSLATION"; return; fi
  if grep -qxE 'kind/(build|lint)' <<<"$labels"; then echo "BUILD"; return; fi
  if grep -qx 'kind/docs' <<<"$labels"; then echo "DOCS"; return; fi
  echo "MISC"
}

# Prints major/minor/patch. Reads PR_LABELS_JSON (Forgejo Label[] JSON),
# PR_TITLE and PR_BODY from the environment.
bump_level() {
  local labels
  labels=$(pr_label_names)
  local title="${PR_TITLE:-}"
  local body="${PR_BODY:-}"

  if grep -qx 'kind/breaking' <<<"$labels"; then
    echo major
    return
  fi
  # Conventional Commits breaking markers: "feat!:", "fix(scope)!:", or a
  # "BREAKING CHANGE:" footer.
  if grep -qE '^[a-zA-Z]+(\([^)]*\))?!:' <<<"$title" || grep -q 'BREAKING CHANGE:' <<<"$body"; then
    echo major
    return
  fi
  if grep -qx 'kind/feature' <<<"$labels"; then
    echo minor
    return
  fi
  if grep -qE '^feat(\([^)]*\))?:' <<<"$title"; then
    echo minor
    return
  fi
  echo patch
}

# Prints the next forgejo/vX.Y.Z tag for the given bump level, based on the
# highest existing stable tag (pre-releases like -alpha are ignored). Must
# run inside a checkout that has the tags fetched.
next_tag() {
  local level="$1"
  local latest
  latest=$(git tag --list 'forgejo/v*' | grep -Ev -- '-' | sort -V | tail -1)
  if [ -z "$latest" ]; then
    echo "no existing forgejo/vX.Y.Z tag found" >&2
    exit 1
  fi
  local ver="${latest#forgejo/v}"
  IFS='.' read -r major minor patch <<<"$ver"
  case "$level" in
    major) major=$((major + 1)); minor=0; patch=0 ;;
    minor) minor=$((minor + 1)); patch=0 ;;
    patch) patch=$((patch + 1)) ;;
    *) echo "unknown bump level: $level" >&2; exit 1 ;;
  esac
  echo "forgejo/v${major}.${minor}.${patch}"
}

# Prints "* PR_TITLE (#PR_NUMBER)" for reuse in both the release body and the
# CHANGELOG.md entry.
render_entry() {
  echo "  * ${PR_TITLE:-} (#${PR_NUMBER:-})"
}

# Prepends a new "## [tag](releases/tag/tag) - date" section, with a single
# group and entry, to the top of the changelog (right after the "# Changelog"
# heading). Writes the result back to the same file.
changelog_insert() {
  local file="$1" tag="$2" group="$3" entry="$4" repo="$5"
  local date url tmp
  date=$(date -u +%Y-%m-%d)
  url="https://codeberg.org/${repo}/releases/tag/${tag}"
  tmp=$(mktemp)
  {
    head -n 1 "$file"
    echo
    echo "## [${tag#forgejo/}](${url}) - ${date}"
    echo
    echo "* ${group}"
    echo "$entry"
    echo
    tail -n +2 "$file"
  } >"$tmp"
  mv "$tmp" "$file"
}

case "${1:-}" in
  bump-level) bump_level ;;
  next-tag) next_tag "${2:?level required}" ;;
  changelog-group) changelog_group_for_labels "$(pr_label_names)" ;;
  render-entry) render_entry ;;
  changelog-insert) changelog_insert "${2:?}" "${3:?}" "${4:?}" "${5:?}" "${6:?}" ;;
  *) usage ;;
esac
