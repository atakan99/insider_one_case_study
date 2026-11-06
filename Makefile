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

.PHONY: up down test lint verify lint-docs

up:
	docker compose up --build -d --wait
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8080/readyz
	@echo
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8081/healthz
	@echo

down:
	docker compose down

test:
	$(GO) test ./...
	if [ -d frontend ]; then pnpm -C frontend test:run; fi

lint:
	$(GO) vet ./...
	@out="$$($(GOFMT) -l .)" || exit 1; \
		if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

verify: lint test
	docker compose config -q

# Warnings only. Not part of verify.
lint-docs:
	git ls-files '*.md' | grep -v -e '^frontend/' -e '^shadcn-vue-admin-main/' \
		| xargs python3 -B .claude/skills/ste-writing/scripts/ste_check.py
