#!/bin/bash

if ! command -v docker >/dev/null 2>&1; then
    info "> Installing Docker."
    curl -sSL https://get.docker.com/ | /bin/sh
    info ""

    info "> Attempting to start Docker daemon."
    # todo: find a better way to ignore service already started error
    if [[ $INIT_SYSTEM = "systemd" ]]; then
        if ! sudo systemctl start docker.service; then
            :
        fi
    else
        if ! sudo service docker start; then
            :
        fi
    fi
    info ""
fi
