# Holmes-Toolbox

# Due to recent upgrades and changes of HolmesProcessing, these scripts are no longer suitable for installing any of the Holmes components. We strongly suggest you perform manual installation instead.

## Overview
The Holmes-Toolbox provides useful scripts for managing Holmes projects.

### Highlights

| Folder | Project | Description |
| --- | --- | --- |
| linux-installer | All | Universal install scripts for convience |
| proxies | [Holmes-Storage](https://github.com/HolmesProcessing/Holmes-Storage) | Docker based proxies for optimizing database connections |
| start-scripts | [Holmes-Totem](https://github.com/HolmesProcessing/Holmes-Totem), [Holmes-Storage](https://github.com/HolmesProcessing/Holmes-Storage) | Automatic start up scripts for systemd and upstart |
| test-scripts | [Holmes-Totem](https://github.com/HolmesProcessing/Holmes-Totem) | Directly tasks [Holmes-Totem](https://github.com/HolmesProcessing/Holmes-Totem) with a series of objects |


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
2. e.g. `go run push_to_holmes.go --gateway https://127.0.0.1:8090 --user test --pw test --tags '["tag1","tag2"]' --comment "mycomment" --insecure --workers 5 --src virusshare --dir $dir`

Alternative way:

1. Move all you samples to one folder
2. `cd` into folder
3. `find `pwd` -type f > out.txt`
4. Make sure your Holmes-Storage and Gateway are running
5. e.g. `go run push_to_holmes.go --gateway https://127.0.0.1:8090 --user test --pw test --tags '["tag1","tag2"]' --comment "mycomment" --insecure --workers 5 --src virusshare --file out.txt`

### How to easily task Holmes-Totem:
1. Create a file containing a line with the SHA256-Sum, the filename, and the source (separated by single spaces) for each sample.
2. e.g. `go run push_to_holmes.go --gateway https://127.0.0.1:8090 --user test --pw test --tags '["tag1","tag2"]' --comment "mycomment" --insecure --tasking --file sampleFile --tasks '{"PEINFO":[], "YARA":[]}'`

### Resuming an incomplete upload
When executing Holmes-Toolbox for uploading samples, Holmes-Toolbox creates a new log-file in the "log"-folder. The name of the log-file is printed after Toolbox started and contains the current timestamp. If your upload crashes at some point, you can resume the upload by specifying the option `--resume`:
```sh
go run push_to_holmes.go --resume log/Holmes-Toolbox_2016-09-25_20:39:44.log --workers 5
```
All the commandline-parameters that were used for the upload which created the log-file, are automatically inserted, except for the "--workers" option. This makes it possible to start the upload with a different number of worker-threads, than before, if you experienced a bad performance before.
When resuming, all the samples that were accepted before, are skipped (i.e. those that returned with a code of 200). All samples that were rejected (different code than 200) and those that were not yet tried, are uploaded.

Resuming an upload will also create a new log-file, where all the previously successful (and therefore skipped) uploads are marked with 200. You can easily get a list of all the files that were not correctly uploaded by executing
```sh
tail log/Holmes-Toolbox_2016-10-03_22:56:38.log -n +2 | grep -v 200
```
