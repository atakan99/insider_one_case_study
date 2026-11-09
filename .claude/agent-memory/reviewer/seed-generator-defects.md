---
name: seed-generator-defects
description: Defects found in the 02-seed-users Task 2 and Task 3 reviews (2026-11-09) of the seed generator and seed.Load; check them when seed, faker, date or loader code changes
metadata:
  type: project
---

Defects found in the 02-seed-users Task 2 review on 2026-11-09:

- `gofakeit.New(0)` in v7 uses a crypto random seed. Thus a seed value of 0 breaks determinism. Verified in the module source of v7.18.1.
- `today.AddDate(-years, 0, -d)` on Feb 29 gives Mar 1 of a non-leap year. The age is then `years-1` (17 for 18). Verified with `go run`.
- A direct import marked `// indirect` in go.mod. `go mod tidy` changes it.
- A default-value test that checks `>=` instead of `==`.

Defects found in the Task 3 review (seed.Load) on 2026-11-09:

- A SQL test restated the goal.md example rule in a different form (`last_purchase_at < last_cart_add_at` instead of `< now - 30d`). The edge-case user satisfies both forms, so the test passes either way.
- No test pins the kurus-to-numeric scale. `Exp: 0` instead of `-2` passes all tests, because summary and event amounts scale together.

Defects found in the final slice review on 2026-11-09:

- A compose comment ("a restart does not reload them") was false when a documented env toggle (`SEED_RESET=true` in `.env`) was set. Each `docker compose up` starts the one-shot service again.
- A decisions-log entry counted statements ("a count", "five statements") that did not match the code (EXISTS, TRUNCATE, two copies, two setval).
- A deferred note about a non-obvious Makefile choice (test network, fixed port) had no decisions-log entry. This breaks a CLAUDE.md hard rule, so it is not optional.

**Why:** these defects pass with the fixed test seed (42), a fixed test date and a planted edge-case user, so the tests do not show them.
**How to apply:** for each generator or random-source change, ask about seed 0, leap days and exact counts. For each SQL test of a spec rule, compare the predicate word for word with the spec. Also see [[test-db-harness-defects]].
