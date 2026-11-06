# Decisions

## Stack

- **Go for the backend.** Workers and rate limiters need cheap concurrency, and each service builds to one static binary.
- **Postgres as the datastore.** Segment queries need SQL. Postgres also holds the delivery status and the idempotency keys.

## Audience

- **Segment rules compile to parameterized SQL.** The backend validates the LLM's JSON rule and builds SQL with `$n` placeholders. Thus no user value or LLM text goes into the SQL text. `sqlc` handles the fixed queries.

## Sending

- **Redis Streams as the queue.** Consumer groups give at-least-once delivery and retries for pending entries.
