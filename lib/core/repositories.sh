#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# IASI Tools - Core repository list
# -----------------------------------------------------------------------------

CORE_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source "$CORE_DIR/vcs.sh" || {
  rc=$?
  return "$rc" 2>/dev/null || exit "$rc"
}

IASI_ORG="${IASI_ORG:-iasi-org}"
IASI_PROJECT_MARKER="$(vcs_project_marker)"

iasi_repositories() {
  gh repo list "$IASI_ORG" \
    --limit 100 \
    --no-archived \
    --json name,sshUrl,defaultBranchRef \
    --jq '.[] | "\(.name)|\(.sshUrl)|\(.defaultBranchRef.name)"'
}

is_project_root() {
  vcs_is_project_root "$1"
}

repository_has_iasi_project() {
  local repository="$1"
  local match=""

  match="$(
    find "$repository" \
      -path "*/$IASI_PROJECT_MARKER" -prune -o \
      -path '*/.quarto' -prune -o \
      -path '*/publish' -prune -o \
      -path '*/.codex*' -prune -o \
      -path '*/tests' -prune -o \
      -path '*/node_modules' -prune -o \
      -path '*/renv' -prune -o \
      -type f -name '_iasi.yml' -print -quit
  )"

  [ -n "$match" ]
}
