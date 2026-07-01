#!/usr/bin/env python3

import argparse
import csv
import json
import os
from collections import defaultdict


def percentile_from_sorted(values, q):
    if not values:
        return 0.0
    if q <= 0:
        return values[0]
    if q >= 1:
        return values[-1]
    idx = int((len(values) - 1) * q)
    if idx < 0:
        idx = 0
    if idx >= len(values):
        idx = len(values) - 1
    return values[idx]


def bool_from_csv(value):
    return str(value).strip().lower() in {"true", "1", "yes"}


def parse_float(value, default=0.0):
    try:
        return float(value)
    except (TypeError, ValueError):
        return default


def collect_pod_dirs(run_dir):
    pod_dirs = []
    for entry in sorted(os.listdir(run_dir)):
        if entry == "aggregate":
            continue
        full_path = os.path.join(run_dir, entry)
        if os.path.isdir(full_path):
            pod_dirs.append((entry, full_path))
    return pod_dirs


def read_pod_results(pod_dirs):
    rows = []
    pod_summaries = {}

    for pod_name, pod_dir in pod_dirs:
        csv_path = os.path.join(pod_dir, "results.csv")
        json_path = os.path.join(pod_dir, "results.json")

        if os.path.isfile(csv_path):
            with open(csv_path, newline="", encoding="utf-8") as handle:
                reader = csv.DictReader(handle)
                for row in reader:
                    rows.append(
                        {
                            "pod": pod_name,
                            "id": row.get("id", ""),
                            "scenario": row.get("scenario", ""),
                            "timestamp": row.get("timestamp", ""),
                            "latency_ms": parse_float(row.get("latency_ms", "0")),
                            "success": bool_from_csv(row.get("success", "false")),
                            "error": row.get("error", ""),
                        }
                    )

        if os.path.isfile(json_path):
            with open(json_path, encoding="utf-8") as handle:
                try:
                    pod_summaries[pod_name] = json.load(handle)
                except json.JSONDecodeError:
                    pod_summaries[pod_name] = {}

    return rows, pod_summaries


def write_aggregate_csv(rows, out_csv):
    rows.sort(key=lambda item: (item["timestamp"], item["pod"], item["id"]))

    with open(out_csv, "w", newline="", encoding="utf-8") as handle:
        writer = csv.writer(handle)
        writer.writerow(["pod", "id", "scenario", "timestamp", "latency_ms", "success", "error"])
        for row in rows:
            writer.writerow(
                [
                    row["pod"],
                    row["id"],
                    row["scenario"],
                    row["timestamp"],
                    f"{row['latency_ms']:.2f}",
                    "true" if row["success"] else "false",
                    row["error"],
                ]
            )


def build_aggregate_json(run_id, run_dir, out_csv, pod_dirs, rows, pod_summaries):
    total_txs = len(rows)
    successful_txs = sum(1 for row in rows if row["success"])
    failed_txs = total_txs - successful_txs

    latencies = sorted(row["latency_ms"] for row in rows)
    latency_avg = sum(latencies) / len(latencies)
    latency_min = latencies[0]
    latency_max = latencies[-1]
    latency_p50 = percentile_from_sorted(latencies, 0.50)
    latency_p95 = percentile_from_sorted(latencies, 0.95)
    latency_p99 = percentile_from_sorted(latencies, 0.99)

    throughput_tps = 0.0
    for pod_data in pod_summaries.values():
        summary = pod_data.get("summary", {}) if isinstance(pod_data, dict) else {}
        throughput_tps += parse_float(summary.get("throughput_tps", 0.0))

    scenario_rows = defaultdict(list)
    for row in rows:
        scenario_rows[row["scenario"]].append(row)

    by_scenario = {}
    for scenario, s_rows in sorted(scenario_rows.items()):
        s_total = len(s_rows)
        s_success = sum(1 for row in s_rows if row["success"])
        s_failed = s_total - s_success
        s_latencies = sorted(row["latency_ms"] for row in s_rows)
        by_scenario[scenario] = {
            "name": scenario,
            "total_txs": s_total,
            "successful_txs": s_success,
            "failed_txs": s_failed,
            "latency_p50_ms": percentile_from_sorted(s_latencies, 0.50),
            "latency_p95_ms": percentile_from_sorted(s_latencies, 0.95),
            "latency_p99_ms": percentile_from_sorted(s_latencies, 0.99),
            "latency_avg_ms": sum(s_latencies) / len(s_latencies),
            "latency_min_ms": s_latencies[0],
            "latency_max_ms": s_latencies[-1],
            "success_rate_percent": (100.0 * s_success / s_total) if s_total else 0.0,
        }

    return {
        "run_id": run_id,
        "pod_count": len(pod_dirs),
        "source_pods": [pod for pod, _ in pod_dirs],
        "summary": {
            "total_transactions": total_txs,
            "successful": successful_txs,
            "failed": failed_txs,
            "success_rate_percent": (100.0 * successful_txs / total_txs) if total_txs else 0.0,
            "throughput_tps": throughput_tps,
        },
        "latency_ms": {
            "average": latency_avg,
            "min": latency_min,
            "max": latency_max,
            "p50": latency_p50,
            "p95": latency_p95,
            "p99": latency_p99,
        },
        "by_scenario": by_scenario,
        "files": {
            "aggregate_csv": os.path.relpath(out_csv, run_dir),
        },
        "per_pod_summary": {
            pod: data.get("summary", {}) if isinstance(data, dict) else {}
            for pod, data in sorted(pod_summaries.items())
        },
    }


def main():
    parser = argparse.ArgumentParser(description="Aggregate exp-app pod results into one CSV/JSON report")
    parser.add_argument("--run-dir", required=True, help="Run directory containing one folder per pod")
    parser.add_argument("--output-csv", required=True, help="Output path for merged CSV")
    parser.add_argument("--output-json", required=True, help="Output path for merged JSON summary")
    parser.add_argument("--run-id", required=True, help="Run ID for metadata")
    args = parser.parse_args()

    pod_dirs = collect_pod_dirs(args.run_dir)
    rows, pod_summaries = read_pod_results(pod_dirs)

    if not rows:
        raise SystemExit(f"no pod results.csv files found under {args.run_dir}")

    os.makedirs(os.path.dirname(args.output_csv), exist_ok=True)
    os.makedirs(os.path.dirname(args.output_json), exist_ok=True)

    write_aggregate_csv(rows, args.output_csv)
    aggregate_json = build_aggregate_json(args.run_id, args.run_dir, args.output_csv, pod_dirs, rows, pod_summaries)

    with open(args.output_json, "w", encoding="utf-8") as handle:
        json.dump(aggregate_json, handle, indent=2)

    print(f"Aggregated CSV: {args.output_csv}")
    print(f"Aggregated JSON: {args.output_json}")


if __name__ == "__main__":
    main()
