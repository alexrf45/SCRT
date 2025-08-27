#!/bin/bash

TOOL_DIR="/home/kali/resources"

mkdir -p "$HOME/.config" && mkdir -p "$HOME/.zsh" &&
  cp -r "$TOOL_DIR"/tmux.conf "$HOME/.tmux.conf" &&
  cp -r "$TOOL_DIR"/shell-upgrade.sh "$HOME/.tools/shell-upgrade.sh" &&
  # cp -r "$TOOL_DIR"/recon.sh "$HOME/.local/bin/recon.sh" && chmod +x "$HOME/.local/bin/recon.sh" &&
  cp -r "$TOOL_DIR"/proxychains.conf "$HOME/.proxychains/proxychains.conf" &&
  cp -r "$TOOL_DIR"/kerbrute "$HOME/.local/bin/kerbrute" && chmod +x "$HOME/.local/bin/kerbrute" &&
  cp -r "$TOOL_DIR"/smbserver.py "$HOME/.tools/smbserver.py" &&
  cp -r "$TOOL_DIR"/zsh/history "$HOME/.history" &&
  cp -r "$TOOL_DIR"/mozilla "$HOME/.mozilla" &&
  cp -r "$TOOL_DIR"/starship.toml "$HOME/.config/starship.toml"

git clone https://github.com/tmux-plugins/tpm "$HOME/.tmux/plugins/tpm"

git clone https://github.com/LazyVim/starter "$HOME/.config/nvim"

rm -rf ~/.config/nvim/.git

cp -r "$TOOL_DIR"/plugins/* "$HOME/.config/nvim/lua/plugins/."

sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended

git clone https://github.com/zsh-users/zsh-autosuggestions "$HOME/.oh-my-zsh/custom/plugins/zsh-autosuggestions"

git clone https://github.com/zsh-users/zsh-syntax-highlighting.git "$HOME/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting"

cp "$TOOL_DIR"/zsh/.zshrc "$HOME/.zshrc"

cp "$TOOL_DIR"/zsh/kali.zsh-theme "$HOME/.oh-my-zsh/custom/themes/kali.zsh-theme"

cp "$TOOL_DIR"/zsh/functions.sh "$HOME/.zsh/functions.sh"

cp "$TOOL_DIR"/zsh/aliases "$HOME/.zsh/aliases"

cp "$TOOL_DIR"/zsh/.zprofile "$HOME/.zprofile"
