#!/bin/bash

function command_exists () {
    if command -v "$1" >/dev/null 2>&1; then
        echo "true"
        return 0
    else
        return 1
    fi
}

function ensure () {
    # $1 = cmd name
    # $2 = package to install if not exists
    cmd="$1"
    pkg="$2"
    if [ ! $(command_exists "$cmd") ]; then
        if [ ! $(command_exists "apt-get") ]; then
            error "> Fatal: apt-get not found, please install to proceed."
            exit 1
        fi
        sudo apt-get update
        sudo apt-get install -y "$pkg"
        if [ ! $(command_exists "$cmd") ]; then
            error "> Fatal: '$cmd' not available and 'sudo apt-get install $pkg' seems to have failed."
            exit 1
        fi
    fi
}

function ensure_lsb_release () {
    ensure lsb_release lsb-release
}

function ensure_curl () {
    ensure curl curl
}

function ensure_wget () {
    ensure wget wget
}

function ensure_add_apt_repository () {
    # special treatment for ubuntu < 12.10
    if [[ ${settings[os]} = "Ubuntu" ]]; then
        if [[ ${settings[os.version.major]} -eq 12 && ${settings[os.version.minor]} -lt 10 ]] ||
           [[ ${settings[os.version.major]} -lt 12 ]]; then
            
            ensure add-apt-repository python-software-properties
            return
        fi
    fi
    ensure add-apt-repository software-properties-common
}

function ensure_apt_https () {
    if [[ -f /usr/lib/apt/methods/https ]]; then
        return
    fi
    sudo apt-get update
    sudo apt-get install -y --force-yes apt-transport-https
}
