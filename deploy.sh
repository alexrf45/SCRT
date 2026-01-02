#!/bin/bash

set -e

echo -e "pulling image now..."

docker pull fonalex45/toolkit:dev

mkdir -p "$HOME/.local/bin"

cp script/toolkit "$HOME/.local/bin/."

setup_toolkit() {
  local toolkit_path="$HOME/.local/bin/toolkit"

  # Check if toolkit exists
  if [[ ! -f "$toolkit_path" ]]; then
    echo "Error: Toolkit not found at $toolkit_path"
    return 1
  fi

  # Detect current shell and set appropriate rc file
  if [[ "$SHELL" == *"zsh"* ]]; then
    local rc_file="$HOME/.zshrc"
  else
    local rc_file="$HOME/.bashrc"
  fi

  # Add source line if not already present
  if ! grep -q "\.local/bin/toolkit" "$rc_file" 2>/dev/null; then
    echo "source '$HOME/.local/bin/toolkit'" >>"$rc_file"
    echo "Added toolkit to $rc_file"
  else
    echo "Toolkit already configured in $rc_file"
  fi

  # Source the updated file
  source "$rc_file"
  echo "Sourced $rc_file"
}

setup_toolkit
