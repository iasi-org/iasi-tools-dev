#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# IASI Tools - Git VCS adapter
# -----------------------------------------------------------------------------

vcs_name() {
  printf '%s\n' "git"
}

vcs_project_marker() {
  printf '%s\n' ".git"
}

vcs_is_project_root() {
  local candidate="$1"
  local marker="$(vcs_project_marker)"

  [ -d "$candidate" ] || return 1
  [ -e "$candidate/$marker" ]
}

vcs_stage_all() {
  local repository="$1"
  git -C "$repository" add -A -- .
}

vcs_has_changes() {
  local repository="$1"

  if git -C "$repository" diff --cached --quiet; then
    return 1
  fi

  return 0
}

vcs_commit() {
  local repository="$1"
  local message="$2"
  git -C "$repository" commit -m "$message"
}

vcs_push() {
  local repository="$1"
  git -C "$repository" push
}

vcs_clone() {
  local url="$1"
  local target="$2"
  git clone "$url" "$target"
}

vcs_configure_remote() {
  local repository="$1"
  local remote="$2"
  local url="$3"

  if git -C "$repository" remote get-url "$remote" >/dev/null 2>&1; then
    git -C "$repository" remote set-url "$remote" "$url"
  else
    git -C "$repository" remote add "$remote" "$url"
  fi
}

vcs_sync_to_remote() {
  local repository="$1"
  local remote="$2"
  local branch="$3"

  git -C "$repository" fetch --prune "$remote" &&
    git -C "$repository" checkout -f -B "$branch" "$remote/$branch" &&
    git -C "$repository" reset --hard "$remote/$branch" &&
    git -C "$repository" clean -fd
}
