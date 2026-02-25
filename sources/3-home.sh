#!/bin/bash
set -euo pipefail

TOOL_DIR="$HOME/resources"

mkdir -p "$HOME/.config"
mkdir -p "$HOME/.zsh"

cp "$TOOL_DIR/tmux.conf" "$HOME/.tmux.conf"
cp "$TOOL_DIR/shell-upgrade.sh" "$HOME/.tools/shell-upgrade.sh"
cp "$TOOL_DIR/proxychains.conf" "$HOME/.proxychains/proxychains.conf"
cp "$TOOL_DIR/kerbrute" "$HOME/.local/bin/kerbrute"
chmod +x "$HOME/.local/bin/kerbrute"
cp "$TOOL_DIR/smbserver.py" "$HOME/.tools/smbserver.py"
cp "$TOOL_DIR/zsh/history" "$HOME/.history"
cp -r "$TOOL_DIR/mozilla" "$HOME/.mozilla"
cp "$TOOL_DIR/starship.toml" "$HOME/.config/starship.toml"

mkdir -p "$HOME/.BurpSuite"
cp "$TOOL_DIR/burp-user.json" "$HOME/.BurpSuite/UserConfigCommunity.json"

git clone --depth 1 https://github.com/tmux-plugins/tpm "$HOME/.tmux/plugins/tpm"

sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended

git clone --depth 1 https://github.com/zsh-users/zsh-autosuggestions "$HOME/.oh-my-zsh/custom/plugins/zsh-autosuggestions"

git clone --depth 1 https://github.com/zsh-users/zsh-syntax-highlighting.git "$HOME/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting"

cp "$TOOL_DIR/zsh/.zshrc" "$HOME/.zshrc"
cp "$TOOL_DIR/zsh/kali.zsh-theme" "$HOME/.oh-my-zsh/custom/themes/kali.zsh-theme"
cp "$TOOL_DIR/zsh/functions.sh" "$HOME/.zsh/functions.sh"
cp "$TOOL_DIR/zsh/aliases" "$HOME/.zsh/aliases"
cp "$TOOL_DIR/zsh/burp.env" "$HOME/.zsh/burp.env"
cp "$TOOL_DIR/zsh/.zprofile" "$HOME/.zprofile"
