FROM golang:alpine

# create folder
RUN mkdir -p /service
WORKDIR /service

# get go dependencies
RUN apk add --no-cache \
		git \
	&& go get github.com/julienschmidt/httprouter \
	&& rm -rf /var/cache/apk/*

###
# {$name} specific options
###

# add the files to the container
COPY LICENSE /service
COPY README.md /service
COPY {$name}.go /service
# build {$name}
RUN go build {$name}.go

# add the configuration file (possibly from a storage uri)
ARG conf=service.conf
ADD $conf /service/service.conf

CMD ["./{$name}", "--config=service.conf"]
