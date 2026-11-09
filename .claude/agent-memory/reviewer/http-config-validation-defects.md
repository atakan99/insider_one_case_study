---
name: http-config-validation-defects
description: Defects found in the 03 mock-provider review - NaN passes float range checks, json.Decoder accepts trailing data, reset reuses IDs, env_file leaks secrets
metadata:
  type: project
---

Recurring defect patterns from the slice 03 review (2026-11-09). Check them in each new HTTP handler or env loader.

- `strconv.ParseFloat` accepts "NaN" and "Inf". A check `f < 0 || f > 1` lets NaN through. Ask for `math.IsNaN` / `math.IsInf` checks.
- `json.NewDecoder(r).Decode` reads one value only. Trailing bytes (`{...} junk`) and padding after the object (past a MaxBytesReader limit) get accepted. Ask for `dec.More()` or a second Decode that must return io.EOF.
- A reset endpoint that sets an ID counter back to 0 reuses IDs that clients already stored.
- `env_file: .env` on a service that needs no secrets gives it ANTHROPIC_API_KEY and any shared HTTP_ADDR.
- Concurrency tests with all failure rates at 0 do not prove counter equations that include lost/slow paths.

**Why:** These passed `make verify` and the implementer's own tests.
**How to apply:** Test these inputs by hand in each review of config loaders and JSON handlers. See [[go-startup-defects]].
