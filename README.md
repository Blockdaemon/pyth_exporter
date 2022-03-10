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
        Run in development mode
  -env string
        Pyth env (devnet, testnet, mainnet) (default "mainnet")
  -listen string
        Address where to serve debug info and metrics HTTP server (default ":8080")
  -log-level value
        Log level
  -products value
        Pyth product keys (default all)
  -program value
        Pyth program key (derived from env)
  -publishers value
        Pyth publishers (default all)
  -rpc string
        Solana RPC URL
  -ws string
        Solana WebSocket RPC URL
```

## Deployment

[./docker-compose.yml](./docker-compose.yml) defines a reference [Docker Compose](https://docs.docker.com/compose/) deployment on a single host.

The compose config includes the following services:
- pyth_exporter (this repo)
- Prometheus monitoring agent
- Grafana monitoring UI

### Requirements

The reference deployment requires Docker Compose: [Compose installation guide](https://docs.docker.com/compose/install/)

### Configuration

The env vars file contains deployment-specific config.
Copy the example config and adjust it.

```shell
cp docker.example.env docker.env
$EDITOR docker.env
```

- `*_IMAGE`: Docker image strings
- `SOLANA_RPC`: Solana RPC access
- `SOLANA_WS`: Solana WebSocket access
- `SOLANA_ENV`: Environment name (devnet, testnet, mainnet)

### Operations

Start all services

```shell
docker-compose --env-file docker.env up -d
```

Stop all services

```shell
docker-compose down
```

Stop all services and delete all data (!)

```shell
docker-compose down -v --remove-orphans
```

View service status

```shell
docker-compose ps
```

View service logs

```shell
docker-compose logs
```
