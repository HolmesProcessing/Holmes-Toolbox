# Template Generator for Holmes Totem Services

## OverView

This program generates the boiler plate code required to create a Holmes Totem Service. Currently this is only for creating Static Services.

### Purpose
To create a service, we generally need to do a lot of copy-paste business. To avoid this extra work, this program generates the all the essential code needed. This template files can also be used as a reference to for how a typical holmes-totem-service looks like. There could be many templates (ex: RESTFull, gRPC) in their preferred language.the user just has to provide which kind of template he need. This helps people who wants to create a new services in Holmes easier easier to choose their template and focus on Service logic.

## Configuration 

Specify the required options for creating service in the configuration file`parse.conf`.

```json
{
	"type" : "RESTFull",
	"language" : "go",
	"servicename" : "helloworld",
	"version" : "1.0"
}
```

## Installation
You need to have Go installed. Configure the required settings in `parse.conf` and just run `parse.go` file.

```
$ go run parse.go --config=parse.conf
```

## Implimentation

The parser takes the template (in the template folder) and configuration file as input and create a directory with the servicename and creates all the boilerplate code required. After this you can directly jump into servicelogic section of the {servicename}.go and add additional configuration settings and finally will add dependencies to the Dockerfile.

The created folder structure will be:
```
ServiceName
		|-----Service.{go,py}
		|-----Dockerfile
		|-----ServiceREST.Scala
		|-----Service.conf
		|-----README.md
```		

## TODO

1. Create a Template for gRPC.
