#!/bin/bash

function install_prerequisites () {
    if ! resume_at "totem.prerequisites"; then
        return
    fi
    checkpoint "totem.prerequisites"
    
    info "> Installing prerequisites for Holmes-Totem."
    sudo apt-get update
    sudo apt-get install -y build-essential \
        python-dev \
        python-pip \
        apt-transport-https \
        software-properties-common \
        git \
        curl
    info ""
}

function install_sbt {
    if ! resume_at "totem.sbt"; then
        return
    fi
    checkpoint "totem.sbt"
    
    info "> Installing SBT."
    echo "deb https://dl.bintray.com/sbt/debian /" | sudo tee /etc/apt/sources.list.d/sbt.list > /dev/null
    sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 2EE0EA64E40A89B84B2DF73499E82A75642AC823
    sudo apt-get update
    sudo apt-get install -y sbt
    info ""
}

function install_services () {
    if [[ ${settings[totem.initscript]} && ${settings[totem.services]} ]]; then
        info "> Installing Holmes-Totem services."
        local kmajor="${settings[os.kernel.major]}"
        local kminor="${settings[os.kernel.minor]}"
        if [[ $kmajor -lt 3 ]] || [[ $kmajor -eq 3 && $kminor -lt 10 ]]; then
            error "> Cannot install services. Your kernel version does not support Docker."
            error "  If you wish to install Holmes-Totems services on this machine, please upgrade your kernel (>=3.10)."
            return
        else
            if command -v docker >/dev/null 2>&1; then
                info "> Detected an existing Docker installation."
            else
                info "> No Docker installation found."
                . totem/docker/install_docker.sh
            fi
            if command -v docker-compose >/dev/null 2>&1; then
                info "> Detected an existing docker-compose installation."
            else
                info "> No docker-dompose installation found."
                . totem/docker/install_docker_compose.sh
            fi
            info ""
        fi
        info "> Assigning user 'totem' to the 'docker' group."
        sudo usermod -aG docker totem
        . totem/init/install_services.sh
        info ""
    fi
}

function install_totem {
    if resume_at "totem.prepare"; then
        checkpoint "totem.prepare"
    
        info "> Preparing Holmes-Totem."
        
        empty_install_path "totem"
        create_user "totem"
        
        sudo mkdir -p "${settings[totem.path]}"
        sudo chown "$USER":"$USER" "${settings[totem.path]}"
        
        git clone "${settings[totem.repo]}" "${settings[totem.path]}"
        if [ ${settings[totem.branch.switch]} ]; then
            cd "${settings[totem.path]}"
            git checkout "${settings[totem.branch]}"
        fi
    
        # check if config exists
        cd "${settings[totem.path]}/config"
        if [[ ! -f "totem.conf" ]]; then
            if [[ -f "totem.conf.example" ]]; then
                cp totem.conf.example totem.conf
            else
                error "> No Holmes-Totem configuration file supplied, please put one into '${settings[totem.path]}/config'."
                error "  See https://www.github.com/HolmesProcessing/Holmes-Totem for an example config file."
            fi
        fi
        if [ ${settings[totem.services]} ]; then
            if [[ ! -f "docker-compose.yml" ]]; then
                if [[ -f "docker-compose.yml.example" ]]; then
                    cp docker-compose.yml.example docker-compose.yml
                else
                    error "> No docker-compose.yml configuration file supplied, but is required to run Holmes-Totem services as a system service."
                    error "  See https://www.github.com/HolmesProcessing/Holmes-Totem for an example config file."
                fi
            fi
        fi
    fi
    
    # build
    if resume_at "totem.build"; then
        checkpoint "totem.build"
        
        info "> Building Holmes-Totem."
        sudo chown -R totem:totem "${settings[totem.path]}"
        sudo su totem -c "cd '${settings[totem.path]}' && sbt assembly"
    fi
    
    # if totem init script should be written do so now
    if resume_at "totem.initscripts"; then
        checkpoint "totem.initscripts"
        
        if [ ${settings[totem.initscript]} ]; then
            info "> Setting up init script(s)"
            cd "${settings[root]}"
            install_services
            . totem/init/install_totem.sh
            # Finish notice
            printf "${GREEN}"
            echo "> Finished installing Holmes-Totem!"
            echo "  To start/stop Totem or its services (if installed), please use your init systems functionality (initctl or systemctl)."
            if [ ${settings[totem.services]} ]; then
                echo "  Please note that docker-compose will take some time to build your services and as such should be run manually at least once (init will time out)."
                echo "  Please also note, that all services need to build successfully for the holmes-totem-services service to start up correctly."
                echo "  If there are any errors, try executing 'docker-compose up' in '${settings[totem.path]}/config' to see what errors are thrown."
            fi
            printf "${ENDC}"
        else
            # Finish notice
            printf "${GREEN}"
            echo "> Finished installing Holmes-Totem."
            echo "  To launch Holmes-Totem change into totem users context (sudo su totem) and issue the following commands:"
            echo "  cd ${settings[totem.path]}/config"
            echo "  docker-compose up -d"
            echo "  cd .."
            echo "  java -jar ./target/scala-2.11/totem-assembly-1.0.jar ./config/totem.conf"
            printf "${ENDC}"
        fi
    fi
}

function main () {
    install_prerequisites
    install_sbt
    install_totem
}

main
