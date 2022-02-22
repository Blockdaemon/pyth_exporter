# Pyth Network Prometheus Exporter

## Summary

Prometheus Exporter for Pyth Network on-chain metrics about a publisher.

## Building

### Building from source

The Go 1.17 toolchain or newer is required to build from source.

To pull dependencies and build the publisher program, run the following:

```shell
go build -o ./pyth_exporter .
```

### Docker image

This program is also available as a Docker image based on [Alpine Linux](https://alpinelinux.org).

To build the Docker image from source:

```shell
docker build -t db-pyth .
```

The GitLab CI integration builds Docker images for every branch and release:

```shell
# Latest master build
docker pull registry.gitlab.com/blockdaemon/solana/pyth_exporter/master:latest

# Latest tagged release
docker pull registry.gitlab.com/blockdaemon/solana/pyth_exporter:latest
```
