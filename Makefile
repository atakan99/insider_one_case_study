# Use local Go when it is installed. Otherwise run Go in a container (D-004).
# The 2>/dev/null makes Make run this through /bin/sh, so the `command` builtin works.
ifeq ($(shell command -v go 2>/dev/null),)
DOCKER_GO = docker run --rm -v "$(CURDIR)/backend:/src" -w /src golang:1.27-alpine
GO = $(DOCKER_GO) go
GOFMT = $(DOCKER_GO) gofmt
else
GO = cd backend && go
GOFMT = cd backend && gofmt
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
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 -o /dev/null -w "frontend %{http_code}\n" localhost:3000/

down:
	docker compose down

test:
	$(GO) test ./...
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
