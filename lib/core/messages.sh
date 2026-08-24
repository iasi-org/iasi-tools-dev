#!/usr/bin/env bash

# -----------------------------------------------------------------------------
# IASI Tools - Core messages
# -----------------------------------------------------------------------------

IASI_VERBOSITY="${IASI_VERBOSITY:-1}"

# Colors
_IASI_BOLD="\033[1m"
_IASI_GREEN="\033[0;32m"
_IASI_CYAN="\033[0;36m"
_IASI_YELLOW="\033[0;33m"
_IASI_RED="\033[0;31m"
_IASI_BOLD_RED="\033[1;31m"
_IASI_RESET="\033[0m"

_message() {
  local color="$1"
  local text="$2"

  printf "%b%s - %s%b\n" "$color" "$(date +%T)" "$text" "$_IASI_RESET"
}

info() {
  [ "$IASI_VERBOSITY" -ge 1 ] || return 0
  _message "$_IASI_BOLD" "$1"
}

detail() {
  [ "$IASI_VERBOSITY" -ge 2 ] || return 0
  _message "$_IASI_BOLD" "$1"
}

info_detail() {
  [ "$IASI_VERBOSITY" -ge 2 ] || return 0
  _message "$_IASI_CYAN" "$1"
}

success() {
  [ "$IASI_VERBOSITY" -ge 2 ] || return 0
  _message "$_IASI_GREEN" "$1"
}

success_detail() {
  [ "$IASI_VERBOSITY" -ge 2 ] || return 0
  _message "$_IASI_GREEN" "$1"
}

warning() {
  [ "$IASI_VERBOSITY" -ge 1 ] || return 0
  _message "$_IASI_YELLOW" "$1"
}

confirm() {
  local text="$1"
  local response=""

  printf "%b%s (Y/N) %b" "$_IASI_BOLD_RED" "$text" "$_IASI_RESET"
  read -r response || return 1

  [[ "$response" =~ ^[Yy]$ ]]
}

error() {
  [ "$IASI_VERBOSITY" -ge 1 ] || return 0
  _message "$_IASI_RED" "$1" >&2
}
