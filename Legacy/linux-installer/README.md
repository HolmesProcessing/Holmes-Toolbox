# Due to recent changes and upgrades of the HolmesProcessing system, these installation scripts are no longer up-to-date and will raise errors. We strongly encourage our users to install the Holmes components manually instead of using these resources.

### How to install Holmes-Totem/Storage on Ubuntu/Debian using the install script

1. `git clone https://github.com/HolmesProcessing/Holmes-Toolbox.git`
2. `cd Holmes-Toolbox/linux-installer`
3. Execute the installer with `-h` and read the help: `./universal-installer.sh -h`
4. Examples:
    - Basic Totem Installion (post-install: totem + service configuration required):
        - `./universal-installer.sh --totem`
    - Custom Totem Installation (custom configuration):
        - `./universal-installer.sh --totem repo:"https://github.com/your-username/Holmes-Totem.git"`
    - Install Storage:
        - `./universal-installer.sh --storage`
    - Single Node setup:
        - `./universal-installer.sh --rabbitmq --cassandra --totem repo:YOUR_REPO --storage create-config:local`
