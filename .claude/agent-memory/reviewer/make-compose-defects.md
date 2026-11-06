---
name: make-compose-defects
description: Makefile and docker-compose defects found in the 00-skeleton Task 4 review (2026-11-06); check them when Makefile, compose or Dockerfile change
metadata:
  type: project
---

Defects found in the 00-skeleton Task 4 review on 2026-11-06:

- `$(shell command -v go)` with no shell metacharacters: GNU Make 3.81 execs `command` directly, not through `/bin/sh`. macOS has `/usr/bin/command`, so it works there. Without that file (many Linux distros), the result is empty and the Makefile picks the Docker path although Go is installed. Test: `PATH=/opt/homebrew/bin:/bin /usr/bin/make` on a 2-line Makefile. Fix: add `2>/dev/null`.
- `out="$$(gofmt -l .)"; [ -n "$$out" ]` drops the gofmt exit code. A parse error prints to stderr only, so lint passes. `go vet` does not cover `testdata/` or files with excluded build tags.
- `pg_isready -U campaign` with no `-h` uses the Unix socket. The postgres entrypoint runs a temporary socket-only server during first init, so the healthcheck can pass before TCP is up.
- `ARG CMD` before `RUN go mod download` makes the download layer differ per CMD value.
- `-v $(CURDIR)/backend:/src` without quotes breaks on paths with spaces.
- Compose v5 `docker compose port <svc> <port>` prints `:0` and exits 0 for an unpublished port. Use `docker inspect -f '{{json .HostConfig.PortBindings}}'` instead.

**Why:** these pass on the author's Mac and fail on a grader machine or on a first clean start.
**How to apply:** for each new Make target, `$(shell ...)` call, compose service or healthcheck, check these patterns again. Also see [[go-startup-defects]].
