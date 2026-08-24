#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# IASI Tools - Version control system abstraction
# -----------------------------------------------------------------------------

IASI_VCS="${IASI_VCS:-git}"

_vcs_core_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
_vcs_tools_dir="$(cd -- "$_vcs_core_dir/../.." && pwd)"
_vcs_adapter="$_vcs_tools_dir/lib/vcs/$IASI_VCS.sh"

if [ ! -f "$_vcs_adapter" ]; then
  printf 'Unsupported VCS: %s\n' "$IASI_VCS" >&2
  return 2 2>/dev/null || exit 2
fi

# shellcheck source=/dev/null
source "$_vcs_adapter"

_required_vcs_functions=(
  vcs_name
  vcs_project_marker
  vcs_is_project_root
  vcs_stage_all
  vcs_has_changes
  vcs_commit
  vcs_push
  vcs_clone
  vcs_configure_remote
  vcs_sync_to_remote
)

for _vcs_function in "${_required_vcs_functions[@]}"; do
  if ! declare -F "$_vcs_function" >/dev/null 2>&1; then
    printf 'Invalid VCS adapter %s: missing %s\n' "$IASI_VCS" "$_vcs_function" >&2
    return 2 2>/dev/null || exit 2
  fi
done

unset _required_vcs_functions _vcs_function

unset _vcs_core_dir _vcs_tools_dir _vcs_adapter
