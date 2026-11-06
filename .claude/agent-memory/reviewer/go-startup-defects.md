---
name: go-startup-defects
description: Go process startup defects found in the 00-skeleton Task 3 review (2026-11-06); check them in each new cmd/ binary
metadata:
  type: project
---

Defects found in `backend/cmd/api/main.go` in the 00-skeleton Task 3 review on 2026-11-06:

- The startup `pool.Ping(ctx)` had no timeout, and it ran before `ListenAndServe`. pgx has no default connect timeout. With a blackholed DB host (for example `10.255.255.1`), the process did not listen for the OS TCP timeout (about 75 s on macOS). Thus `/healthz` failed.
- The `select` on `errc` and `ctx.Done()` can drop a listen error when a signal arrives at the same time. The process then exits 0.
- `signal.NotifyContext` stayed active during shutdown, so a second SIGINT did not force an exit.

Checked and correct: pgx redacts the password in parse errors (`xxxxx`). pgx connect errors show user, database and host, but not the password.

**Why:** the plan says a failed initial DB connection must not stop the API from serving `/readyz`.
**How to apply:** for each new binary (worker, seed), check the timeouts of blocking calls before the listener starts. Test with a blackholed host, not only with a refused port. On macOS, a Python IPv4 listener does not block a Go `[::]` bind, so use `nc -l` on the same family to test "port in use".
