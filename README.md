# blobstreamx-monitor

Simple monitoring tool for BlobstreamX contract. It allows provers/relayers for BlobstreamX to monitor the BlobstreamX updates, and be notified.

The tool currently sends an OTEL message whenever a new batch is submitted, verified and committed by the contract. This message is a histogram `blobstreamx_monitor_submitted_heights` of type `Int64Histogram`.

## Install

1. [Install Go](https://go.dev/doc/install) 1.21
2. Clone this repo
3. Install the BlobstreamX-monitor CLI

 ```shell
make install
```

## Usage

```sh
# Print help
blobstreamx-monitor --help
```

## How to run

To run the monitoring tool, make sure you have access to an [otel collector](https://opentelemetry.io/docs/collector/installation/), by default it targets the `"localhost:4318"` endpoint:

```shell
blobstreamx-monitor start \
  --evm.rpc <evm_chain_rpc> \
  --evm.contract-address <blobstreamx_contract_address> \
  --metrics.endpoint <otel_collector_endpoint> \
  --log.level debug
```

To start a local monitoring environment, refer to the example setup in the `example` folder that spins up an otel collector, prometheus and grafana which can be used to check the above metric from BlobstreamX contract.

After running the tool, operators can setup alerts to alarm them if there is a BlobstreamX liveness issue and begin investigating running backup provers.
