#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$(cd -- "$SCRIPT_DIR/../.." && pwd)"

source "$TOOLS_DIR/lib/core/messages.sh"
source "$TOOLS_DIR/lib/core/arguments.sh"
source "$TOOLS_DIR/lib/core/repositories.sh"

usage() {
  cat <<EOF
Usage: iasi clone [options] [workspace]

Recreates from scratch all repositories from $IASI_ORG.
Existing repository destinations are removed before cloning.
Command output is written to logs/iasi-clone-YYYYMMDDhhmmss.log in the workspace.

If workspace is omitted, the current directory is used.

Options:
  -h, --help   Show this help
  -v           Detailed information
  -s           Silent mode
  -y, --yes    Do not ask for confirmation
  -r, --resume Clone only repositories that do not exist locally
EOF
}

clone_repository() {
  local name="$1"
  local url="$2"
  local target="$3"
  local attempt=1
  local delay=0
  local clone_output=""

  while [ "$attempt" -le 3 ]; do
    if clone_output="$(vcs_clone "$url" "$target" 2>&1)"; then
      printf "Attempt %s for %s\n%s\n" "$attempt" "$name" "$clone_output" >> "$LOG_FILE"
      return 0
    fi

    printf "Attempt %s for %s failed\n%s\n" "$attempt" "$name" "$clone_output" >> "$LOG_FILE"

    if [ "$attempt" -ge 3 ] || \
       ! printf "%s\n" "$clone_output" | grep -Eqi \
         'connection (closed|reset|timed out|refused)|reset by peer|remote end hung up|early EOF|unexpected disconnect|operation timed out|could not resolve host|network is unreachable'; then
      return 1
    fi

    case "$attempt" in
      1) delay=5 ;;
      2) delay=15 ;;
    esac

    warning "Conexión interrumpida al clonar $name. Reintentando en ${delay}s."

    if [ -e "$target" ] || [ -L "$target" ]; then
      if ! rm -rf -- "$target" >> "$LOG_FILE" 2>&1; then
        return 1
      fi
    fi

    sleep "$delay"
    attempt=$((attempt + 1))
  done
}

workspace_argument=""
options=()
assume_yes=0
resume=0

for argument in "$@"; do
  case "$argument" in
    -y|--yes)
      assume_yes=1
      ;;
    -r|--resume)
      resume=1
      ;;
    -*)
      options+=("$argument")
      ;;
    *)
      if [ -n "$workspace_argument" ]; then
        error "Solo se puede indicar un directorio de trabajo."
        usage >&2
        exit 2
      fi
      workspace_argument="$argument"
      ;;
  esac
done

if ! iasi_parse_arguments "${options[@]}"; then
  error "Opción no válida: $IASI_ARGUMENT_ERROR"
  usage >&2
  exit 2
fi

if [ "$IASI_HELP" -eq 1 ]; then
  usage
  exit 0
fi

if [ -z "$workspace_argument" ]; then
  workspace_argument="$PWD"
fi

if [ "$resume" -eq 0 ] && [ "$assume_yes" -eq 0 ]; then
  if ! confirm "Los repositorios existentes en $workspace_argument se eliminarán y se clonarán de nuevo."; then
    info "Operación cancelada."
    exit 1
  fi
fi

if [ ! -d "$workspace_argument" ]; then
  if ! mkdir -p -- "$workspace_argument"; then
    error "No se pudo crear el directorio de trabajo: $workspace_argument"
    exit 1
  fi

  info "Directorio de trabajo creado: $workspace_argument"
fi

WORKSPACE_DIR="$(cd -- "$workspace_argument" && pwd)"
LOG_DIR="$WORKSPACE_DIR/logs"
SCRIPT_NAME="$(basename "$0" .sh)"
LOG_FILE="$LOG_DIR/$SCRIPT_NAME-$(date +%Y%m%d%H%M%S).log"

if ! mkdir -p -- "$LOG_DIR"; then
  error "No se pudo crear el directorio de logs: $LOG_DIR"
  exit 1
fi

cd -- "$WORKSPACE_DIR"

{
  printf "IASI clone started at %s\n" "$(date --iso-8601=seconds)"
  printf "Organization: %s\n" "$IASI_ORG"
  printf "Workspace: %s\n\n" "$WORKSPACE_DIR"
} > "$LOG_FILE"

if ! repositories="$(iasi_repositories 2>> "$LOG_FILE")"; then
  error "No se pudo obtener la lista de repositorios de $IASI_ORG."
  detail "Consulta el log: $LOG_FILE"
  exit 1
fi

while IFS='|' read -r name url default_branch; do
  [ -n "$name" ] || continue

  if [[ ! "$name" =~ ^[a-zA-Z0-9._-]+$ ]] || [ "$name" = "." ] || [ "$name" = ".." ]; then
    error "Nombre de repositorio no válido: $name"
    detail "Consulta el log: $LOG_FILE"
    exit 1
  fi

  target="$WORKSPACE_DIR/$name"

  if [ -e "$target" ] || [ -L "$target" ]; then
    if [ "$resume" -eq 1 ]; then
      detail "$name ya existe; se omite."
      continue
    fi

    if ! rm -rf -- "$target" >> "$LOG_FILE" 2>&1; then
      error "No se pudo eliminar $name."
      detail "Consulta el log: $LOG_FILE"
      exit 1
    fi
  fi

  info "Clonando $name."
  detail "$url"
  detail "$target"

  if ! clone_repository "$name" "$url" "$target"; then
    error "No se pudo clonar $name."
    detail "Consulta el log: $LOG_FILE"
    exit 1
  fi

  success_detail "$name clonado."
done <<< "$repositories"

if [ "$resume" -eq 1 ]; then
  success "Repositorios pendientes clonados correctamente."
else
  success "Repositorios recreados correctamente."
fi
detail "Log: $LOG_FILE"
