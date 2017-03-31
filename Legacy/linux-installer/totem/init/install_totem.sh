#!/bin/bash

function install_initscript () {
    template=$(cat "$template" | sed -e 's~INSTALL_DIRECTORY~'"${settings[totem.path]}"'~')
    echo "$template" | sudo tee "$initpath" >/dev/null
    enable_script
}

if [[ ${settings[os.initsystem]} = "init" ]]; then
    template="${settings[root]}/totem/init/upstart.totem.template"
    initpath="/etc/init/holmes-totem.conf"
    function enable_script () {
        sudo initctl reload-configuration
        sudo service holmes-totem start
    }
elif [[ ${settings[os.initsystem]} = "systemd" ]]; then
    template="${settings[root]}/totem/init/systemd.totem.template"
    initpath="/etc/systemd/system/holmes-totem.service"
    function enable_script () {
        sudo systemctl enable holmes-totem.service
        sudo systemctl start holmes-totem.service
    }
else
    error "> Unknown init system ${settings[os.initsystem]}, failed to install totem as a system service."
    function install_initscript () {
        :
    }
fi

install_initscript
