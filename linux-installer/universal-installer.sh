#!/bin/bash
set -e

. helper/log.sh
. helper/output.sh
. helper/commands.sh
. helper/subinstaller_utility.sh

#-#-#-#-#-#---------------------------------------------------------------------
# Setup log
#
logger_start "$(pwd)/universal-installer.log"

#-#-#-#-#-#---------------------------------------------------------------------
# Gather settings (default + command line + computed)
#
declare -A settings=()

settings['totem.path']="/data/holmes-totem"
settings['totem.repo']="https://github.com/HolmesProcessing/Holmes-Totem.git"
settings['totem.branch']=""
settings['totem.branch.switch']=
settings['totem.erase']=
settings['totem.initscript']=1
settings['totem.services']=1

settings['storage.path']="/data/holmes-storage"
settings['storage.repo']="github.com/HolmesProcessing/Holmes-Storage"
settings['storage.config-url']=
settings['storage.config-create']="cluster"
settings['storage.erase']=
settings['storage.initscript']=1
settings['storage.skip-setup']=

settings['install.rabbitmq']=
settings['install.cassandra']=
settings['install.storage']=
settings['install.totem']=
settings['install.java']=
settings['java.version.major.minimum']=1
settings['java.version.minor.minimum']=0

settings['resume']=
settings['resume.checkoint.file']="$(pwd)/universal-installer.checkpoint"
touch "${settings[resume.checkoint.file]}"  # make sure file exists to avoid needless error msg
settings['resume.checkpoint']="$(cat "${settings[resume.checkoint.file]}")"

function select_install_cassandra () {
    settings["install.cassandra"]=1
    settings["install.java"]=1
    if [ "${settings['java.version.minor.minimum']}" -lt 7 ]; then
        settings["java.version.minor.minimum"]=7
    fi
}
function select_install_totem () {
    settings["install.totem"]=1
    settings["install.java"]=1
    if [ "${settings['java.version.minor.minimum']}" -lt 8 ]; then
        settings["java.version.minor.minimum"]=8
    fi
}

parameters_totem=()
parameters_storage=()
#
function display_options {
    error ""
    error_bg "*****************************************************************************************************************************************************"
    error ""
    error "Holmes-Toolbox Universal Installer"
    if [[ $# -gt 0 ]] && [[ $1 != "--help" && $1 != "-h" ]]; then
        error ""
        if [[ $opt = "__MISSING_ARGUMENT__" ]]; then
            error "Missing an argument."
        else
            error "Invalid option '$1'."
        fi
    fi
    error ""
    error "Usage: ./universal-installer.sh [--rabbitmq] [--cassandra] [--storage [path:...] [repo:...] [[config-url:...]|[config-create:...]] [no-initscript] [erase] [skip-setup]] [--totem [path:...] [repo:...] [branch:...] [no-services] [no-initscript] [erase]] [--resume]"
    error ""
    error ""
    error "--rabbitmq       : Install RabbitMQ (Task scheduler for Holmes-Totem)"
    error "--cassandra      : Install Apache-Cassandra (Database)"
    error ""
    error "--storage        : Install Holmes-Storage (Storage backend for Holmes-Totem)"
    error "  path:PATH          : Path to install Holmes-Storage in (defaults to /data/holmes-storage)"
    error "  repo:REPOSITORY    : Repository that Holmes-Storage is pulled from, without protocoll and without .git extension (defaults to github.com/HolmesProcessing/Holmes-Storage)"
    error "  config-url:URL     : Specify the location of the configuration file, will be loaded with curl. (incompatible with config-create)"
    error "  config-create:TYPE : Specify the configuration to be semi-automatically created (incompatible with config-url)"
    error "                       Available modes are:"
    error "                        - local              [install cassandra, objects stored in the local filesystem]"
    error "                        - local-objstorage   [external cassandra, objects stored in the local filesystem]"
    error "                        - local-cassandra    [install cassandra, objects stored in external S3 storage]"
    error "                        - cluster            [external cassandra, objects stored in external S3 storage]"
    error "                        - local-mongodb      [install mongodb]"
    error "                        - cluster-mongodb    [external mongodb]"
    error "                       Default is cluster."
    error "  no-initscript      : Don't install init scripts for upstart/systemd"
    error "  erase              : Purge any previous installation in the specified location"
    error "  skip-setup         : Don't run setup routines for the configured storages"
    error ""
    error "--totem          : Install Holmes-Totem"
    error "  path:PATH          : Path to install Holmes-Totem in (defaults to /data/holmes-totem)"
    error "  repo:REPOSITORY    : Repository that Holmes-Totem is pulled from (defaults to https://github.com/HolmesProcessing/Holmes-Totem.git)"
    error "  branch:BRANCH      : Branch to checkout on the repository (if none, no checkout will occur)"
    error "  no-services        : Don't install Totems default services (implied by no-initscript)"
    error "  no-initscript      : Don't install init scripts for upstart/systemd"
    error "  erase              : Purge any previous installation in the specified location"
    error ""
    error "--resume         : !!! Use with care and only with the exact same other parameters as before. This will attempt to resume execution at the last checkpoint. !!!"
    error ""
    error "Note: Some install option require paths, those must be unique, some options may allow for further customization by sub-options."
    error ""
    error ""
    error "Examples:"
    error ""
    error "  A complete setup (RabbitMQ + Cassandra + Storage + Totem) on one server:"
    error "    ./universal-installer.sh --rabbitmq --cassandra --storage config-create:local --totem repo:\"your-git-repo\""
    error ""
    error "  Only Storage, in cluster mode (configure Cassandra + S3 external):"
    error "    ./universal-installer.sh --storage"
    error ""
    error "  Only Totem, without its services as init job, erase an old installation if there is one in the same installation directory:"
    error "    ./universal-installer.sh --totem repo:\"your-git-repo\" no-services erase"
    error ""
    error "  Install Storage and Totem in custom directories:"
    error "    ./universal-installer.sh --totem repo:\"your-git-repo\" path:/custom/install/path/totem --storage path:/custom/install/path/storage"
    error ""
    error_bg "*****************************************************************************************************************************************************"
    error ""
    exit 1
}
if [ $# -eq 0 ]; then
    display_options
fi
while [ $# -gt 0 ]
do
    opt="$1"
    shift
    case "$opt" in
        
        "--rabbitmq")
            last_processed=
            settings["install.rabbitmq"]=1
            ;;
        
        "--cassandra")
            last_processed=
            select_install_cassandra
            ;;
        
        "--storage")
            last_processed="storage"
            settings["install.storage"]=1
            ;;
        
        "--totem")
            last_processed="totem"
            select_install_totem
            ;;
        
        "--resume")
            settings["resume"]=1
            ;;
        
        *)
            processed=
            if [[ "$opt" != --* ]] && [[ "$last_processed" = "totem" || "$last_processed" = "storage" ]]; then
                processed=1
                while [[ "$opt" != --* ]]; do
                    opt_key=$(echo "$opt" | sed -e 's/:.*//')
                    opt_val=$(echo "$opt" | sed -e 's/^[^:]*://')
                    if [[ "$last_processed" = "totem" ]]; then
                        case "$opt_key" in
                            "path")
                                settings['totem.path']="$opt_val"
                                ;;
                            "repo")
                                settings['totem.repo']="$opt_val"
                                ;;
                            "branch")
                                settings['totem.branch']="$opt_val"
                                settings['totem.branch.switch']=1
                                ;;
                            "no-services")
                                settings['totem.services']=
                                ;;
                            "no-initscripts")
                                settings['totem.initscript']=
                                ;;
                            "erase")
                                settings['totem.erase']=1
                                ;;
                            *)
                                processed=
                                opt="$opt' within '$last_processed"
                                break
                                ;;
                        esac
                    else
                        if [[ "$last_processed" = "storage" ]]; then
                            case "$opt_key" in
                                "path")
                                    settings['storage.path']="$opt_val"
                                    ;;
                                "repo")
                                    opt_val=$(echo "$opt_val" | sed -e "s/^https:\/\/\(.*\)\.git$/\1/")
                                    settings['storage.repo']="$opt_val"
                                    ;;
                                "config-url")
                                    settings['storage.config-url']="$opt_val"
                                    ;;
                                "config-create" | "create-config")
                                    settings['storage.config-create']="$opt_val"
                                    case "$opt_val" in
                                        "local" | 'local-cassandra')
                                            select_install_cassandra
                                            ;;
                                        "local-mongodb")
                                            settings['install.mongodb']=1
                                            ;;
                                    esac
                                    ;;
                                "no-initscript")
                                    settings['storage.initscript']=
                                    ;;
                                "erase")
                                    settings['storage.erase']=1
                                    ;;
                                "skip-setup")
                                    settings['storage.skip-setup']=1
                                    ;;
                                *)
                                    processed=
                                    opt="$opt' within '$last_processed"
                                    break
                                    ;;
                            esac
                        fi
                    fi
                    if [ $# -le 0 ]; then
                        break
                    fi
                    if [[ "$1" != --* ]]; then
                        opt="$1"
                        shift
                    else
                        break
                    fi
                done
            fi
            if [ ! $processed ]; then
                display_options "$opt"
            fi
            ;;
    esac
done

settings['root']=$(pwd) # remember starting directory

# find out the operating system flavor (Ubuntu/Debian/etc)
ensure_lsb_release
settings['os']=$(lsb_release -si)
settings['os.version']=$(lsb_release -sr)
settings['os.codename']=$(lsb_release -sc)
settings['os.initsystem']=
get_initsystem # sets os.initsystem
settings['os.kernel']=$(uname -r | sed -e 's/-.*//;s/\(.*\)\..*/\1/')
settings['os.kernel.major']=$(echo ${settings[os.kernel]} | sed -e 's/\..*//')
settings['os.kernel.minor']=$(echo ${settings[os.kernel]} | sed -e 's/.*\.//')

#-#-#-#-#-#---------------------------------------------------------------------
if [[ ${settings[os]} == "Ubuntu" || ${settings[os]} == "Debian" ]]; then
    if [[ ${settings[os]} = "Ubuntu" ]]; then
        _ifs=IFS
        IFS='.' read -r -a ubuntu_version <<< "${settings[os.version]}"
        IFS=_ifs
        settings["os.version.major"]=ubuntu_version[0]
        settings["os.version.minor"]=ubuntu_version[1]
    fi
    info "> Found OS flavor ${settings[os]}"
    info ""
    
    # -------------------------------------------------------------------- #
    if [ ${settings[resume]} ]; then
        info "> Attempting to resume at ${settings[resume.checkpoint]}"
    fi
    
    # -------------------------------------------------------------------- #
    if [ ${settings[install.rabbitmq]} ] && resume_at rabbitmq; then
        checkpoint rabbitmq
        info "> Installing RabbitMQ."
        ensure_apt_https
        ensure_wget
        echo 'deb http://www.rabbitmq.com/debian/ testing main' | sudo tee /etc/apt/sources.list.d/rabbitmq.list
        wget -O- https://www.rabbitmq.com/rabbitmq-signing-key-public.asc | sudo apt-key add -
        sudo apt-get update
        sudo apt-get install -y rabbitmq-server
        sudo service rabbitmq-server start
        info ""
    fi
    
    # -------------------------------------------------------------------- #
    if [ ${settings[install.java]} ] && resume_at java; then
        checkpoint java
        install_java8=
        if command -v java >/dev/null 2>&1; then
            version=$(java -version 2>&1 | tr '\n' '\t' | sed -e 's/\t.*//' -e 's/.*"\(.*\)".*/\1/' -e 's/_/./')
            _ifs=IFS
            IFS='.' read -r -a java_version <<< "$version"
            IFS=_ifs
            if [[ "${java_version[0]}" -eq "${settings[java.version.major.minimum]}" && "${java_version[1]}" -ge "${settings[java.version.minor.minimum]}" ]] ||
                [[ "${java_version[0]}" -gt "${settings[java.version.major.minimum]}" ]]; then
                :
                # java is already installed
            else
                install_java8=1
            fi
        fi
        if $install_java8; then
            # http://www.webupd8.org/2012/09/install-oracle-java-8-in-ubuntu-via-ppa.html
            info "> Installing Java8."
            ensure_apt_https
            ensure_add_apt_repository
            sudo add-apt-repository -y ppa:webupd8team/java
            sudo apt-get update
            echo oracle-java8-installer shared/accepted-oracle-license-v1-1 select true | sudo debconf-set-selections
            sudo apt-get install -y oracle-java8-installer
            sudo update-alternatives --config javac  # fix for whenever it is not automatically updated
        fi
    fi
    
    # -------------------------------------------------------------------- #
    if [ ${settings[install.cassandra]} ] && resume_at cassandra; then
        checkpoint cassandra
        info "> Installing Apache-Cassandra."
        ensure_apt_https
        # https://wiki.apache.org/cassandra/DebianPackaging
        # http://cassandra.apache.org/download/
        
        # install pre-requisites for cassandra
        sudo apt-get install -y python python3 libjna-java curl pgp
        
        if [[ ${settings[os]} = "Ubuntu" && ${settings[os.version.major]} -eq 16 ]]; then
            if ! sudo dpkg -s python-support 2>&1 >/dev/null; then
                # potential fix for ubuntu 16.04 missing python-support package in it's default repositories
                sudo curl -o /tmp/python-support_1.0.15_all.deb http://launchpadlibrarian.net/109052632/python-support_1.0.15_all.deb
                sudo dpkg -i /tmp/python-support_1.0.15_all.deb
                sudo rm /tmp/python-support_1.0.15_all.deb
            fi
        fi
        
        # the preferred installation method is broken, we have to install
        # cassandra from the datastax repository
        # version is fixed to 3.0.5 as that one is known to work with totem
        #
        # http://docs.datastax.com/en/cassandra/2.0/cassandra/install/installDeb_t.html
        echo "deb http://debian.datastax.com/community stable main" | sudo tee /etc/apt/sources.list.d/cassandra.sources.list
        curl -L http://debian.datastax.com/debian/repo_key | sudo apt-key add -
        
        sudo apt-get update
        sudo apt-get install -y cassandra=3.0.5
    fi
    
    # -------------------------------------------------------------------- #
    if [ ${settings[install.storage]} ] && resume_at storage; then
        checkpoint storage
        info "> Installing Holmes-Storage."
        . storage/install.sh
    fi
    
    # -------------------------------------------------------------------- #
    if [ ${settings[install.totem]} ] && resume_at totem; then
        checkpoint totem
        info "> Installing Holmes-Totem."
        . totem/install.sh
    fi
    
    # end ubuntu/debian

else
    
    error "> Unsupported Linux distribution."
    
fi

info "" || 1
