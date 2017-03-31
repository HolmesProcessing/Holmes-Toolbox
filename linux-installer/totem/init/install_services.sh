#!/bin/bash

function install_initscript () {
    template=$(cat "$template" | sed -e 's~INSTALL_DIRECTORY~'"${settings[totem.path]}"'~')
    echo "$template" | sudo tee "$initpath" >/dev/null
    enable_script
}

if [[ ${settings[os.initsystem]} = "init" ]]; then
    template="${settings[root]}/totem/init/upstart.services.template"
    initpath="/etc/init/holmes-totem-services.conf"
    function enable_script () {
        sudo initctl reload-configuration
        sudo service holmes-totem-services start
    }
elif [[ ${settings[os.initsystem]} = "systemd" ]]; then
    template="${settings[root]}/totem/init/systemd.services.template"
    initpath="/etc/systemd/system/holmes-totem-services.service"
    function enable_script () {
        sudo systemctl enable holmes-totem-services.service
        sudo systemctl start holmes-totem-services.service
    }
else
    error "> Unknown init system ${settings[os.initsystem]}, failed to install totem services as a system service."
    function install_initscript () {
        :
    }
fi

install_initscript
