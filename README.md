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

## Metrics

```
% curl http://localhost:8080/metrics

# HELP pyth_exporter_rpc_requests_total Number of outgoing RPC requests from pyth_exporter to RPC nodes
# TYPE pyth_exporter_rpc_requests_total counter
pyth_exporter_rpc_requests_total 24
# HELP pyth_exporter_ws_active_conns Number of active WebSockets between pyth_exporter and RPC nodes
# TYPE pyth_exporter_ws_active_conns gauge
pyth_exporter_ws_active_conns 1
# HELP pyth_exporter_ws_events_total Number of WebSocket events delivered from RPC nodes to pyth_exporter
# TYPE pyth_exporter_ws_events_total counter
pyth_exporter_ws_events_total 14408
# HELP pyth_oracle_aggregated_conf_amount Last aggregated conf of Pyth product
# TYPE pyth_oracle_aggregated_conf_amount gauge
pyth_oracle_aggregated_conf_amount{pyth_product="EWxGfxoPQSNA2744AYdAKmsQZ8F9o9M7oKkvL3VM1dko"} 2e-05
# HELP pyth_oracle_aggregated_price Last aggregated price of Pyth product
# TYPE pyth_oracle_aggregated_price gauge
pyth_oracle_aggregated_price{pyth_product="EWxGfxoPQSNA2744AYdAKmsQZ8F9o9M7oKkvL3VM1dko"} 1.1326100000000001
# HELP pyth_oracle_publisher_conf_amount Last published product confidence by Pyth publisher
# TYPE pyth_oracle_publisher_conf_amount gauge
pyth_oracle_publisher_conf_amount{pyth_product="EWxGfxoPQSNA2744AYdAKmsQZ8F9o9M7oKkvL3VM1dko",pyth_publisher="AKPWGLY5KpxbTx7DaVp4Pve8JweMjKbb1A19MyL2nrYT"} 0.00014000000000000001
# HELP pyth_oracle_publisher_price Last published product price by Pyth publisher
# TYPE pyth_oracle_publisher_price gauge
pyth_oracle_publisher_price{pyth_product="EWxGfxoPQSNA2744AYdAKmsQZ8F9o9M7oKkvL3VM1dko",pyth_publisher="AKPWGLY5KpxbTx7DaVp4Pve8JweMjKbb1A19MyL2nrYT"} 1.1326500000000002
# HELP pyth_oracle_publisher_slot Last observed slot for Pyth publisher
# TYPE pyth_oracle_publisher_slot gauge
pyth_oracle_publisher_slot{pyth_product="EWxGfxoPQSNA2744AYdAKmsQZ8F9o9M7oKkvL3VM1dko",pyth_publisher="AKPWGLY5KpxbTx7DaVp4Pve8JweMjKbb1A19MyL2nrYT"} 1.16427278e+08
# HELP pyth_solana_publish_account_balance SOL balance of Pyth publish account in lamports
# TYPE pyth_solana_publish_account_balance gauge
pyth_solana_publish_account_balance{pyth_publisher="AKPWGLY5KpxbTx7DaVp4Pve8JweMjKbb1A19MyL2nrYT"} 4.950539e+10
# HELP pyth_txs_total Approximate number of Pyth transactions sent
# TYPE pyth_txs_total counter
pyth_txs_total{pyth_publisher="AKPWGLY5KpxbTx7DaVp4Pve8JweMjKbb1A19MyL2nrYT",tx_status="failed"} 10
pyth_txs_total{pyth_publisher="AKPWGLY5KpxbTx7DaVp4Pve8JweMjKbb1A19MyL2nrYT",tx_status="success"} 67
```
