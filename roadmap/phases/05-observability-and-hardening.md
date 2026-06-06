# Phase 5: Observability And Hardening

## Goal

Prove reliability and performance with measurements, not assumptions.

## Duration

2-3 weeks

## Learning Focus

- Prometheus and Grafana dashboards
- Tracing fundamentals with OpenTelemetry
- Load testing and performance analysis
- Failure-mode thinking

## Build Tasks

- [ ] Add request duration histograms and endpoint counters.
- [ ] Instrument traces for redirect and shorten paths.
- [ ] Create Grafana dashboards (latency, errors, cache hit rate, consumer lag).
- [ ] Run load tests for baseline, read-heavy, and write-burst scenarios.
- [ ] Compare sharded vs non-sharded results.
- [ ] Document failure scenarios and fallback behavior.

## Deliverables

- Benchmark report with p50/p95/p99 and throughput.
- At least one dashboard screenshot per critical subsystem.
- Written explanation of bottlenecks and next optimizations.

## Exit Criteria

- You can defend architecture choices using real measurements.
- README benchmark sections are filled with real data.
