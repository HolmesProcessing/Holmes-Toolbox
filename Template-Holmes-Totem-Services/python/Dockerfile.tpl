FROM python:alpine

# add tornado
RUN pip3 install tornado

# create folder
RUN mkdir -p /service
WORKDIR /service

###
# {$name} v2 specific options
##

# add dependencies for {$name}


# add the files to the container
COPY LICENSE /service
COPY README.md /service
COPY {name}.py /service

# add the configuration file (possibly from a storage uri)
ARG conf=service.conf
ADD $conf /service/service.conf

CMD ["python3", "{$name}.py"]
