# Toolbox

Helper scripts for managing Holmes


### How to easily move files from CRITs to Holmes-Storage

1. Adjust the filters in dump.js to match the samples you want to move
2. `mongo crits --quiet dump.js > out.txt`
3. Compile [crits-sample-server](https://github.com/cynexit/crits-sample-server) somewhere
4. `crits-sample-server --mongoServer=10.0.4.51 --dbName=crits --httpBinding=:8019`
5. Make sure your Holmes-Storage is running
6. `go run push_to_holmes.go --file=out.txt`


### How to easily move files from a local folder to Holmes-Storage

1. Move all you samples to one folder
2. `cd` into folder
3. `find `pwd` -type f > out.txt`
4. Make sure your Holmes-Storage is running
5. `go run push_to_holmes.go --file=out.txt`

