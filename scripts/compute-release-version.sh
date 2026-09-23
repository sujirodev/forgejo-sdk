#!/usr/bin/env bash
# Helper for .forgejo/workflows/release.yml: decides the semver bump for a
# release train and computes/renders the resulting release metadata. Kept
# as a standalone script (rather than inline YAML) so it can be run and
# tested locally without pushing a workflow change.
#
# A "train" is the set of PRs merged into develop since the last stable
# tag, all reachable (via --first-parent) from the tip of develop that the
# develop -> main merge commit incorporated. See docs/PLANO-RELEASE-TRAIN.md
# for the full design and why the range must end at develop's tip, not at
# main's HEAD.
set -euo pipefail

usage() {
  echo "usage: $0 {bump-level|next-tag <level>|changelog-group|render-entry|latest-tag|train-head <merge-sha>|range-prs <range>|aggregate-bump <dir>|render-body <dir>|changelog-insert <file> <tag> <body> <repo-full-name>}" >&2
  exit 1
}

command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }

# PR metadata comes either from PR_NUMBER/PR_TITLE/PR_BODY/PR_LABELS_JSON in
# the environment (handy for local runs) or from a Forgejo PullRequest JSON
# document at PR_JSON_FILE, which takes precedence. The workflow always uses
# the file: the Forgejo runner re-evaluates every value it loads from
# $GITHUB_ENV as an expression, so a PR body that merely quotes "${{ ... }}"
# (like #22's did) aborts the job before any step runs when passed as an
# environment variable.
#
# Loads a single PullRequest JSON file into PR_*. Called once at startup for
# single-PR subcommands (PR_JSON_FILE), and once per file by the aggregating
# subcommands (aggregate-bump, render-body) so bump_level/
# changelog_group_for_labels/render_entry can run per PR.
load_pr() {
  local f="$1"
  PR_NUMBER=$(jq -r '.number' "$f")
  PR_TITLE=$(jq -r '.title // ""' "$f")
  PR_BODY=$(jq -r '.body // ""' "$f")
  PR_LABELS_JSON=$(jq -c '.labels // []' "$f")
  PR_AUTHOR=$(jq -r '.user.login // ""' "$f")
  export PR_NUMBER PR_TITLE PR_BODY PR_LABELS_JSON PR_AUTHOR
}
[ -n "${PR_JSON_FILE:-}" ] && load_pr "$PR_JSON_FILE"

# Prints one label name per line from PR_LABELS_JSON. A jq failure (e.g.
# malformed JSON) must abort the script rather than be swallowed into an
# empty list, which would silently misclassify a breaking change as patch.
pr_label_names() {
  jq -r '.[].name' <<<"${PR_LABELS_JSON:-[]}"
}

# kind/* label -> CHANGELOG.md group, mirrors .changelog.yml. The groups are
# upstream's (GENERAL/FEATURES/FIXES). Without a kind/* label the PR title's
# Conventional Commits type decides (feat -> FEATURES, fix -> FIXES), so an
# unlabeled "feat: ..." PR doesn't land in GENERAL. Reads PR_TITLE.
changelog_group_for_labels() {
  local labels="$1"
  if grep -qx 'kind/breaking' <<<"$labels"; then echo "BREAKING"; return; fi
  if grep -qx 'kind/feature' <<<"$labels"; then echo "FEATURES"; return; fi
  if grep -qx 'kind/bug' <<<"$labels"; then echo "FIXES"; return; fi
  if grep -qx 'kind/security' <<<"$labels"; then echo "SECURITY"; return; fi
  if grep -q '^kind/' <<<"$labels"; then echo "GENERAL"; return; fi
  if grep -qE '^feat(\([^)]*\))?!?:' <<<"${PR_TITLE:-}"; then echo "FEATURES"; return; fi
  if grep -qE '^fix(\([^)]*\))?!?:' <<<"${PR_TITLE:-}"; then echo "FIXES"; return; fi
  echo "GENERAL"
}

# Prints major/minor/patch. Reads PR_LABELS_JSON (Forgejo Label[] JSON),
# PR_TITLE and PR_BODY from the environment. Note: a skip-changelog label
# (checked by render_body, not here) does not affect the bump -- it only
# hides the entry from the changelog text, not the version it implies.
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
  latest=$(latest_tag)
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

# Latest stable tag (pre-releases like -alpha are excluded), the same base
# next_tag uses to decide the next number and the start of a train's range.
latest_tag() {
  git tag --list 'forgejo/v*' | grep -Ev -- '-' | sort -V | tail -1
}

# The tip of develop that the train's merge commit into main incorporated.
# Takes the SHA of that merge commit (a PR's merge_commit_sha) and prints
# its second parent. Fails if the commit is not a merge: that means the
# train was merged by squash or rebase, and the list of PRs it carried is
# not recoverable from it -- merge commit is the only merge style the
# release train plan allows for a train (decision 4).
train_head() {
  local merge_sha="$1" parents
  parents=$(git rev-list --parents -n1 "$merge_sha" | wc -w)
  if [ "$parents" -lt 3 ]; then
    echo "${merge_sha} is not a merge commit; a train must be merged with a merge commit (release train plan, decision 4)" >&2
    exit 1
  fi
  git rev-parse "${merge_sha}^2"
}

# Numbers of the PRs that entered the given range, in the order they
# entered. Must be called with the range ending at develop's tip (the
# second parent of the train's merge commit), never at main: walking
# --first-parent from main only ever sees the train's own merge commit.
# Assumes every PR was merged with a merge commit (decision 4): rebase
# leaves no "(#N)" behind, so a rebased PR would silently not appear here.
range_prs() {
  local range="$1"
  git log --first-parent --reverse --format=%s "$range" \
    | sed -n 's/.*(#\([0-9]\{1,\}\)).*/\1/p' \
    | awk 'NF && !seen[$0]++'
}

# Highest bump among the PRs in dir (one PR JSON file per PR, as fetched by
# release.yml's "Collect the PRs in this train" step). major short-circuits:
# release.yml already fails an automatic MAJOR release (ADR 0005), so there
# is no point evaluating the rest, and the offending PR is named on stderr
# because the workflow's own error otherwise only knows the train PR number.
aggregate_bump() {
  local dir="$1" level=patch f l
  shopt -s nullglob
  for f in "$dir"/*.json; do
    load_pr "$f"
    l=$(bump_level)
    case "$l" in
      major) echo "PR #${PR_NUMBER} signals a breaking change: ${PR_TITLE}" >&2; echo major; return ;;
      minor) level=minor ;;
    esac
  done
  echo "$level"
}

# Renders "* PR_TITLE ([#N](.../issues/N))", with a "(thanks [@author](...))"
# suffix when the PR wasn't opened by the repo owner, for reuse in both the
# release body and the CHANGELOG.md entry. GITHUB_REPOSITORY and
# GITHUB_REPOSITORY_OWNER are set by the runner; the owner check also covers
# Renovate, which opens PRs under the owner's own token.
render_entry() {
  local suffix=""
  local owner="${GITHUB_REPOSITORY_OWNER:-sujirodev}"
  local repo="${GITHUB_REPOSITORY:-sujirodev/forgejo-sdk}"
  if [ -n "${PR_AUTHOR:-}" ] && [ "${PR_AUTHOR}" != "$owner" ]; then
    suffix=" (thanks [@${PR_AUTHOR}](https://codeberg.org/${PR_AUTHOR}))"
  fi
  echo "  * ${PR_TITLE:-} ([#${PR_NUMBER:-}](https://codeberg.org/${repo}/issues/${PR_NUMBER:-}))${suffix}"
}

# Changelog body for every PR in dir: groups in .changelog.yml order, each
# with the PRs that fell into it. A PR with a skip label (skip-changelog,
# backport/*, has/backport -- same regex as .changelog.yml's skip-labels)
# is left out of the text, but its bump level was already counted by
# aggregate_bump: the label speaks to the changelog, not to semver.
#
# Groups and entries follow upstream's release notes: "* GROUP" lines with
# their entries nested under them, no blank lines in between.
#
# Empty output (every PR in the train skipped, or dir has no files) is a
# caller error: release.yml's "Update CHANGELOG.md" step must treat it as
# a hard failure rather than publish a release with no notes.
render_body() {
  local dir="$1" f group labels
  declare -A entries
  shopt -s nullglob
  local files=("$dir"/*.json)
  [ "${#files[@]}" -gt 0 ] || return 0
  # Newest PR first within each group, as upstream lists them.
  for f in $(printf '%s\n' "${files[@]}" | sort -V -r); do
    load_pr "$f"
    labels=$(pr_label_names)
    grep -qE '^(skip-changelog|backport/.+|has/backport)$' <<<"$labels" && continue
    group=$(changelog_group_for_labels "$labels")
    entries[$group]+="$(render_entry)"$'\n'
  done
  for group in BREAKING SECURITY GENERAL FEATURES FIXES; do
    [ -n "${entries[$group]:-}" ] || continue
    echo "* ${group}"
    printf '%s' "${entries[$group]}"
  done
}

# Prepends a new "## [tag](releases/tag/tag) - date" section, with the given
# body, to the top of the changelog (right after the "# Changelog" heading).
# Writes the result back to the same file. No-op when the file already has
# a section for that tag.
changelog_insert() {
  local file="$1" tag="$2" body="$3" repo="$4"
  local date url tmp
  # Idempotent: a re-run of a release job that already committed this
  # section (but failed later, e.g. publishing the release) must not insert
  # it a second time.
  if grep -qF "## [${tag#forgejo/}](" "$file"; then
    echo "${file} already has a ${tag} section, leaving it unchanged" >&2
    return
  fi
  date=$(date -u +%Y-%m-%d)
  url="https://codeberg.org/${repo}/releases/tag/${tag}"
  tmp=$(mktemp)
  {
    head -n 1 "$file"
    echo
    echo "## [${tag#forgejo/}](${url}) - ${date}"
    echo
    printf '%s\n' "$body"
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
  latest-tag) latest_tag ;;
  train-head) train_head "${2:?merge commit SHA required}" ;;
  range-prs) range_prs "${2:?range required}" ;;
  aggregate-bump) aggregate_bump "${2:?directory required}" ;;
  render-body) render_body "${2:?directory required}" ;;
  changelog-insert) changelog_insert "${2:?}" "${3:?}" "${4:?}" "${5:?}" ;;
  *) usage ;;
esac
