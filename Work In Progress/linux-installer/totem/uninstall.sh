#!/bin/bash

if [[ -f "${settings[totem.path]}/uninstall.sh" ]]; then
    cd "${settings[totem.path]}"
    (uninstall.sh --keep-docker --remove-data --keep-sbt --keep-java8)
    cd "${settings[root]}"
fi
sudo rm -rf "${settings[totem.path]}" &>/dev/null
