# Phase 1: Orientation And Foundation

## Goal

Set up your development environment and understand the system before building features.

## Duration

1-2 weeks

## Learning Focus

- Basic HTTP flow (request -> handler -> response)
- Go project structure basics
- Docker Compose fundamentals
- PostgreSQL basics (table, insert, select)

## Build Tasks

- [ ] Create folder structure matching planned architecture.
- [ ] Initialize Go module and basic app entrypoint.
- [ ] Add config loader and structured logging.
- [ ] Create Docker Compose with app + PostgreSQL.
- [ ] Create first migration for `url_mappings`.
- [ ] Add `/health` endpoint.

## Deliverables

- App starts locally.
- Postgres is running via Docker.
- Migration can run successfully.
- Health endpoint returns 200.

## Exit Criteria

- You can explain request flow from browser to handler.
- You can run the project from scratch on your machine in under 10 minutes.

## Common Beginner Traps

- Spending too much time perfecting folder structure before running code.
- Writing feature code before confirming DB and app boot process.
- Skipping migration scripts.
