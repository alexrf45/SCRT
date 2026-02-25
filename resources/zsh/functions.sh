extract() {
  if [ -f $1 ]; then
    case $1 in
    *.tar.bz2) tar xjf $1 ;;
    *.tar.gz) tar xzf $1 ;;
    *.bz2) bunzip2 $1 ;;
    *.rar) rar x $1 ;;
    *.gz) gunzip $1 ;;
    *.tar) tar xf $1 ;;
    *.tbz2) tar xjf $1 ;;
    *.tgz) tar xzf $1 ;;
    *.zip) unzip $1 ;;
    *.Z) uncompress $1 ;;
    *.7z) uncompress $1 ;;
    *.apk) uncompress $1 ;;
    *.bz2) uncompress $1 ;;
    *.tar.gz) uncompress $1 ;;
    *.tar.lz) uncompress $1 ;;
    *.war) uncompress $1 ;;
    *.xz) xz -d -v $1 ;;
    *) echo "'$1' cannot be extracted via extract()" ;;
    esac
  else
    echo "'$1' is not a valid file"
  fi
}

up() {
  for i in $(seq 1 $1); do
    cd ../
  done
}

virtual_env() {
  mkdir $1 &&
    cd $1 &&
    virtualenv .venv &&
    source .venv/bin/activate
}

webserver() {
  miniserve \
    -t "fr3d" \
    -v \
    -F \
    -H \
    -c monokai \
    -p 9443 \
    --header "Cache-Control:no-cache" \
    --header "Server: Server" \
    --header "X-Frame-Options: DENY" \
    --header "Referrer-Policy: strict-origin-when-cross-origin" \
    --header "Content-Type: text/html; charset=UTF-8" \
    --header "Strict-Transport-Security: max-age=63072000; includeSubDomains; preload" \
    --header "Permissions-Policy: geolocation=(), camera=(), microphone=(), bluetooth=(), usb=(), payment=(), display-capture=()" \
    --auth-file ./auth.txt \
    --tls-cert $HOME/.local/$1-cert.pem \
    --tls-key $HOME/.local/$1-key.pem \
    -u
}

cert-gen() {
  mkcert -key-file $HOME/.local/$1-key.pem -cert-file $HOME/.local/$1-cert.pem localhost
}

timestamp() {
  date +%Y%m%d-%T
}

encrypt_age() {
  age \
    --passphrase \
    --output $1.enc \
    $2

}

decrypt_age() {
  age -d \
    $1 >$2
}

mkcd() {
  if [ ! -n "$1" ]; then
    echo "Enter a directory name"
  elif [ -d $1 ]; then
    echo "\`$1' already exists"
  else
    mkdir $1 && cd $1
  fi
}
