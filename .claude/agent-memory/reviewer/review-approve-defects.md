---
name: review-approve-defects
description: Defects found in the slice 06 review (draft cache, async load race, count under row lock, stale first_name_runes after 422)
metadata:
  type: project
---

Patterns found in the slice 06 review on 2026-11-10. Check for them in later slices (07 send, 08 dashboard).

- Component cache keyed by id only ("reload only when id changes"). A regenerate keeps the same id but changes the server data, so the dialog shows stale content. Check every "loadedFor === id" style guard.
- Async load with no stale-response guard. Two opens in a row: the older response can overwrite the newer one. Check that code compares the id after `await`.
- A second pool query (audience count) inside a transaction that holds `SELECT ... FOR UPDATE`. With N >= pool size concurrent requests this starves the pool. Waiters also hit the 2 s statement_timeout and get 503, not the idempotent 200.
- After a server 422 because the audience changed, the browser keeps the old counts and `first_name_runes`. The counter says OK, the server says no.
- Trigger reads the parent status without `FOR SHARE`. Cascade delete of an approved parent passes the child trigger.
- Slice 08 fix round: an unmount test asserted `wrapper.html()` and `text()` after a late response. Vue stops render effects on unmount, so the DOM never changes. The assertion passes with the `if (unmounted) return` guard removed. Check that a "no state write" test can observe something that changes, such as a timer, a request count, or an exposed ref.
- Slice 08 fix round: a column header test used `toContain` on the whole tbody. Header order is only proven by `thead th` text equality.

**Why:** the tests pinned the happy concurrency path but not these.
**How to apply:** grep for row locks plus external calls, and for id-only caches in Vue components. Related: [[http-config-validation-defects]].
