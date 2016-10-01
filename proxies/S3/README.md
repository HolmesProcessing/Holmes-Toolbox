# Docker S3 Proxy

## Overview
Docker S3 Proxy provides an optimized way to connect to compatible S3 backend when using [Holmes-Storage](https://github.com/HolmesProcessing/Holmes-Storage). This has been written for [Riak-CS](http://docs.basho.com/riak/cs/2.1.1/) but should work for other S3 backends with minimal to no changes.

## Configuration

### Configure host
Ensure the ulimit hard and soft on your system supports higher then 256000.

### Configure Docker
Copy the `haproxy.cfg.example` to haproxy.cfg`

Edit the file `haproxy.cfg` so it includes the IP addresses or Domain Names of all the Riak-CS worker nodes. For example, if your first Riak-CS node `riak-cs-1` is located at 192.168.0.50 change the following line:

FROM:

```
server riak-cs-1 <IP or Domain Name>:8080 weight 1 maxconn 1024 check
```

TO:
```
server riak-cs-1 192.168.0.50:8080 weight 1 maxconn 1024 check
```

Proceed with the same steps for riak-cs-1 to riak-cs-n nodes.

## Running
Build the Docker container:
```
docker build -t S3-Proxy .
```

Run the Docker Container:
```
docker run -d --name S3-Proxy --ulimit nofile=256000 S3-Proxy sh -c "ulimit -n"
```

## Acknowledgments
Much of this work is derived from the [official Bacho Riak-CS documentation](http://docs.basho.com/riak/cs/2.1.1/cookbooks/configuration/load-balancing-proxy/)
