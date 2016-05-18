#!/bin/bash

if [[ -f "${settings[storage.path]}/uninstall.sh" ]]; then
    cd "${settings[storage.path]}"
    (uninstall.sh --keep-go --remove-data)
    cd "${settings[root]}"
fi
sudo rm -rf "${settings[storage.path]}" &>/dev/null
