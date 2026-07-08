# Experiments

This directory contains the deployment and workload tooling for BETS performance experiments on Minikube and AKS.

## Current Execution Cycle

1. Deploy Fabric network + chaincodes + in-cluster exp-app pods:

```bash
make experiments
```

2. Run experiment cycle (global setup once, then all exp-app pods concurrently, then download artifacts):

```bash
make experiments-run
```

3. Results are downloaded to:

```text
experiments/deploy/vars/exp-app-runs/<run-id>/<pod>/
```

Run-level artifacts are also written under:

```text
experiments/deploy/vars/exp-app-runs/<run-id>/
```

- `exp_app_flags.json` (captured run configuration)
- `cluster-metrics-baseline.json`
- `cluster-metrics-final.json`
- `cluster-metrics-delta.json`
- `aggregate/results.csv` and `aggregate/results.json`
- `aggregate/monitoring-exports/` (copied from one pod)

## Topology And Flow

```mermaid
flowchart LR
    A["Developer machine: make experiments"] --> K["Kubernetes cluster"]
    C["Developer machine: make experiments-run"] --> K
    K --> D["Downloaded artifacts: experiments/deploy/vars/exp-app-runs/run-id/"]

    subgraph K["Kubernetes cluster (namespace fabric-experiments)"]
        subgraph FAB["Fabric network"]
            O["Orderers: orderer0, orderer1"]
            P["Peers: mma, farmers, companies"]
            CC["Chaincodes: carbon v1.0, interop v1.0"]
            S["SICAR mock service: sicar-mock:8443"]

            O <--> P
            P <--> CC
            P <--> S
        end

        subgraph EXP["Experiment apps"]
            G["exp-app-setup (run once)"]
            X["exp-app pods (N = peer orgs with peers, default 3)"]
            M["Prometheus endpoints: chaincodes, peers, orderers"]

            G --> X
            X --> M
        end

        X --> CC
        X --> P
        X --> S
    end
```

Each pod directory contains:

- `results.json` and `results.csv` (aggregate runtime metrics in that pod)
- `results.json.user-XX` and `results.csv.user-XX` (per-runtime metrics)
- `monitoring-exports/` with:
  - `metrics-baseline.json`
  - `metrics-final.json`
  - `metrics-delta.json`
  - `metrics-report.json` (if requested)
  - `report.html`
  - `report.pdf` (if requested)
  - `charts/*.png`

`cluster-metrics-final.json` includes role-based resource snapshots:

- `metrics.component_totals` for peers/orderers/exp-app/chaincodes (CPU and memory)
- `metrics.component_per_pod` for per-pod CPU and memory maps
- `timeseries` with the workload window (`start_ts`, `end_ts`, `step`) and range-query series for totals and per-pod values

`cluster-metrics-delta.json` includes deep numeric deltas for all fields under `metrics`, including role totals and per-pod maps, plus the final `timeseries` payload for charting.

## Important Design Notes

- Global setup is separated from workload execution:
  - `exp-app-setup` performs one-time global operations (`Init`, SICAR trusted provider, active policies, TEE setup).
  - `exp-app` performs identity-scoped setup and workload only.
- Multi-pod execution is concurrent after global setup.
- Metrics collection includes Fabric and chaincode Prometheus endpoints plus app-side runtime metrics.
- Chaincode function metrics are collected by function name labels (`tx_name`, with fallback support for older labels).

## Useful Variables for `experiments-run`

`make experiments-run` uses `experiments/deploy/scripts/run_exp_app_experiments.bash`. Common overrides:

```bash
NAMESPACE=fabric-experiments \
DURATION=10m \
CONCURRENCY=30 \
METRICS_FORMATS=png,html,pdf,json \
RUN_GLOBAL_SETUP=true \
SETUP_USER_INDEX=0 \
make experiments-run
```

Other available knobs include `TPS`, `BURST`, `USER_COUNT`, `MINT_INTERVAL`, `BUY_BID_INTERVAL`, `SELL_BID_INTERVAL`, `AUCTION_INTERVAL`.

To tune cluster-metrics range granularity, set `CLUSTER_METRICS_STEP` (default: `15s`).

To have Prometheus/Grafana scrape Fabric peer/orderer/chaincode metrics, enable ServiceMonitors during deploy:

```bash
MONITORING_SERVICEMONITORS_ENABLED=true \
MONITORING_RELEASE_NAME=monitoring \
MONITORING_NAMESPACE=monitoring \
make experiments
```

## Cleanup

```bash
make experiments-clean
```
