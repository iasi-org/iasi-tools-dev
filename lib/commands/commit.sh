#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$(cd -- "$SCRIPT_DIR/../.." && pwd)"

source "$TOOLS_DIR/lib/core/messages.sh"
source "$TOOLS_DIR/lib/core/repositories.sh"

usage() {
  cat <<'EOF'
Usage: iasi-dev commit -m "message" [project...]

Stages all changes, creates a commit, and pushes it to the configured remote.
Without a repository, Git repositories directly below the current directory are
selected.

Arguments:
  project      Optional Git repository directory or path

Options:
  -m, --message MESSAGE
               Required commit message
  -v           Show detailed operational and success messages
  -s           Silent mode; show no messages
  -h, --help   Show this help
EOF
}

commit_message=""
silent=0
repository_arguments=()

while [ "$#" -gt 0 ]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    -m|--message)
      if [ "$#" -lt 2 ] || [ -z "$2" ]; then
        error "La opción $1 requiere un mensaje."
        exit 2
      fi
      commit_message="$2"
      shift 2
      ;;
    -v)
      if [ "$silent" -eq 0 ]; then
        IASI_VERBOSITY=2
      fi
      shift
      ;;
    -s)
      silent=1
      IASI_VERBOSITY=0
      shift
      ;;
    --message=*)
      commit_message="${1#*=}"
      shift
      ;;
    -*)
      error "Opción desconocida: $1"
      usage >&2
      exit 2
      ;;
    *)
      repository_arguments+=("$1")
      shift
      ;;
  esac
done

export IASI_VERBOSITY

if [ -z "$commit_message" ]; then
  error "El mensaje del commit es obligatorio; usa -m o --message."
  usage >&2
  exit 2
fi

repositories=()

if [ "${#repository_arguments[@]}" -gt 0 ]; then
 for repository_argument in "${repository_arguments[@]}"; do
  if [ ! -d "$repository_argument" ]; then
    error "No existe el directorio de repositorio: $repository_argument"
    exit 2
  fi

  repository_path="$(cd -- "$repository_argument" && pwd)"

  if [ ! -d "$repository_path/.git" ]; then
    error "El directorio no es un repositorio Git: $repository_argument"
    exit 2
  fi

  repositories+=("$repository_path")
  workspace_dir="$(dirname -- "$repository_path")"
 done
else
  workspace_dir="$PWD"

  for candidate in "$workspace_dir"/*; do
    [ -d "$candidate/.git" ] || continue
    repositories+=("$candidate")
  done

  if [ "${#repositories[@]}" -eq 0 ]; then
    error "No se encontraron repositorios Git en $workspace_dir."
    exit 1
  fi
fi

log_dir="$workspace_dir/logs"
log_file="$log_dir/iasi-commit-$(date +%Y%m%d%H%M%S).log"

if ! mkdir -p -- "$log_dir"; then
  error "No se pudo crear el directorio de logs: $log_dir"
  exit 1
fi

{
  printf "IASI commit started at %s\n" "$(date --iso-8601=seconds)"
  printf "Workspace: %s\n" "$workspace_dir"
  printf "Projects: %s\n" "${repository_arguments[*]:-all}"
  printf "Message: %s\n\n" "$commit_message"
} > "$log_file"

committed=0

for repository in "${repositories[@]}"; do
  name="$(basename -- "$repository")"
  info "Procesando $name."

  printf "[%s]\n" "$name" >> "$log_file"

  if ! git -C "$repository" add -A . >> "$log_file" 2>&1; then
    error "No se pudieron preparar los cambios de $name."
    warning "Consulta el log: $log_file"
    exit 1
  fi

  if git -C "$repository" diff --cached --quiet >> "$log_file" 2>&1; then
    detail "$name no tiene cambios."
  else
    if ! git -C "$repository" commit -m "$commit_message" >> "$log_file" 2>&1; then
      error "No se pudo crear el commit de $name."
      warning "Consulta el log: $log_file"
      exit 1
    fi

    committed=$((committed + 1))
  fi

  if ! git -C "$repository" push >> "$log_file" 2>&1; then
    error "No se pudo publicar el commit de $name."
    warning "Consulta el log: $log_file"
    exit 1
  fi

  printf "\n" >> "$log_file"
  success_detail "$name desplegado."
done

success "$committed proyecto(s) confirmado(s) y publicado(s)."
