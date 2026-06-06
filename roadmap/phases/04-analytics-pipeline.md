# Phase 4: Analytics Pipeline

## Goal

Track click events asynchronously without slowing redirects.

## Duration

2-3 weeks

## Learning Focus

- Event-driven design
- Kafka producer/consumer basics
- ClickHouse data modeling
- Eventual consistency and idempotency

## Build Tasks

- [ ] Publish click event on redirect path (non-blocking).
- [ ] Create Kafka topic and producer retry policy.
- [ ] Build consumer service for batched writes.
- [ ] Insert analytics data into ClickHouse.
- [ ] Build analytics API for per-link stats.
- [ ] Build minimal dashboard page (clicks over time + top referrers).

## Deliverables

- Click appears in analytics within a defined SLA (for example, <10 seconds).
- Redirect remains fast even if analytics backend is slow.

## Exit Criteria

- You can show click flow: redirect -> Kafka -> consumer -> ClickHouse -> API.
- Consumer behavior under duplicates is documented.

## Suggested Weekly Split

- Week 1: producer and topic setup
- Week 2: consumer and ClickHouse writes
- Week 3: analytics API + dashboard
