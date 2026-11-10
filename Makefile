# Use local Go when it is installed. Otherwise run Go in a container (D-004).
# The 2>/dev/null makes Make run this through /bin/sh, so the `command` builtin works.
ifeq ($(shell command -v go 2>/dev/null),)
DOCKER_GO = docker run --rm -v "$(CURDIR)/backend:/src" -w /src golang:1.27-alpine
GO = $(DOCKER_GO) go
GOFMT = $(DOCKER_GO) gofmt
# Tests join the cc-test network to reach the test Postgres by name.
GO_TEST = docker run --rm --network cc-test -e TEST_DATABASE_URL -e TEST_REDIS_URL -v "$(CURDIR)/backend:/src" -w /src golang:1.27-alpine go
TEST_DATABASE_URL = postgres://test:test@cc-test-pg:5432/test?sslmode=disable
TEST_REDIS_URL = redis://cc-test-redis:6379/0
else
GO = cd backend && go
GOFMT = cd backend && gofmt
GO_TEST = cd backend && go
TEST_DATABASE_URL = postgres://test:test@127.0.0.1:55432/test?sslmode=disable
TEST_REDIS_URL = redis://127.0.0.1:56379/0
endif

# Same for pnpm. The anonymous /app/node_modules volume keeps Linux packages
# out of the host's frontend/node_modules. The pnpm store stays in the
# container (/tmp). The container runs as root, so at the end it gives the
# files that it wrote (dist/, src/types/) back to the host user.
ifeq ($(shell command -v pnpm 2>/dev/null),)
PNPM = docker run --rm -v "$(CURDIR)/frontend:/app" -v /app/node_modules -e CI=true \
	-e HOST_ID=$(shell id -u):$(shell id -g) -w /app node:24-alpine \
	sh -c 'corepack enable && pnpm install --frozen-lockfile --store-dir /tmp/pnpm-store && pnpm "$$@"; rc=$$?; \
	chown -R "$$HOST_ID" dist src/types 2>/dev/null; exit $$rc' pnpm
else
PNPM = pnpm -C frontend
endif

.PHONY: up down test lint verify lint-docs

up:
	docker compose up --build -d --wait
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8080/readyz
	@echo
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8081/healthz
	@echo
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8082/healthz
	@echo
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 -o /dev/null -w "frontend %{http_code}\n" localhost:3000/

down:
	docker compose down

# Starts a temporary Postgres and Redis, runs the Go tests and removes the containers.
# The exit code is the exit code of the Go tests.
test:
	@docker network inspect cc-test >/dev/null 2>&1 || docker network create cc-test >/dev/null
	@docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1 || true
	docker run -d --name cc-test-pg --network cc-test \
		-e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test \
		-p 127.0.0.1:55432:5432 postgres:17-alpine >/dev/null
	docker run -d --name cc-test-redis --network cc-test \
		-p 127.0.0.1:56379:6379 redis:8-alpine >/dev/null
	@i=0; until docker exec cc-test-pg pg_isready -U test -d test -h 127.0.0.1 >/dev/null 2>&1; do \
		i=$$((i+1)); if [ $$i -gt 60 ]; then echo "test Postgres not ready"; docker rm -f cc-test-pg cc-test-redis >/dev/null; exit 1; fi; \
		sleep 1; done
	@i=0; until docker exec cc-test-redis redis-cli ping >/dev/null 2>&1; do \
		i=$$((i+1)); if [ $$i -gt 60 ]; then echo "test Redis not ready"; docker rm -f cc-test-pg cc-test-redis >/dev/null; exit 1; fi; \
		sleep 1; done
	@export TEST_DATABASE_URL='$(TEST_DATABASE_URL)' TEST_REDIS_URL='$(TEST_REDIS_URL)'; $(GO_TEST) test ./...; rc=$$?; \
		docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1; exit $$rc
	$(PNPM) test:run

lint:
	$(GO) vet ./...
	@out="$$($(GOFMT) -l .)" || exit 1; \
		if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	CI=true $(PNPM) lint

verify: lint test
	$(PNPM) build
	docker compose config -q

# Warnings only. Not part of verify.
lint-docs:
	git ls-files '*.md' | grep -v -e '^frontend/' -e '^shadcn-vue-admin-main/' \
		| xargs python3 -B .claude/skills/ste-writing/scripts/ste_check.py
