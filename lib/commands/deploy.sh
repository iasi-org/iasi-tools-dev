#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$(cd -- "$SCRIPT_DIR/../.." && pwd)"

source "$TOOLS_DIR/lib/core/messages.sh"
source "$TOOLS_DIR/lib/core/repositories.sh"

usage() {
  cat <<'EOF'
Usage: iasi-dev deploy [-f|--full] [-m|--message "message"] [project...]

Commits the current project state and pushes it. With --full, repositories
managed by IASI Quarto (`_iasi.yml`) run their own incremental deploy before
the normal commit and push. Other repositories keep the normal Git flow.

Arguments:
  project      Optional project or workspace directory; current directory
               by default

Options:
  -f, --full   Run each project deploy before committing and pushing
  -m, --message MESSAGE
               Base commit message; "deploy" by default
  -v           Show detailed information, including success messages
  -h, --help   Show this help
EOF
}

full=0
commit_message="deploy"
target_arguments=()

while [ "$#" -gt 0 ]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    -f|--full)
      full=1
      shift
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
      IASI_VERBOSITY=2
      shift
      ;;
    --message=*)
      commit_message="${1#*=}"
      if [ -z "$commit_message" ]; then
        error "La opción --message requiere un mensaje."
        exit 2
      fi
      shift
      ;;
    -*)
      error "Opción desconocida: $1"
      usage >&2
      exit 2
      ;;
    *)
      target_arguments+=("$1")
      shift
      ;;
  esac
done

if [ "${#target_arguments[@]}" -gt 1 ]; then
  for selected_target in "${target_arguments[@]}"; do
    deploy_options=(-m "$commit_message")
    [ "$full" -eq 1 ] && deploy_options=(-f "${deploy_options[@]}")
    "$0" "${deploy_options[@]}" "$selected_target"
  done
  exit 0
fi

target_argument="${target_arguments[0]:-}"
target="${target_argument:-$PWD}"

if [ ! -d "$target" ]; then
  error "No existe el directorio: $target"
  exit 2
fi

TARGET_DIR="$(cd -- "$target" && pwd)"

if [ -n "$target_argument" ]; then
  info "Desplegando $(basename -- "$TARGET_DIR")."
else
  info "Desplegando todos los proyectos."
fi

repositories=()

if repository_root="$(git -C "$TARGET_DIR" rev-parse --show-toplevel 2>/dev/null)"; then
  repositories+=("$repository_root")
else
  for candidate in "$TARGET_DIR"/*; do
    [ -d "$candidate/.git" ] || continue
    repositories+=("$candidate")
  done
fi

if [ "${#repositories[@]}" -eq 0 ]; then
  error "No se encontraron repositorios Git en $TARGET_DIR."
  exit 1
fi

if target_repository="$(git -C "$TARGET_DIR" rev-parse --show-toplevel 2>/dev/null)"; then
  LOG_DIR="$(dirname -- "$target_repository")/logs"
else
  LOG_DIR="$TARGET_DIR/logs"
fi
LOG_FILE="$LOG_DIR/iasi-deploy-$(date +%Y%m%d%H%M%S).log"
mkdir -p -- "$LOG_DIR"

{
  printf "IASI deploy started at %s\n" "$(date --iso-8601=seconds)"
  printf "Directory: %s\n" "$TARGET_DIR"
  printf "Full: %s\n" "$full"
  printf "Message: %s\n\n" "$commit_message"
} > "$LOG_FILE"

commits=0

for repository in "${repositories[@]}"; do
  name="$(basename -- "$repository")"
  {
    printf "============================================================\n"
    printf "PROJECT: %s\n" "$name"
    printf "============================================================\n"
  } >> "$LOG_FILE"

  if [ "$full" -eq 1 ] && repository_has_iasi_project "$repository"; then
    info "Ejecutando deploy de $name."

    if IASI_QUARTO_OPERATION=deploy \
      IASI_LOG_FILE="$LOG_FILE" \
      "$TOOLS_DIR/lib/commands/publish.sh" "$repository"; then
      :
    else
      deploy_rc=$?
      printf "PROJECT RESULT: FAILED (exit %s)\n\n" "$deploy_rc" >> "$LOG_FILE"
      error "Falló el deploy de $name."
      warning "Consulta el log: $LOG_FILE"
      exit "$deploy_rc"
    fi
  fi

  if ! git -C "$repository" add -A -- . >> "$LOG_FILE" 2>&1; then
    printf "PROJECT RESULT: FAILED (git add)\n\n" >> "$LOG_FILE"
    error "No se pudieron preparar los cambios de $name."
    warning "Consulta el log: $LOG_FILE"
    exit 1
  fi

  if git -C "$repository" diff --cached --quiet >> "$LOG_FILE" 2>&1; then
    :
  else
    if ! git -C "$repository" commit -m "$commit_message" >> "$LOG_FILE" 2>&1; then
      printf "PROJECT RESULT: FAILED (git commit)\n\n" >> "$LOG_FILE"
      error "No se pudo crear el commit de $name."
      warning "Consulta el log: $LOG_FILE"
      exit 1
    fi
    commits=$((commits + 1))
  fi

  if ! git -C "$repository" push >> "$LOG_FILE" 2>&1; then
    printf "PROJECT RESULT: FAILED (git push)\n\n" >> "$LOG_FILE"
    error "No se pudieron subir los commits de $name."
    warning "Consulta el log: $LOG_FILE"
    exit 1
  fi

  printf "PROJECT RESULT: OK\n\n" >> "$LOG_FILE"
done

success "$commits commit(s) creados."
