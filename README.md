# Pyth Network Prometheus Exporter

## Summary

Prometheus Exporter for Pyth Network on-chain metrics about a publisher.

- Main repo: https://gitlab.com/Blockdaemon/solana/pyth_exporter
- GitHub mirror: https://github.com/Blockdaemon/pyth_exporter

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

## Configuration

```
Usage of pyth_exporter:
  -dev
        Run in development mode?
  -listen string
        Address where to serve debug info and metrics HTTP server (default ":8080")
  -log-level value
        Log level
  -products value
        Pyth product keys (space separated)
  -program value
        Pyth program key
  -publishers value
        Pyth publishers (space separated)
  -rpc string
        RPC URL
  -ws string
        WebSocket RPC URL
```
