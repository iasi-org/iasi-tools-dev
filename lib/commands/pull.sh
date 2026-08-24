#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$(cd -- "$SCRIPT_DIR/../.." && pwd)"

source "$TOOLS_DIR/lib/core/messages.sh"
source "$TOOLS_DIR/lib/core/arguments.sh"
source "$TOOLS_DIR/lib/core/repositories.sh"

usage() {
  cat <<EOF
Usage: iasi-dev pull [options] [repository...]

Synchronizes repositories from $IASI_ORG, always prioritizing the remote state.
The workspace is the directory that contains iasi-tools-dev. Without repository,
all repositories in that workspace are synchronized. With repository names,
only those root-level repository directories are synchronized.

The repository argument is a directory name inside the IASI workspace, not a
path. The project marker is defined by the configured VCS. Git is used by default.

Missing repositories are cloned. Existing repositories are reset to the remote
default branch and untracked files are removed. Command output is written to
logs/iasi-pull-YYYYMMDDhhmmss.log next to the repositories.

Options:
  -h, --help   Show this help
  -v           Detailed information
  -s           Silent mode
  -y, --yes    Do not ask for confirmation
EOF
}

repository_arguments=()
options=()
assume_yes=0

for argument in "$@"; do
  case "$argument" in
    -y|--yes)
      assume_yes=1
      ;;
    -*)
      options+=("$argument")
      ;;
    *)
      repository_arguments+=("$argument")
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

workspace_argument="$(cd -- "$TOOLS_DIR/.." && pwd)"

for selected_repository in "${repository_arguments[@]}"; do
  if [[ ! "$selected_repository" =~ ^[a-zA-Z0-9._-]+$ ]] || \
     [ "$selected_repository" = "." ] || [ "$selected_repository" = ".." ]; then
    error "Indica solo nombres de directorio de repositorio dentro de $(basename -- "$workspace_argument"): $selected_repository"
    exit 2
  fi
done

if [ "$assume_yes" -eq 0 ]; then
  if [ "${#repository_arguments[@]}" -gt 0 ]; then
    confirmation_text="El estado local de ${repository_arguments[*]} se sobrescribirá con la versión remota."
  else
    confirmation_text="Los repositorios locales en $workspace_argument se sobrescribirán con las versiones remotas."
  fi

  if ! confirm "$confirmation_text"; then
    info "Operación cancelada."
    exit 0
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
  printf "IASI pull started at %s\n" "$(date --iso-8601=seconds)"
  printf "Organization: %s\n" "$IASI_ORG"
  printf "Workspace: %s\n" "$WORKSPACE_DIR"
  printf "Repositories: %s\n\n" "${repository_arguments[*]:-all}"
} > "$LOG_FILE"

if ! repositories="$(iasi_repositories 2>> "$LOG_FILE")"; then
  error "No se pudo obtener la lista de repositorios de $IASI_ORG."
  detail "Consulta el log: $LOG_FILE"
  exit 1
fi

found_repositories=()

while IFS='|' read -r name url default_branch; do
  [ -n "$name" ] || continue

  if [ "${#repository_arguments[@]}" -gt 0 ]; then
    selected=0
    for requested in "${repository_arguments[@]}"; do
      [ "$name" = "$requested" ] && selected=1
    done
    [ "$selected" -eq 1 ] || continue
  fi

  found_repositories+=("$name")

  if [[ ! "$name" =~ ^[a-zA-Z0-9._-]+$ ]] || [ "$name" = "." ] || [ "$name" = ".." ]; then
    error "Nombre de repositorio no válido: $name"
    detail "Consulta el log: $LOG_FILE"
    exit 1
  fi

  if [ -z "$default_branch" ]; then
    error "GitHub no indicó la rama por defecto de $name."
    detail "Consulta el log: $LOG_FILE"
    exit 1
  fi

  target="$WORKSPACE_DIR/$name"

  if [ ! -e "$target" ] && [ ! -L "$target" ]; then
    info "Clonando $name."
    detail "$url"
    detail "$target"

    if ! vcs_clone "$url" "$target" >> "$LOG_FILE" 2>&1; then
      error "No se pudo clonar $name."
      detail "Consulta el log: $LOG_FILE"
      exit 1
    fi
  elif ! is_project_root "$target"; then
    info "Sustituyendo $name por el repositorio remoto."
    detail "$target"

    if ! rm -rf -- "$target" >> "$LOG_FILE" 2>&1 || \
       ! vcs_clone "$url" "$target" >> "$LOG_FILE" 2>&1; then
      error "No se pudo recrear $name."
      detail "Consulta el log: $LOG_FILE"
      exit 1
    fi
  else
    info "Sincronizando $name con origin/$default_branch."
    detail "$target"

    if ! vcs_configure_remote "$target" "origin" "$url" >> "$LOG_FILE" 2>&1; then
      error "No se pudo configurar origin para $name."
      detail "Consulta el log: $LOG_FILE"
      exit 1
    fi

    if ! vcs_sync_to_remote "$target" "origin" "$default_branch" >> "$LOG_FILE" 2>&1; then
      error "No se pudo sincronizar $name."
      detail "Consulta el log: $LOG_FILE"
      exit 1
    fi
  fi

  success_detail "$name sincronizado."
done <<< "$repositories"

for requested in "${repository_arguments[@]}"; do
  found=0
  for name in "${found_repositories[@]}"; do
    [ "$requested" = "$name" ] && found=1
  done
  if [ "$found" -eq 0 ]; then
    error "$requested no pertenece a $IASI_ORG o está archivado."
    detail "Consulta el log: $LOG_FILE"
    exit 1
  fi
done

success "Repositorios sincronizados correctamente."
detail "Log: $LOG_FILE"
