# NexusLink Documentation

Deep-dive, decision-oriented material that lives behind the top-level
[README](../README.md). The README remains the single source of truth for
requirements, API, and benchmarks; this folder explains the *why* and the *how*.

## Contents

| Path | What it covers |
|------|----------------|
| [`adr/`](./adr/) | Architecture Decision Records — one per significant choice |
| [`architecture/`](./architecture/) | Component and request-flow notes |
| [`benchmarks/`](./benchmarks/) | Load-test methodology and results (Phase 4+) |
| [`runbook.md`](./runbook.md) | Local development, operations, and failure-mode playbook |

Related, outside this folder: the machine-readable REST contract at
[`api/openapi.yaml`](../api/openapi.yaml).

## Where to start

- **New to the project?** Read the [top-level README](../README.md), then
  [`architecture/`](./architecture/).
- **Reviewing a decision?** Start at the [ADR index](./adr/README.md).
- **Running it?** See the [runbook](./runbook.md).
