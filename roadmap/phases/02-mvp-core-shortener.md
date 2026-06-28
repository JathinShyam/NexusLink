# Phase 2: MVP Core Shortener

## Goal

Build the smallest end-to-end working product: shorten URL and redirect.

## Duration

2-3 weeks

## Learning Focus

- API design basics
- Base62 encoding
- DB read/write patterns
- Unit testing handlers and core logic

## Build Tasks

- [x] Implement base62 encoder utility.
- [x] Implement ID generation (simple first, Snowflake-ready abstraction).
- [x] Build `POST /api/v1/shorten`.
- [x] Build `GET /{short_code}` with 302 redirect.
- [x] Add URL validation (allow only http/https).
- [x] Add error handling (`400`, `404`, `409` where relevant).
- [x] Write unit tests for base62 + handlers.

## Deliverables

- Able to create short links.
- Redirect endpoint resolves links correctly.
- Basic tests pass locally.

## Exit Criteria

- End-to-end flow works from API request to redirect in browser.
- New developer can run and test with a small README script.

## Suggested Weekly Split

- Week 1: base62 + shorten endpoint
- Week 2: redirect endpoint + validation + tests
- Week 3 (buffer): cleanup + bug fixes
