---
name: dashboard-metrics-defects
description: Slice 08 final review defects - go-redis ignores ctx deadlines, grid items without min-w-0 overflow at 360 px, allSettled couples endpoints, missing api FR-19 scrape
metadata:
  type: project
---

Defect patterns from the slice 08 final review (2026-11-10). Check them again in later slices.

- go-redis v9 (v9.23) has `ContextTimeoutEnabled` false by default. Then a `context.WithTimeout` does not bound socket reads. ReadTimeout defaults to 5 s. A "1 s check" on Redis can take 5 s when Redis stalls. `cmd/api` and `cmd/worker` build the client with `redis.ParseURL` and do not set the option.
  **Why:** the per-task reviews and the unit tests used fake pingers or a closed client, which fail fast. Only a stalled TCP peer shows the bug.
  **How to apply:** for each Redis call with a timeout, check the client options. Ask for a test with a listener that accepts and never answers.
- A Tailwind `grid` without `grid-cols-1` or `min-w-0` on its items lets a table inside an `overflow-x-auto` div widen the page. At 360 px the page scrolled 6 px.
  **How to apply:** for each new grid or flex parent of a table or chart, check `min-w-0` or `minmax(0,1fr)` columns.
- `Promise.allSettled` over two endpoints makes the fast one wait for the slow one. This broke the D-069 claim "a slow health check does not delay the statistics".
- A spec line "a test scrapes both /metrics" got a test for the worker only. Check each "both" in the requirements against the test list.

Related: [[go-startup-defects]], [[http-config-validation-defects]].
