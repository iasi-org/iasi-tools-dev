#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$(cd -- "$SCRIPT_DIR/../.." && pwd)"

source "$TOOLS_DIR/lib/core/messages.sh"
source "$TOOLS_DIR/lib/core/repositories.sh"

usage() {
  cat <<'EOF'
Usage: iasi-dev sync path [path...]

Propagates files or directories from iasi-common to every existing copy in the
IASI workspace. A simple name is matched by name; a path such as dir/subdir is
resolved relative to iasi-common and matched by that same path. Directory
contents are merged: existing files are updated, missing files are created, and
destination-only files are preserved. iasi-common and Git metadata are excluded,
and missing copies are not created.

Options:
  -h, --help   Show this help
EOF
}

if [ "$#" -eq 0 ]; then
  error "Debes indicar al menos un archivo o directorio."
  usage >&2
  exit 2
fi

if [ "$#" -eq 1 ] && { [ "$1" = "-h" ] || [ "$1" = "--help" ]; }; then
  usage
  exit 0
fi

for argument in "$@"; do
  if [[ "$argument" == -* ]]; then
    error "Opción no válida: $argument"
    usage >&2
    exit 2
  fi
done

WORKSPACE_DIR="$(cd -- "$TOOLS_DIR/.." && pwd)"
COMMON_DIR="${IASI_COMMON_DIR:-$WORKSPACE_DIR/iasi-common}"

if [ ! -d "$COMMON_DIR" ]; then
  error "No se encontró iasi-common: $COMMON_DIR"
  exit 1
fi

synced=0

for argument in "$@"; do
  entry_name="$(basename -- "$argument")"
  source=""
  source_type=""
  path_argument=false

  if [[ "$argument" == */* ]]; then
    path_argument=true

    if [[ "$argument" = /* || "/$argument/" == *"/../"* ]]; then
      error "La ruta debe ser relativa a iasi-common: $argument"
      exit 1
    fi

    source="$COMMON_DIR/${argument#./}"

    if [ ! -e "$source" ]; then
      error "No se encontró $argument en iasi-common."
      exit 1
    fi

    if [ -d "$source" ]; then
      source_type="directory"
    else
      source_type="file"
    fi
  else
    while IFS= read -r -d '' candidate; do
      if [ -n "$source" ]; then
        error "Hay más de una entrada llamada $entry_name en iasi-common."
        exit 1
      fi

      source="$candidate"

      if [ -d "$candidate" ]; then
        source_type="directory"
      else
        source_type="file"
      fi
    done < <(
      find "$COMMON_DIR" \
        -path "*/$IASI_PROJECT_MARKER" -prune -o \
        \( -type f -o -type d \) -name "$entry_name" -print0
    )

    if [ -z "$source" ]; then
      error "No se encontró $entry_name en iasi-common."
      exit 1
    fi
  fi

  found=0

  if [ "$source_type" = "file" ]; then
    while IFS= read -r -d '' target; do
      if ! cp -- "$source" "$target"; then
        error "No se pudo actualizar: $target"
        exit 1
      fi

      success_detail "$target"
      found=$((found + 1))
      synced=$((synced + 1))
    done < <(
      if [ "$path_argument" = true ]; then
        find "$WORKSPACE_DIR" \
          -path "$COMMON_DIR" -prune -o \
          -path "*/$IASI_PROJECT_MARKER" -prune -o \
          -type f -path "*/${argument#./}" -print0
      else
        find "$WORKSPACE_DIR" \
          -path "$COMMON_DIR" -prune -o \
          -path "*/$IASI_PROJECT_MARKER" -prune -o \
          -type f -name "$entry_name" -print0
      fi
    )
  else
    while IFS= read -r -d '' target; do
      case "$target" in
        "$WORKSPACE_DIR"/*) ;;
        *)
          error "Destino fuera del workspace: $target"
          exit 1
          ;;
      esac

      if ! cp -a -- "$source/." "$target/"; then
        error "No se pudo sincronizar el directorio: $target"
        exit 1
      fi

      success_detail "$target"
      found=$((found + 1))
      synced=$((synced + 1))
    done < <(
      if [ "$path_argument" = true ]; then
        find "$WORKSPACE_DIR" \
          -path "$COMMON_DIR" -prune -o \
          -path "*/$IASI_PROJECT_MARKER" -prune -o \
          -type d -path "*/${argument#./}" -print0 -prune
      else
        find "$WORKSPACE_DIR" \
          -path "$COMMON_DIR" -prune -o \
          -path "*/$IASI_PROJECT_MARKER" -prune -o \
          -type d -name "$entry_name" -print0 -prune
      fi
    )
  fi

  display_name="$entry_name"
  if [ "$path_argument" = true ]; then display_name="${argument#./}"; fi

  if [ "$found" -eq 0 ]; then
    warning "No existen copias de $display_name fuera de iasi-common."
  else
    info "$display_name: $found copia(s) actualizada(s)."
  fi
done

success "$synced copia(s) sincronizada(s) desde iasi-common."
