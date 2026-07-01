# exp-app

Performance workload tools for BETS experiments.

The module provides three binaries:

- `exp-app`: identity-scoped setup + workload execution + report export
- `exp-app-setup`: global one-time setup (`Init`, SICAR trusted provider, policies, TEE setup)
- `generate-profile`: network profile generator from deployment artifacts

## Build

```bash
cd experiments/exp-app
make build-all
```

Build outputs:

- `bin/exp-app`
- `bin/exp-app-setup`
- `bin/generate-profile`

## New Deployment/Execution Cycle

The recommended flow is orchestrated by deployment scripts in `experiments/deploy/scripts`:

1. Deploy infra and exp-app pods (`make experiments` from repo root).
2. Run experiments (`make experiments-run` from repo root), which does:
   - run `exp-app-setup` once in a selected pod;
   - run `exp-app` concurrently in all exp-app pods;
   - collect and download artifacts per pod.

This split keeps global setup out of `exp-app`, avoiding cross-pod race conditions and making runs repeatable.

## `exp-app-setup` (global setup)

Runs only global operations.

Example:

```bash
./bin/exp-app-setup \
  --profile=network-profile.json \
  --organization=mma \
  --user-index=0 \
  --arm-template=../../tee_auction/azure/arm_template.json
```

Flags:

- `--profile` (required)
- `--organization` (optional, default first organization)
- `--user-index` (default `0`)
- `--arm-template` (default from `EXP_APP_ARM_TEMPLATE` env, fallback to repo-relative path)

## `exp-app` (identity setup + workload)

`exp-app` now performs only identity-scoped setup and workload execution.
It does not run global setup.

Example:

```bash
./bin/exp-app \
  --profile=network-profile.json \
  --organization=mma \
  --user-count=5 \
  --duration=5m \
  --concurrency=20 \
  --enable-metrics \
  --metrics-output=monitoring-exports \
  --metrics-formats=png,html,pdf,json \
  --output-json=results.json \
  --output-csv=results.csv
```

Key workload flags:

- `--profile` (required)
- `--organization` (optional)
- `--user-count` (0 means all users from profile)
- `--duration`, `--concurrency`
- `--tps`, `--burst`
- `--mint-interval`, `--buy-bid-interval`, `--sell-bid-interval`, `--auction-interval`
- `--run-coupled` (defaults from `EXP_APP_RUN_COUPLED` env)

Metrics/report flags:

- `--enable-metrics`
- `--metrics-output`
- `--metrics-formats` (`png,html,pdf,json`)
- `--output-json`, `--output-csv`

## Metrics Collection Details

When `--enable-metrics` is enabled, exp-app collects:

- baseline Prometheus snapshot (before workload)
- final Prometheus snapshot (after workload)
- delta snapshot (`final - baseline`)

Sources include:

- chaincode metrics endpoints
- peer metrics endpoints
- orderer metrics endpoints
- exp-app runtime metrics collector

### Chaincode function metrics

Chaincode metrics are captured per function using transaction labels.
Current parser supports:

- preferred label: `tx_name`
- fallbacks: `function`, `txName`

Supported metric families include:

- `<chaincode>_chaincode_tx_requests_total`
- `<chaincode>_chaincode_tx_requests`
- `<chaincode>_chaincode_tx_duration_seconds`
- `<chaincode>_chaincode_tx_duration_seconds_sum`

If request counters are not present, request counts are derived from histogram sample count.

## Output Artifacts

Typical per-run files:

- `results.json` / `results.csv` (aggregate in-pod runtime view)
- `results.json.user-XX` / `results.csv.user-XX` (per-runtime exports)
- `monitoring-exports/metrics-baseline.json`
- `monitoring-exports/metrics-final.json`
- `monitoring-exports/metrics-delta.json`
- `monitoring-exports/report.html`
- `monitoring-exports/report.pdf` (when requested)
- `monitoring-exports/charts/*.png`

HTML report chart references are relative (`charts/*.png`) so copied reports remain portable.

## Docker Image

`experiments/exp-app/Dockerfile` builds and ships:

- `/app/exp-app`
- `/app/exp-app-setup`
- `/app/generate-profile`

## Local Dev Commands

```bash
make build
make build-setup
make build-all
make test
make fmt
```
