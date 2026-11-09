---
name: test-db-harness-defects
description: Defects found in the 02-seed-users Task 1 review (2026-11-09) of storetest.NewDB, migrate tests and the make test Postgres; check them when DB tests or the test target change
metadata:
  type: project
---

Defects found in the 02-seed-users Task 1 review on 2026-11-09:

- `url.Parse(base)` then `u.Path = "/" + name`: a `?dbname=test` query parameter overrides the path in pgx. Thus the test runs in the shared database, not in the new one. A keyword/value DSN (`host=x dbname=y`) gives the string `/t_x`. Verified with `go run`.
- Constraint tests that check only `err == nil` pass for any error. Ask for a SQLSTATE check (`23514` check, `23505` unique, `23503` FK) with `pgconn.PgError`. Same pattern as the vacuous asserts in [[frontend-trim-defects]].
- `make test` with Go in a container still publishes `127.0.0.1:55432`. A busy host port fails the run, although that mode does not use the port.
- Ctrl-C during `go test` leaves `cc-test-pg` running. The next run removes it with `docker rm -f`, so this is acceptable.

Checked and correct: `pg_isready -h 127.0.0.1` waits for the real server, not the socket-only init server. `docker run` returns the exit code of the container, so `rc=$$?` keeps the test result.

**Why:** these harness defects make DB tests pass while they test the wrong thing.
**How to apply:** in each later slice that adds DB tests, check the error assertions and the URL that the harness returns. Also see [[make-compose-defects]].
