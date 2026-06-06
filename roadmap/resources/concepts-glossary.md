# Beginner Concepts Glossary

- `Base62`: Encoding numbers using 62 characters (`0-9`, `a-z`, `A-Z`) to create short URL codes.
- `Snowflake ID`: Distributed unique ID strategy based on timestamp + machine + sequence bits.
- `Read-through cache`: First check cache; if miss, read DB and populate cache.
- `Rate limiting`: Restrict request frequency per IP/user to prevent abuse.
- `p99 latency`: 99th percentile response time; only 1% requests are slower.
- `Sharding`: Splitting data across multiple DB instances.
- `Eventual consistency`: Data may take short time to appear across systems.
- `At-least-once delivery`: Message may be delivered more than once, so consumers must handle duplicates.
- `Idempotent consumer`: Processing same message multiple times does not corrupt final state.
- `Observability`: Ability to understand system behavior using logs, metrics, and traces.
