#!/bin/bash

function get_initsystem {
    local init=$(cat /proc/1/comm)
    if [[ $init != "systemd" && $init != "init" ]]; then
        info "> Unknown INIT system (neither systemd, nor upstart, but rather reporting ${init})"
    else
        info "> Init system is ${init}"
        settings['os.initsystem']="${init}"
    fi
}

function create_user {
    if id -u "$1" >/dev/null 2>&1; then
        info "> User '$1' already exists, re-using it."
    else
        info "> Creating user '$1'."
        sudo useradd "$1"
        # just useradd is not enough, sbt will crash because it can't
        # create its cash in ~/.sbt
        sudo /sbin/mkhomedir_helper "$1"
    fi
    info ""
}

function empty_install_path {
    # $1: totem/storage
    path="${settings[$1.path]}"
    root="${settings[root]}"
    
    if [[ "$path" != "$root" ]] &&
       [[ ! ${settings[$1.erase]} ]] &&
       [[ -d "$path" && "$(ls -A "$path")" ]]; then
        
        error "> Fatal: Folder '$path' exists and is not empty."
        exit 1
    else
        if [[ "$path" != "$root" ]] &&
           [[ -d "$path" && "$(ls -A "$path")" ]]; then
            
            . "$root/$1/uninstall.sh"
        fi
    fi
}

function resume_at () {
    if [ ! ${settings[resume]} ] || [[ "${settings[resume.checkpoint]}" = "" ]]
    then
        return 0
    fi
    
    checkpoint="${settings[resume.checkpoint]}"
    major=$(echo "$checkpoint" | sed -e "s/\..*//")
    
    if [[ "$checkpoint" = "$1" ]]; then
        # reset our resume
        settings[resume]=
        return 0
    elif [[ "$major" = "$1" ]]; then
        return 0
    else
        return 1
    fi
}
function checkpoint () {
    echo "$1" > "${settings[root]}/universal-installer.checkpoint"
}
