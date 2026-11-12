# Recurring "unknown" in every run

- Session: `9d719346-fa4c-415a-b51f-0be34cb8e5ee`
- Model: `claude-opus-5-5`
- Time (UTC): 2026-11-12 06:53:08 → 2026-11-12 06:53:58
- Subagent runs: 0

---

## 👤 User  <sub>2026-11-12 06:53:08</sub>

what is the reason i keep seein unknown in every run

**🔧 Bash** _Read progress log and search for "unknown"_
```bash
tail -60 tasks/progress.md; grep -rniI "unknown" backend frontend/src --include=*.go --include=*.vue --include=*.ts | grep -v node_modules | head -60
```

<details><summary>Result</summary>

```
- The progress stream has no SSE keepalive comment.

## 2026-11-10 — 08-dashboard-metrics
- Done: slice 08 dashboard and metrics. Added the dashboard store and the health checker. Added the two dashboard endpoints, the api `/metrics`, the worker `/metrics` and the Redis check in `/readyz`. Added the `/dashboard` page. `/` goes to `/dashboard`. Added D-068 to D-082 and the architecture Flow step 5. The backlog item has the spec and the acceptance lines.
- Verified with: a compose run with `-p cc-verify` on ports 13000, 18080, 18081 and 18082, with `LLM_MODE=mock`. The example brief gave campaign 1 with SMS 96 sent and 5 unknown, email 140 sent and 8 unknown. `GET /api/dashboard/stats` matched `GET /api/campaigns/1/progress`. The highest throughput value was 21.2 per second. `GET /api/dashboard/health` gave four `ok` and pending 0. `docker compose stop mock-provider` made `provider` `down` at once, and `start` made it `ok` in 1 s. Both `/metrics` returned the metrics of `goal.md`. Headless Chrome at 1280 px showed the five parts in the light theme and the dark theme with no horizontal scroll.
- Fix wave after the final review: the grid is `grid grid-cols-1 gap-6 lg:grid-cols-2` with `min-w-0` on both cards. At 360 px, `document.documentElement.scrollWidth` is 360 in the light theme and the dark theme. NFR-3 holds. The x-axis labels of the chart no longer overlap. The page sets `stats` and `health` independently, so a slow `health` call does not delay the statistics. The api Redis client and a second worker client for the stream collector honor context deadlines (D-083). A new test covers the labels of the api `/metrics` (FR-19).
- NFR-4: 50,000 send jobs in 5 campaigns, 4,276 with `sent_at` in the last 10 minutes. Five calls of `GET /api/dashboard/stats` took 25, 19, 19, 18 and 19 ms. The limit is 300 ms.
- Next: user review of slice 08. The user sets `08-dashboard-metrics.passes`.
- Open questions:
  - `XLEN` of the streams does not fall after a send, because the wor
… [5554 more chars truncated]
```

</details>

**🔧 Bash** _Find where unknown status is set_
```bash
grep -rniI "unknown" backend --include='*.go' | grep -v _test | head -40; ls backend/cmd; grep -niE "unknown|timeout|ambig" docs/decisions-log.md | head -30
```

<details><summary>Result</summary>

```
backend/internal/llm/llm.go:87:		return nil, fmt.Errorf("llm: unknown mode %q", cfg.LLMMode)
backend/internal/segment/validate.go:104:			v.fail(join(path, truncate(k, 40)), "unknown key")
backend/internal/segment/validate.go:184:		v.fail(join(path, "field"), "unknown field %q", truncate(name, 40))
backend/internal/content/validate.go:56:	// Unknown keys, in a stable order.
backend/internal/content/validate.go:57:	var unknown []string
backend/internal/content/validate.go:66:			unknown = append(unknown, k)
backend/internal/content/validate.go:69:	sort.Strings(unknown)
backend/internal/content/validate.go:70:	for _, k := range unknown {
backend/internal/content/validate.go:71:		fail(truncate(k, 40), "unknown key")
backend/internal/content/vars.go:19:			return 0, errors.New("only the variable " + VarFirstName + " is allowed; found an unknown variable or a stray brace")
backend/internal/httpapi/delivery.go:157:			delivery.StatusFailed, delivery.StatusUnknown, delivery.StatusSkipped:
backend/internal/dashboard/types.go:15:	Unknown    int64 `json:"unknown"`
backend/internal/dashboard/types.go:46:	Unknown     int64  `json:"unknown"`
backend/internal/dashboard/store.go:67:		case delivery.StatusUnknown:
backend/internal/dashboard/store.go:68:			t.Unknown = n
backend/internal/dashboard/store.go:119:FROM send_jobs WHERE status IN ('failed', 'unknown', 'skipped')
backend/internal/dashboard/store.go:158:  coalesce(a.total, 0), coalesce(a.sent, 0), coalesce(a.failed, 0), coalesce(a.unknown, 0),
backend/internal/dashboard/store.go:166:         count(*) FILTER (WHERE status = 'unknown') AS unknown,
backend/internal/dashboard/store.go:176:		if err := rows.Scan(&c.ID, &c.Description, &c.Status, &c.Total, &c.Sent, &c.Failed, &c.Unknown, &c.DurationMS); err != nil {
backend/internal/campaign/check.go:60:		var unknown []string
backend/internal/campaign/check.go:63:				unknown = append(unknown, k)
backend/internal/campaign/check.go:66:		sort.Strings(unknown)
backend/internal/campaign/chec
… [7591 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-12 06:53:30</sub>

Now let me check the mock provider's failure rates.

**🔧 Bash** _Check mock provider failure modes and config_
```bash
grep -rniE "slow|lost|rate|percent|prob|rand" backend/cmd/mockprovider/*.go backend/internal/mockprovider/*.go 2>/dev/null | grep -v _test | head -30; grep -niE "MOCK|FAIL|SLOW|LOST|SEND_TIMEOUT" docker-compose*.yml compose*.yml .env* 2>/dev/null
```

<details><summary>Result</summary>

```
backend/cmd/mockprovider/main.go:38:		"rate_per_sec", cfg.RatePerSec, "burst", cfg.Burst,
backend/cmd/mockprovider/main.go:39:		"error_rate", cfg.ErrorRate, "slow_rate", cfg.SlowRate, "lost_rate", cfg.LostRate,
backend/cmd/mockprovider/main.go:40:		"slow_delay", cfg.SlowDelay.String(), "latency_max", cfg.LatencyMax.String(),
backend/cmd/mockprovider/main.go:41:		"random", cfg.Random)
backend/cmd/mockprovider/main.go:46:	// No WriteTimeout: slow responses wait for cfg.SlowDelay, and the server
backend/cmd/mockprovider/main.go:51:			RatePerSec: cfg.RatePerSec,
backend/cmd/mockprovider/main.go:53:			ErrorRate:  cfg.ErrorRate,
backend/cmd/mockprovider/main.go:54:			SlowRate:   cfg.SlowRate,
backend/cmd/mockprovider/main.go:55:			LostRate:   cfg.LostRate,
backend/cmd/mockprovider/main.go:56:			SlowDelay:  cfg.SlowDelay,
backend/cmd/mockprovider/main.go:58:			Random:     cfg.Random,
backend/internal/mockprovider/failure.go:5:	"math/rand/v2"
backend/internal/mockprovider/failure.go:17:	modeSlow
backend/internal/mockprovider/failure.go:18:	modeLost
backend/internal/mockprovider/failure.go:30:	case modeSlow:
backend/internal/mockprovider/failure.go:31:		return "slow"
backend/internal/mockprovider/failure.go:32:	case modeLost:
backend/internal/mockprovider/failure.go:33:		return "lost"
backend/internal/mockprovider/failure.go:41:// empty, so the random picker decides.
backend/internal/mockprovider/failure.go:54:	case "slow":
backend/internal/mockprovider/failure.go:55:		return modeSlow, true, nil
backend/internal/mockprovider/failure.go:56:	case "lost":
backend/internal/mockprovider/failure.go:57:		return modeLost, true, nil
backend/internal/mockprovider/failure.go:59:	return modeNone, false, fmt.Errorf("X-Mock-Fail must be one of none, 429, 500, 503, slow, lost, got %q", header)
backend/internal/mockprovider/failure.go:62:// picker selects random failure modes and latencies from one seeded source.
backend/internal/mockprovider/failure.go:65:	rnd                           *ra
… [363 more chars truncated]
```

</details>

**🔧 Bash** _Find default failure rates_
```bash
grep -rniE "SLOW_RATE|LOST_RATE|ERROR_RATE|SLOW_DELAY|SEND_TIMEOUT" --include='*.go' --include='*.yml' --include='*.yaml' --include='.env*' --include='*.md' . 2>/dev/null | grep -v node_modules | grep -v _test | head -20
```

<details><summary>Result</summary>

```
.env.example:26:# MOCK_ERROR_RATE=0.03
.env.example:27:# MOCK_SLOW_RATE=0.03
.env.example:28:# MOCK_LOST_RATE=0.01
.env.example:29:# MOCK_SLOW_DELAY=5s
.env.example:35:# SEND_TIMEOUT must be shorter than MOCK_SLOW_DELAY. SEND_CONSUMERS is for each channel.
.env.example:38:# SEND_TIMEOUT=3s
docker-compose.yml:95:      MOCK_ERROR_RATE: ${MOCK_ERROR_RATE:-0.03}
docker-compose.yml:96:      MOCK_SLOW_RATE: ${MOCK_SLOW_RATE:-0.03}
docker-compose.yml:97:      MOCK_LOST_RATE: ${MOCK_LOST_RATE:-0.01}
docker-compose.yml:98:      MOCK_SLOW_DELAY: ${MOCK_SLOW_DELAY:-5s}
docker-compose.yml:120:      SEND_TIMEOUT: ${SEND_TIMEOUT:-3s}
tasks/progress.md:65:- Open questions: four Minor items stay open. The decoder accepts data after the JSON object. The 10 s shutdown timeout is too short for `MOCK_SLOW_DELAY` over 10 s. Two tests can be stronger (slow disconnect, bucket after a `429`). Slice 07 must not send `Idempotency-Key` or `X-Idempotency-Key`.
tasks/progress.md:115:  - A `Retry-After` longer than `SEND_TIMEOUT` plus 30 s can set a waiting send job to `unknown`. The mock provider sends 1 s.
backend/cmd/mockprovider/main.go:39:		"error_rate", cfg.ErrorRate, "slow_rate", cfg.SlowRate, "lost_rate", cfg.LostRate,
backend/cmd/mockprovider/main.go:40:		"slow_delay", cfg.SlowDelay.String(), "latency_max", cfg.LatencyMax.String(),
backend/internal/config/config.go:207:		{"MOCK_ERROR_RATE", &cfg.ErrorRate},
backend/internal/config/config.go:208:		{"MOCK_SLOW_RATE", &cfg.SlowRate},
backend/internal/config/config.go:209:		{"MOCK_LOST_RATE", &cfg.LostRate},
backend/internal/config/config.go:223:		errs = append(errs, fmt.Errorf("the sum of MOCK_ERROR_RATE, MOCK_SLOW_RATE and MOCK_LOST_RATE must be 1 or less, got %g", sum))
backend/internal/config/config.go:229:		{"MOCK_SLOW_DELAY", &cfg.SlowDelay},
```

</details>

## 🤖 Claude  <sub>2026-11-12 06:53:58</sub>

The "unknown" results are on purpose. The mock provider fails some requests in a way that doesn't tell the worker whether the message got delivered, and the worker marks those requests `unknown`. It is not a bug.

**Why it happens**

The mock provider picks a random failure for part of the requests. The defaults are in [docker-compose.yml:95-98](docker-compose.yml#L95-L98):

| Variable | Default | What happens | Result |
|---|---|---|---|
| `MOCK_ERROR_RATE` | 3% | `429` / `500` / `503` | Retried, then `sent` or `failed` |
| `MOCK_SLOW_RATE` | 3% | The provider delivers the message but waits 5 s (`MOCK_SLOW_DELAY`). The worker gives up at 3 s (`SEND_TIMEOUT`) | **`unknown`** |
| `MOCK_LOST_RATE` | 1% | The provider delivers the message, then drops the connection | **`unknown`** |

So about 4% of each run should end up `unknown`. Your slice 08 run matches that: 5 of 101 SMS and 8 of 148 emails.

**Why the worker doesn't retry them**

This is D-019, which you chose, and D-048 in [docs/decisions-log.md](docs/decisions-log.md). After a timeout or a dropped connection, the message may already be delivered. A retry could then send it a second time. The worker records `unknown` and stops, which follows the "never send twice" rule. D-063 adds that the worker only gives `unknown` when the request was actually written to the provider. A connection error before the write is still retried.

**To get runs without `unknown`**

Set these in `.env`, then run `make up`:

```
MOCK_SLOW_RATE=0
MOCK_LOST_RATE=0
```

Raising `SEND_TIMEOUT` above `MOCK_SLOW_DELAY` (for example `SEND_TIMEOUT=6s`) also works, but only for the slow cases. Lost connections still give `unknown`. The `.env.example` comment expects `SEND_TIMEOUT` to stay below `MOCK_SLOW_DELAY` so the timeout path gets exercised.

For a demo I'd keep the defaults: the `unknown` count shows that the worker handles ambiguous failures.

