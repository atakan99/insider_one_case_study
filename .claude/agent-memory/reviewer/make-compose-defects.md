---
name: make-compose-defects
description: Makefile, docker-compose, Dockerfile and nginx defects found in the 00 Task 4 and 01 Task 3 reviews (2026-11-06); check them when Makefile, compose or Dockerfile change
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

Defects found in the 01-frontend-trim Task 3 review on 2026-11-06:

- `docker run -v frontend:/app -v /app/node_modules ... pnpm install`: pnpm put its store (432 MB) in `/app/.pnpm-store`, thus in the host's `frontend/`. Neither `.gitignore` nor `.dockerignore` listed it. Check for new untracked folders after each Docker run with a bind mount.
- A container that runs as root writes root-owned files (`dist/`, the store) into a bind mount on Linux. Docker Desktop on macOS hides this.
- nginx with `try_files $uri $uri/ ...` and a port mapping: a request for a folder without a slash (`/assets`) gets `301 Location: http://host:<container port>/`. Use `absolute_redirect off`.
- nginx `expires` plus `add_header Cache-Control` sends two Cache-Control headers.
