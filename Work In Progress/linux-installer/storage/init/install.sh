#!/bin/bash

info "> ${settings[os.initsystem]}"

if [[ ${settings[os.initsystem]} = "init" ]]; then
    template="storage/init/upstart.storage.template"
    initpath="/etc/init/holmes-storage.conf"
    function enable_script {
        sudo initctl reload-configuration
        sudo service holmes-storage start
    }
else
    template="storage/init/systemd.storage.template"
    initpath="/etc/systemd/system/holmes-storage.service"
    function enable_script {
        sudo systemctl enable holmes-storage.service
        sudo systemctl start holmes-storage.service
    }
fi

template=$(cat "$template" | sed -e 's~INSTALL_DIRECTORY~'"${settings[storage.path]}"'~')
template=$(echo "$template" | sed -e 's~CONFIG_FILE~'"${settings[storage.path]}"/config.json'~')
echo "$template" | sudo tee "$initpath" >/dev/null
enable_script
