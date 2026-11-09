# 03 — Mock Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `mock-provider` accepts SMS and email sends like Twilio or SendGrid. It returns random and forced `429`, `5xx`, slow and lost responses. It keeps an audit of all deliveries, which includes duplicate deliveries.

**Architecture:** This slice adds one package and changes one binary:

```
internal/config        LoadMockProvider(getenv) -> MockProviderConfig
internal/mockprovider  New(Options) http.Handler: limiter, failure modes, store, handlers
cmd/mockprovider       reads the config, serves mockprovider.New
```

The handler does the steps in "Order of checks" in `goal.md`. All time and all waits go through `Options.Now` and `Options.Sleep`. Thus tests run with no real waits.

**Tech Stack:** Go 1.27, `golang.org/x/time/rate` (new, D-020), `math/rand/v2`, `net/http`, `net/http/httptest`.

**Spec:** `docs/specs/03-mock-provider/goal.md` and `requirements.md`.

## Global Constraints

- The provider does no dedup. It does not read `Idempotency-Key` (D-018).
- `500` and `503` store nothing. `slow` and `lost` store the message (D-019).
- Count characters with `utf8.RuneCountInString`. SMS `body`: 1 to 160. Email `subject`: 1 to 80. `reference`: 0 to 128.
- Defaults: `MOCK_RATE_PER_SEC=100`, `MOCK_RATE_BURST=10`, `MOCK_ERROR_RATE=0.03`, `MOCK_SLOW_RATE=0.03`, `MOCK_LOST_RATE=0.01`, `MOCK_SLOW_DELAY=5s`, `MOCK_LATENCY_MAX=50ms`, `HTTP_ADDR=:8081`.
- All error bodies are `{"error":{"code":"...","message":"..."}}`. Codes: `invalid_request`, `rate_limited`, `provider_error`, `unavailable`.
- No test sleeps in real time, and no test depends on the default rates (NFR-2, NFR-3).
- `internal/mockprovider` imports nothing from `internal/` (NFR-6).
- No test reads `ANTHROPIC_API_KEY`.
- Do not commit after each task. The user commits the slice after review, as in slices 00 to 02.

## Review Focus

1. A client times out during a `slow` wait → the provider must still store the message. If the handler uses the request context for the wait, the ambiguous case disappears. Pinned by Task 3 `TestSlow_StoresAfterClientLeaves`.
2. The same `reference` from many goroutines at once → `duplicate_deliveries` must be `delivered - 1`, with no lost update. Pinned by Task 2 `TestConcurrent_DuplicateCount`.
3. JSON with wrong types, for example `"to": 123`, or a body that is not JSON → `400 invalid_request`, not `500`. A `500` here would look like a random provider error to the worker. Pinned by Task 2 `TestSend_BadJSON`.
4. An empty `reference` on many messages → these messages are not duplicates. Pinned by Task 2 `TestSend_EmptyReferenceNotDuplicate`.
5. `limit=0`, `limit=1001`, `offset=-1`, `channel=push` on `GET /v1/messages` → `400`. Large limits must not return the full store. Pinned by Task 2 `TestMessages_BadQuery`.

---

### Task 1: `config.LoadMockProvider`

**Files:**
- Modify: `backend/internal/config/config.go`, `backend/internal/config/config_test.go`

**Interfaces:**
- Produces:

```go
type MockProviderConfig struct {
    HTTPAddr   string
    RatePerSec float64
    Burst      int
    ErrorRate  float64
    SlowRate   float64
    LostRate   float64
    SlowDelay  time.Duration
    LatencyMax time.Duration
    Random     uint64 // from MOCK_RANDOM, or from the clock if it is not set
}
func LoadMockProvider(getenv func(string) string) (MockProviderConfig, error)
```

- [ ] **Step 1: Write the failing tests in `config_test.go`.**
  - `TestLoadMockProvider_Defaults`: empty env gives the defaults from Global Constraints.
  - `TestLoadMockProvider_Values`: each `MOCK_*` var and `HTTP_ADDR` set to a valid value reaches its field. `MOCK_RANDOM=7` gives `Random == 7`.
  - `TestLoadMockProvider_Errors`: table test. Each case gives an error whose text contains the var name:
    - `MOCK_ERROR_RATE=1.5`, `MOCK_SLOW_RATE=-0.1`, `MOCK_LOST_RATE=abc`
    - `MOCK_ERROR_RATE=0.5` with `MOCK_SLOW_RATE=0.4` and `MOCK_LOST_RATE=0.2`. The sum is over 1. The text must contain `sum`.
    - `MOCK_SLOW_DELAY=5`, `MOCK_LATENCY_MAX=-1s`
    - `MOCK_RATE_PER_SEC=0`, `MOCK_RATE_BURST=0`, `MOCK_RANDOM=-3`

  Run: `cd backend && go test ./internal/config/...`
  Expected: compile error, because `LoadMockProvider` does not exist.

- [ ] **Step 2: Write `LoadMockProvider`.** Use the same pattern as `LoadSeed`: collect all errors, then `errors.Join`. Use `time.ParseDuration` for durations. Reject a negative duration. A rate per second must be greater than 0, and a burst must be at least 1. If `MOCK_RANDOM` is not set, use `uint64(time.Now().UnixNano())`.

  Run the tests again. Expected: PASS.

---

### Task 2: Handlers, validation and the message store (normal path)

**Files:**
- Create: `backend/internal/mockprovider/options.go`, `store.go`, `validate.go`, `server.go`
- Create: `backend/internal/mockprovider/store_test.go`, `server_test.go`

**Interfaces:**
- Consumes: nothing from Task 1. Task 4 maps the config to `Options`.
- Produces:

```go
type Options struct {
    RatePerSec float64
    Burst      int
    ErrorRate, SlowRate, LostRate float64
    SlowDelay, LatencyMax         time.Duration
    Random     uint64
    Now        func() time.Time     // nil means time.Now
    Sleep      func(time.Duration)  // nil means time.Sleep
}

type Message struct {
    ID         string    `json:"id"`
    Channel    string    `json:"channel"` // "sms" or "email"
    To         string    `json:"to"`
    Subject    string    `json:"subject"`
    Body       string    `json:"body"`
    Reference  string    `json:"reference"`
    ReceivedAt time.Time `json:"received_at"`
}

type ChannelStats struct {
    Requests            int64 `json:"requests"`
    Delivered           int64 `json:"delivered"`
    Rejected            int64 `json:"rejected"`
    RateLimited         int64 `json:"rate_limited"`
    Errors              int64 `json:"errors"`
    Slow                int64 `json:"slow"`
    Lost                int64 `json:"lost"`
    DuplicateDeliveries int64 `json:"duplicate_deliveries"`
}

// New returns the provider handler with all routes from goal.md,
// including GET /healthz.
func New(o Options) http.Handler
```

The store type is internal to the package. Task 3 adds the limiter and failure modes to the same handler.

- [ ] **Step 1: Write the failing tests in `server_test.go`.** Each test uses `httptest.NewServer(New(opts))`. `opts` has all rates at 0, `LatencyMax: 0`, a high `RatePerSec` and `Burst`, and a fixed `Now`. Write a helper `post(t, url, path, body string, header map[string]string) *http.Response`.
  - `TestSend_SMSAccepted`, `TestSend_EmailAccepted`: `202`, body `{"id": <non-empty>, "status": "accepted"}`. `GET /v1/messages` then has the message with `received_at` equal to `Now()` in UTC. Checks FR-1.
  - `TestSend_Invalid`: table test, each gives `400` with `code: invalid_request`, and `total` stays 0. Cases: SMS `body` of 161 `a`, empty `body`, `to` = `05321234567`, missing `to`. Email `subject` of 81 runes, empty `subject`, `to` = `not-an-email`, `to` = `Ali <ali@example.com>`, empty `body`. `reference` of 129 runes. Checks FR-2.
  - `TestSend_RuneCount`: SMS `body` of 160 `ş` gives `202`. 161 `ş` gives `400`. Email `subject` of 80 `ğ` gives `202`. Checks FR-3.
  - `TestSend_BadJSON`: `{"to": 123}`, `not json` and a 2 MiB body each give `400 invalid_request`.
  - `TestSend_RepeatDelivers`: the same SMS with `reference: "c1:u1:sms"` twice gives two `202` with different `id`s. `total` is 2 and `duplicate_deliveries` is 1. Checks FR-4.
  - `TestSend_EmptyReferenceNotDuplicate`: three SMS with no `reference` give `duplicate_deliveries == 0`.
  - `TestSend_DuplicatePerChannel`: the same `reference` on one SMS and one email gives `duplicate_deliveries == 0` on both channels.
  - `TestMessages_Filters`: store 5 messages across both channels, with 2 `to` values and 2 references. Check each filter, newest first (advance `Now` between sends), `limit=2&offset=1`, and `total` before paging. Checks FR-13.
  - `TestMessages_BadQuery`: `limit=0`, `limit=1001`, `limit=x`, `offset=-1`, `channel=push` each give `400`.
  - `TestStats_Counts`: after 2 accepted SMS, 1 rejected SMS and 1 accepted email, the stats are `sms: requests 3, delivered 2, rejected 1` and `email: requests 1, delivered 1`. All other counters are 0. Checks FR-14 for the normal path.
  - `TestReset`: after sends, `DELETE /v1/messages` gives `204`. Then `total == 0`, all counters are 0, and a repeat of an old `reference` is not a duplicate. Checks FR-15.
  - `TestConcurrent_DuplicateCount`: 50 goroutines send 10 SMS each with the same `reference`. Then `delivered == 500` and `duplicate_deliveries == 499`. Checks NFR-1.

  Run: `cd backend && go test ./internal/mockprovider/...`
  Expected: compile error, because `New` does not exist.

- [ ] **Step 2: Write the store in `store.go`.** One `sync.Mutex` guards the messages, the per-channel references (`map[string]struct{}`) and the `ChannelStats`. The store method that adds a message also updates `Delivered` and `DuplicateDeliveries` in the same lock. IDs are `sms_` or `email_` plus a counter in 8 digits, for example `sms_00000001`.

- [ ] **Step 3: Write `validate.go` and `server.go`.** Use `http.NewServeMux` with method patterns. Decode with `http.MaxBytesReader` (1 MiB) and `json.Decoder`. Do not call `DisallowUnknownFields`. For email `to`, accept the address only if `mail.ParseAddress(to)` returns `Name == ""` and `Address == to`.

  Run the tests again. Expected: PASS.

---

### Task 3: Rate limit and failure injection

**Files:**
- Create: `backend/internal/mockprovider/failure.go`, `failure_test.go`, `limit_test.go`
- Modify: `backend/internal/mockprovider/server.go`, `backend/go.mod`, `backend/go.sum`

**Interfaces:**
- Consumes: `Options`, `New`, `ChannelStats` (Task 2).
- Produces: no new exported names. Task 4 uses only `New`.

- [ ] **Step 1: Write the failing tests.** Use a fake clock that the test moves. Use a fake `Sleep` that records each duration and returns at once.
  - `TestRateLimit_Burst` (`limit_test.go`): `RatePerSec: 100, Burst: 10`. 200 SMS at one instant give 10 `202` and 190 `429`. Each `429` has `Retry-After` of 1 or more and `code: rate_limited`. `rate_limited == 190`. Checks FR-5, FR-6.
  - `TestRateLimit_Steady`: the same options. 100 SMS, with the clock moved 10 ms before each one, give 0 `429`. Checks FR-6.
  - `TestRateLimit_PerChannel`: use all SMS tokens. The next email gets `202`. Checks FR-5.
  - `TestRateLimit_429UsesNoToken`: `Burst: 1`. A forced `429`, then a normal SMS at the same instant, gives `202`.
  - `TestForced` (`failure_test.go`): table test for `X-Mock-Fail`:
    - `500` and `503` give that status, with codes `provider_error` and `unavailable`. `delivered` stays 0 and `errors` goes up. Checks FR-7.
    - `slow` gives `202`. The fake `Sleep` got `SlowDelay`. `delivered` and `slow` are 1. Checks FR-8.
    - `lost`: `client.Do` returns an error and no response. `delivered` and `lost` are 1. Checks FR-9.
    - `429` gives `Retry-After: 1`. `none` with `ErrorRate: 1` gives `202`. `boom` gives `400`. Checks FR-10.
    - `500` on an SMS with an empty `body` gives `400`.
  - `TestSlow_StoresAfterClientLeaves`: the fake `Sleep` for `slow` cancels the client request context, then returns. The client gets a context error. Poll `GET /v1/messages` for at most 1 s, and expect `total == 1`.
  - `TestRandom_ZeroRates`: all rates at 0. 1,000 SMS give 1,000 `202`. Checks FR-11.
  - `TestRandom_Distribution`: test the failure picker directly, not through HTTP. With `Random: 1` and the default rates, 10,000 picks give each mode within 1 percentage point of its rate. `500` and `503` share the error rate. Checks FR-12.
  - `TestLatency`: `LatencyMax: 50ms`. Each normal send calls `Sleep` with a value from 0 to 50 ms. `LatencyMax: 0` makes no `Sleep` call.

  Run: `cd backend && go test ./internal/mockprovider/...`
  Expected: the new tests fail.

- [ ] **Step 2: Write `failure.go`.** Define the modes `none`, `error500`, `error503`, `slow`, `lost` and `forced429`. Write a picker with one `rand.New(rand.NewPCG(o.Random, 0))` behind a mutex. One `Float64()` draw selects the mode by cumulative rate. A second draw selects `500` or `503` for the error mode. Write `parseForced(header string) (mode, ok bool, err error)`.

- [ ] **Step 3: Add the limiter and the modes to `server.go`.** Add `golang.org/x/time/rate`. Create one `rate.NewLimiter(rate.Limit(o.RatePerSec), o.Burst)` for each channel.
  - Use `lim.AllowN(o.Now(), 1)`. Do not use `Allow()`, because it reads the real clock.
  - For `Retry-After`, compute `(1 - lim.TokensAt(now)) / RatePerSec` seconds. Round it up, with a minimum of 1.
  - Follow "Order of checks" in `goal.md`. Do not pass `r.Context()` to the wait or to the store.
  - For `lost`, use `http.NewResponseController(w).Hijack()`, then close the connection.

  Run the tests again. Expected: PASS.

---

### Task 4: Binary, compose, docs and the end-to-end check

**Files:**
- Modify: `backend/cmd/mockprovider/main.go`, `docker-compose.yml`, `.env.example`, `docs/architecture.md`

**Interfaces:**
- Consumes: `config.LoadMockProvider` (Task 1), `mockprovider.New` and `Options` (Tasks 2, 3).

- [ ] **Step 1: Change `cmd/mockprovider/main.go`.**
  1. Call `config.LoadMockProvider(os.Getenv)`. On error, log it and exit 1.
  2. Map the config to `mockprovider.Options`, with `Now` and `Sleep` left nil.
  3. Serve `mockprovider.New(opts)`. Keep the shutdown logic as it is.
  4. Log the settings once at startup, with `Random`, so that a demo run can be repeated.
  5. Set `WriteTimeout` to more than `SlowDelay`, or leave it unset. Otherwise the server ends slow responses itself.

- [ ] **Step 2: Update compose and `.env.example`.**
  - `mock-provider` gets no `env_file`. Compose reads the `${...}` values from `.env`, and the provider must not get `ANTHROPIC_API_KEY`.
  - Add `environment` with each `MOCK_*` var as `${MOCK_X:-<default>}`. `MOCK_RANDOM` gets no default.
  - `.env.example` gets a "Mock provider" block with each var as a comment and its default, as for the seed vars.

- [ ] **Step 3: Update `docs/architecture.md`.** Add `mockprovider/` to `internal/` in the folder tree. Under "Send" in the flow, add one line: the provider has no dedup, and the worker owns it (D-018, D-019).

- [ ] **Step 4: Check NFR-6.**

  Run: `cd backend && go list -f '{{join .Imports "\n"}}' ./cmd/mockprovider ./internal/mockprovider | grep campaigncomposer`
  Expected: only `campaigncomposer/internal/config` and `campaigncomposer/internal/mockprovider`.

- [ ] **Step 5: Check the running service with a separate project.**

  Run:
  ```bash
  docker compose -p cc-verify up --build -d --wait mock-provider; echo "exit=$?"
  curl -s -o /dev/null -w '%{http_code}\n' -X POST localhost:8081/v1/sms -H 'X-Mock-Fail: none' \
    -d '{"to":"+905321234567","body":"Merhaba Ayşe","reference":"c1:u1:sms"}'
  curl -s -o /dev/null -w '%{http_code}\n' -X POST localhost:8081/v1/sms -H 'X-Mock-Fail: 503' \
    -d '{"to":"+905321234567","body":"x"}'
  curl -s localhost:8081/v1/stats
  docker compose -p cc-verify down -v
  ```
  Expected: `exit=0`, then `202`, then `503`. The stats show `sms.delivered: 1` and `sms.errors: 1`. If port 8081 is in use by the default project, run `make down` first.

- [ ] **Step 6: Run `make verify`.** Expected: PASS.

---

## Out of scope (owned by later slices)

- The send worker, the `deliveries` table, the retry rules and the `unknown` status: slice 07.
- The dashboard view of `GET /v1/stats`: slice 08.
- Delivery webhooks, storage that stays after a restart, auth: not planned.
