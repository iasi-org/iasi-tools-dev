#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# IASI Tools - Core repository list
# -----------------------------------------------------------------------------

IASI_ORG="${IASI_ORG:-iasi-org}"

iasi_repositories() {
  gh repo list "$IASI_ORG" \
    --limit 100 \
    --no-archived \
    --json name,sshUrl,defaultBranchRef \
    --jq '.[] | "\(.name)|\(.sshUrl)|\(.defaultBranchRef.name)"'
}

repository_has_iasi_project() {
  local repository="$1"
  local match=""

  match="$(
    find "$repository" \
      -path '*/.git' -prune -o \
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
