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
the normal commit and publish flow. Other repositories keep the normal VCS flow.
Git is the default VCS.

Arguments:
  project      Optional project or workspace directory; current directory
               by default

Options:
  -f, --full   Run each project deploy before committing and pushing
  -m, --message MESSAGE
               Base commit message; "deploy" by default
  -v           Show detailed operational and success messages
  -s           Silent mode; show no messages
  -h, --help   Show this help
EOF
}

full=0
commit_message="deploy"
silent=0
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

export IASI_VERBOSITY

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

repositories=()

if is_project_root "$TARGET_DIR"; then
  repositories+=("$TARGET_DIR")
else
  for candidate in "$TARGET_DIR"/*; do
    is_project_root "$candidate" || continue
    repositories+=("$candidate")
  done
fi

if [ "${#repositories[@]}" -eq 0 ]; then
  error "No se encontraron proyectos en $TARGET_DIR (marcador: $IASI_PROJECT_MARKER)."
  exit 1
fi

if is_project_root "$TARGET_DIR"; then
  LOG_DIR="$(dirname -- "$TARGET_DIR")/logs"
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
  info "Desplegando $name."
  {
    printf "============================================================\n"
    printf "PROJECT: %s\n" "$name"
    printf "============================================================\n"
  } >> "$LOG_FILE"

  if [ "$full" -eq 1 ] && repository_has_iasi_project "$repository"; then

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

  if ! vcs_stage_all "$repository" >> "$LOG_FILE" 2>&1; then
    printf "PROJECT RESULT: FAILED (vcs stage)\n\n" >> "$LOG_FILE"
    error "No se pudieron preparar los cambios de $name."
    warning "Consulta el log: $LOG_FILE"
    exit 1
  fi

  project_changed=0
  project_pushed=0

  if vcs_has_changes "$repository" >> "$LOG_FILE" 2>&1; then
    project_changed=1
    info_detail "Commit $name."
    if ! vcs_commit "$repository" "$commit_message" >> "$LOG_FILE" 2>&1; then
      printf "PROJECT RESULT: FAILED (vcs commit)\n\n" >> "$LOG_FILE"
      error "No se pudo crear el commit de $name."
      warning "Consulta el log: $LOG_FILE"
      exit 1
    fi
    commits=$((commits + 1))
  fi

  if vcs_has_outgoing "$repository" >> "$LOG_FILE" 2>&1; then
    info_detail "Push $name."
    if ! vcs_push "$repository" >> "$LOG_FILE" 2>&1; then
      printf "PROJECT RESULT: FAILED (vcs publish)\n\n" >> "$LOG_FILE"
      error "No se pudieron subir los commits de $name."
      warning "Consulta el log: $LOG_FILE"
      exit 1
    fi
    project_pushed=1
  fi

  if [ "$project_changed" -eq 0 ] && [ "$project_pushed" -eq 0 ]; then
    info_detail "No tiene cambios."
  else
    success_detail "Desplegado."
  fi

  printf "PROJECT RESULT: OK\n\n" >> "$LOG_FILE"
done
