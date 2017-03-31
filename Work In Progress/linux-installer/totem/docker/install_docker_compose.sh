#!/bin/bash

if ! command -v "docker-compose" >/dev/null 2>&1; then
    info "> Installing Docker-Compose."
    sudo sh -c "curl -L https://github.com/docker/compose/releases/download/1.7.0/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose"
    sudo chmod +x /usr/local/bin/docker-compose
    info ""
fi
