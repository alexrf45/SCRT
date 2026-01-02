setopt AUTO_CD
setopt AUTO_PUSHD
setopt EXTENDED_GLOB
setopt EXTENDED_HISTORY
setopt NOMATCH
setopt MENU_COMPLETE
setopt GLOB_DOTS
setopt INTERACTIVE_COMMENTS
setopt HIST_IGNORE_SPACE  # Don't save when prefixed with space
setopt HIST_IGNORE_DUPS # Don't save duplicate lines
setopt HIST_VERIFY
setopt INC_APPEND_HISTORY
setopt SHARE_HISTORY      # Share history between sessions
setopt PROMPT_SUBST
unsetopt beep

#history config
HISTFILE=~/.history
HISTSIZE=10000
SAVEHIST=10000
HIST_STAMPS="mm/dd/yyyy"

ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE="fg=#ffffff,standout"
ZSH_AUTOSUGGEST_BUFFER_MAX_SIZE="20"
ZSH_AUTOSUGGEST_USE_ASYNC=1

#vi key bindings
bindkey -v

ZSH_THEME="kali"

zstyle ':omz:update' mode auto      # update automatically without asking

#source aliases and env
source "$HOME/.zprofile"

for file in $HOME/.zsh/*; do
    source "$file"
done

autoload -Uz colors; colors
autoload -Uz compinit && compinit

fpath=(/tmp/zsh-completions/src $fpath)

plugins=(
colorize
python
ssh
tmux
git
zsh-autosuggestions
zsh-syntax-highlighting
)

source $HOME/.oh-my-zsh/oh-my-zsh.sh

#displays saying in every new prompt
source <(fzf --zsh)

#persistant ssh agent
#eval $(ssh-agent) &> /dev/null

#starship prompt
eval "$(starship init zsh)"

