# Toolbox

Helper scripts for managing Holmes


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


### How to easily move files from CRITs to Holmes-Storage

1. Adjust the filters in dump.js to match the samples you want to move
2. `mongo crits --quiet dump.js > out.txt`
3. Compile [crits-sample-server](https://github.com/cynexit/crits-sample-server) somewhere
4. `crits-sample-server --mongoServer=10.0.4.51 --dbName=crits --httpBinding=:8019`
5. Make sure your Holmes-Storage is running
6. `go run push_to_holmes.go --file=out.txt`


### How to easily move files from a local folder to Holmes-Storage

1. Make sure your Holmes-Storage and your Holmes-Mastergateway are running
2. e.g. `go run push_to_holmes.go --gateway https://127.0.0.1:8090 --user test --pw test --dir $dir --uid 1 --src foo --comment something --workers 5 --insecure`

Alternative way:

1. Move all you samples to one folder
2. `cd` into folder
3. `find `pwd` -type f > out.txt`
4. Make sure your Holmes-Storage and Gateway are running
5. e.g. `go run push_to_holmes.go --gateway https://127.0.0.1:8090 --user test --pw test --file out.txt --uid 1 --src foo --comment something --workers 5 --insecure`

### How to easily task Holmes-Totem:
1. Create a file containing a line with the SHA256-Sum, the filename, and the source (separated by single spaces) for each sample.
2. e.g. `go run push_to_holmes.go --gateway https://127.0.0.1:8090 --tasking --file sampleFile --user test --pw test --tasks '{"PEINFO":[""], "YARA":[""]}' --tags '["mytag"]' --comment 'mycomment' --insecure`

