#!/bin/bash

function install_prerequisites () {
    if ! resume_at "storage.prerequisites"; then
        return
    fi
    checkpoint "storage.prerequisites"
    
    info "> Installing prerequisites."
    sudo apt-get update
    sudo apt-get install -y curl libmagic-dev gcc
    ensure_apt_https
}

function install_go () {
    if ! resume_at "storage.install_go"; then
        return
    fi
    checkpoint "storage.install_go"
    
    info "> Installing Golang."
    if ! command -v go >/dev/null 2>&1; then  # if go can't be found, attempt to fix eventually broken ENV
        export PATH=$PATH:/usr/local/go/bin
    fi
    if command -v go >/dev/null 2>&1; then
        version=$(go version | awk '{ print $3 }' | sed -e 's/^go//')
        _ifs=IFS
        IFS='.' read -r -a golang_version <<< "$version"
        IFS=_ifs
        if [[ golang_version[0] -gt 1 ]] ||
           [[ golang_version[0] -eq 1 && golang_version[1] -ge 6 ]]; then
           info "> Golang version greater or equal to 1.6, that's sufficient"
            return
        fi
    fi
    # https://golang.org/doc/install
    url="https://storage.googleapis.com/golang/go1.6.1.linux-amd64.tar.gz"
    curl -o /tmp/go.tar.gz -L "$url"
    if [[ -d /usr/local/go && $(ls -A /usr/local/go) ]]; then
        sudo rm -rf /usr/local/go/*
    fi
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
}

function install_configuration () {
    if [ ${settings[storage.config-url]} ]; then
        info "> Downloading configuration file"
        ensure_curl
        curl -o "config.json" -L "${settings[storage.config-url]}"
    else
        if [[ "${settings[storage.config-create]}" != "" ]]; then
            go run "${settings[root]}/storage/config_helper.go" --config="${settings[storage.config-create]}"
            # ugly construct to avoid "same file" error of move
            mv config.json "${settings[storage.path]}/config.json2"
            mv "${settings[storage.path]}/config.json2" "${settings[storage.path]}/config.json"
        fi
    fi
}

function install_storage () {
    if resume_at "storage.prepare"; then
        checkpoint "storage.prepare"
        
        info "> Preparing installation directory."
        empty_install_path "storage"
        create_user "totem"
        sudo mkdir -p "${settings[storage.path]}"
        cd "${settings[storage.path]}"
        sudo chown -R "$USER":"$USER" "${settings[storage.path]}"
    fi
    
    # build
    info "> Setting up golang environment variables."
    export GOPATH="$HOME/go"
    export PATH=$PATH:/usr/local/go/bin
    export GOROOT=/usr/local/go
    mkdir -p "$GOPATH"
    
    if resume_at "storage.build"; then
        checkpoint "storage.build"
        
        info "> Building Holmes-Storage, this may take a while, please be patient!"
        echo "go get -v -x -u '${settings[storage.repo]}'"
        go get -v -x -u "${settings[storage.repo]}"
    fi
    
    # grab/create config
    if resume_at "storage.config"; then
        checkpoint "storage.config"
        install_configuration
    fi
    
    # install
    if resume_at "storage.install"; then
        checkpoint "storage.install"
        cp "$HOME/go/bin/Holmes-Storage" "${settings[storage.path]}/Holmes-Storage"
        sudo chown -R totem:totem "${settings[storage.path]}"
    fi
    
    # db setup
    if resume_at "storage.db_setup"; then
        checkpoint "storage.db_setup"
        if [ ! ${settings[storage.skip-setup]} ]; then
            sudo su totem -c "cd ${settings[storage.path]} && ./Holmes-Storage --setup --objSetup"
        fi
    fi
    
    # init scripts
    if resume_at "storage.initscripts"; then
        checkpoint "storage.initscripts"
        if [ ${settings[storage.initscript]} ]; then
            cd "${settings[root]}"
            . storage/init/install.sh
        fi
    fi
}

function main () {
    install_prerequisites
    install_go
    install_storage
}

main
