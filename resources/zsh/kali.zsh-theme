
# #PROMPT='$fg_bold[blue][ $fg[green]$(date +"%a %b %d %Y %l:%M%p") $fg_bold[blue]] $fg_bold[blue] [ $fg[green]%n@%m:%~$(git_prompt_info)$fg[yellow]$(ruby_prompt_info)$fg_bold[blue] ]$reset_color
# # $ '
#
# PROMPT='$fg_bold[yellow][project: $NAME] $fg_bold[blue][ $fg[green]$(date +"%a %b %d %Y %l:%M%p") $fg_bold[blue]] $fg_bold[blue] [ $fg[green]%n@%m:%~$(git_prompt_info)$fg[yellow]$(ruby_prompt_info)$fg_bold[blue] ]$reset_color
#  $ '
#
# # git theming
# ZSH_THEME_GIT_PROMPT_PREFIX="$fg_bold[green]("
# ZSH_THEME_GIT_PROMPT_SUFFIX=")"
# ZSH_THEME_GIT_PROMPT_CLEAN="✔"
# ZSH_THEME_GIT_PROMPT_DIRTY="✗"
# Function to get VPN IP address
function vpn_ip() {
    # Try common VPN interfaces (tun0, tun1, wg0, etc.)
    local vpn_ip=""
    for interface in tun0 tun1 wg0 wg1 ppp0; do
        vpn_ip=$(ip addr show $interface 2>/dev/null | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1)
        if [[ -n "$vpn_ip" ]]; then
            echo "$vpn_ip"
            return
        fi
    done
    
    # Alternative method: check for any tun/tap interfaces
    vpn_ip=$(ip addr show | grep -E '^[0-9]+: (tun|tap|wg)' -A 2 | grep 'inet ' | head -1 | awk '{print $2}' | cut -d'/' -f1)
    if [[ -n "$vpn_ip" ]]; then
        echo "$vpn_ip"
    else
        echo "No VPN"
    fi
}

PROMPT='$fg_bold[yellow][project: $NAME] $fg_bold[red][VPN: $(vpn_ip)] $fg_bold[blue][ $fg[green]$(date +"%a %b %d %Y %l:%M%p") $fg_bold[blue]] $fg_bold[blue] [ $fg[green]%n@%m:%~$(git_prompt_info)$fg[yellow]$(ruby_prompt_info)$fg_bold[blue] ]$reset_color
 $ '

# git theming
ZSH_THEME_GIT_PROMPT_PREFIX="$fg_bold[green]("
ZSH_THEME_GIT_PROMPT_SUFFIX=")"
ZSH_THEME_GIT_PROMPT_CLEAN="✔"
ZSH_THEME_GIT_PROMPT_DIRTY="✗"
