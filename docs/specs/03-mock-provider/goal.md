# Goal: Mock SMS and email provider

## Why

The case sends no message to real people. The send worker sends each message to a mock provider. The case says that the mock provider must sometimes return errors, delays and `429`. Thus we can test the retry and rate-limit logic of the worker.

The mock provider acts like a real provider, for example Twilio Messaging or SendGrid Mail Send. A worker that works against the mock provider works against a real provider with few changes.

## Input

- The case brief, sections "Gönderim" and "Kaynaklar".
- The research of 2026-11-09 on Twilio Messaging and SendGrid Mail Send:
  - Neither API supports an `Idempotency-Key` header on the send endpoint.
  - A repeat request sends the message again.
  - SendGrid `custom_args` lets the client attach its own tracking ID. SendGrid stores it but does not use it to dedupe.
- User decisions of 2026-11-09:
  - The mock provider does no dedup and has no idempotency feature. The worker owns all dedup logic. See D-018.
  - After a timeout or a dropped connection, the worker does not retry. It marks the message `unknown`. See D-019.
  - The mock provider keeps a `reference` field and a stats endpoint as an audit. Only tests and the dashboard read them. The worker does not.

## Output

1. A Go package `internal/mockprovider` with the HTTP handlers, the rate limiter, the failure injection and the message store.
2. The binary `cmd/mockprovider` uses this package. It reads its settings from env vars.
3. The `mock-provider` service in `docker-compose.yml` gets the env vars, with the defaults from this spec.

This slice does not build the send worker. Slice 07 builds it. This slice does not add delivery webhooks, storage that stays after a restart, or auth.

## API

All bodies are JSON. All errors have this shape:

```json
{"error": {"code": "invalid_request", "message": "body is longer than 160 characters"}}
```

### `POST /v1/sms`

Request:

| Field | Type | Rule |
|---|---|---|
| `to` | string | Required. E.164: `^\+[1-9]\d{7,14}$`. |
| `body` | string | Required. 1 to 160 characters. |
| `reference` | string | Optional. Up to 128 characters. The provider stores it and does not use it. |

### `POST /v1/email`

Request:

| Field | Type | Rule |
|---|---|---|
| `to` | string | Required. One address that `net/mail.ParseAddress` accepts, with no display name. |
| `subject` | string | Required. 1 to 80 characters. |
| `body` | string | Required. Not empty. It can contain HTML. |
| `reference` | string | Optional. Same rule as for SMS. |

The provider counts characters as Unicode code points (Go runes), not bytes. Thus `ş` counts as 1 character. A real SMS provider counts GSM-7 and UCS-2 segments. The case gives a limit in characters, so we count characters.

The provider ignores unknown JSON fields. A request body over 1 MiB gets `400`.

### Send responses

| Status | When | Delivered | Body or header |
|---|---|---|---|
| `202` | The provider accepted the message. | Yes | `{"id": "<id>", "status": "accepted"}` |
| `400` | The JSON or a field is not valid. | No | `code: invalid_request` |
| `429` | The channel is over its rate limit. | No | `code: rate_limited`, header `Retry-After` |
| `500` | Random provider error. | No | `code: provider_error` |
| `503` | Random provider error. | No | `code: unavailable` |
| `202` after a delay | Random slow response. | Yes | Same as `202` |
| No response | Random lost response. The provider closes the connection. | Yes | None |

The `id` is an opaque string, unique for each delivery. A repeat request gets a new `id`. `received_at` is UTC in RFC 3339 format.

### Order of checks

For each send request, the provider does these steps in this sequence:

1. Read `X-Mock-Fail`. If the value is not in the table in "Forced failures", return `400`.
2. Rate limit. For a forced `429`, return `429` and use no token. If the channel has no token, return `429`. A `429` does not use a token.
3. Parse and validate the body. If it is not valid, return `400`.
4. Select the failure mode. See "Failure injection".
5. For `500` or `503`, return the error. Do not store the message.
6. Wait for the latency. For `slow`, wait for the slow delay instead.
7. Store the message.
8. For `lost`, close the connection with no response. Otherwise return `202`.

In steps 6 and 7, the provider ignores a client that disconnects. It stores the message also when the client timed out. This is the case that makes a retry not safe.

### `GET /v1/messages`

This endpoint lists the stored messages, newest first. Tests and the dashboard use it. The worker does not.

Query parameters, all optional: `channel` (`sms` or `email`), `to`, `reference`, `limit` (default 100, maximum 1000), `offset` (default 0).

Response:

```json
{"total": 2, "messages": [{"id": "...", "channel": "sms", "to": "+905321234567", "subject": "", "body": "...", "reference": "c1:u7:sms", "received_at": "2026-11-09T10:00:00Z"}]}
```

The filters are exact matches. `total` is the count of all matches, before `limit` and `offset`. A bad `channel`, `limit` or `offset` gets `400`.

### `GET /v1/stats`

Response, with one object for each channel:

```json
{"sms": {"requests": 0, "delivered": 0, "rejected": 0, "rate_limited": 0, "errors": 0, "slow": 0, "lost": 0, "duplicate_deliveries": 0}, "email": {}}
```

| Counter | Meaning |
|---|---|
| `requests` | All send requests. |
| `delivered` | Stored messages. This includes `slow` and `lost`. |
| `rejected` | `400` responses. |
| `rate_limited` | `429` responses. |
| `errors` | `500` and `503` responses. |
| `slow` | Slow responses. |
| `lost` | Lost responses. |
| `duplicate_deliveries` | Stored messages with a non-empty `reference` that the channel already had. |

The slice 07 tests and the demo check that `duplicate_deliveries` is 0.

### `DELETE /v1/messages`

This endpoint deletes all stored messages and all stored references, and sets all counters to 0. It returns `204`. Tests and the demo use it to start from a clean state.

### `GET /healthz`

This endpoint exists now. It does not change.

## Rate limit

- Each channel has its own token bucket. Use `golang.org/x/time/rate`. See D-020.
- `MOCK_RATE_PER_SEC` sets the rate. The default is `100`, the same as the case's limit for the worker.
- `MOCK_RATE_BURST` sets the bucket size. The default is `10`.
- `Retry-After` gives whole seconds, rounded up, with a minimum of `1`.

The provider limit is the same as the worker limit. Thus a correct worker sometimes gets a `429` from timing jitter, as with a real provider. A worker that sends in bursts gets many `429` responses.

## Failure injection

### Random failures

For each valid request, the provider selects one mode with these rates:

| Env var | Default | Mode |
|---|---|---|
| `MOCK_ERROR_RATE` | `0.03` | `500` or `503`, half each |
| `MOCK_SLOW_RATE` | `0.03` | `slow` |
| `MOCK_LOST_RATE` | `0.01` | `lost` |

The other requests are normal. Each rate is from 0 to 1. The sum must be 1 or less. If a rate is not valid, the process does not start.

- `MOCK_SLOW_DELAY` sets the delay for `slow`. The default is `5s`. Slice 07 sets the worker timeout below this value.
- `MOCK_LATENCY_MAX` sets the normal latency. Each normal request waits a random time from 0 to this value. The default is `50ms`. `0` turns the latency off.
- `MOCK_RANDOM` sets the random seed. If it is not set, the seed comes from the clock.

### Forced failures

The header `X-Mock-Fail` sets the mode for one request. It replaces the random choice. Tests use it to get each path exactly.

| Value | Result |
|---|---|
| `none` | Normal. No random failure. |
| `429` | `429` with `Retry-After: 1`. No token is used. |
| `500` | `500` |
| `503` | `503` |
| `slow` | `slow` |
| `lost` | `lost` |

- A forced `429` comes before validation, as in step 2. The other forced modes replace the random choice in step 4. Thus a forced `500` on a body that is not valid gets `400`.
- Any other value gets `400`, as in step 1.

## What the worker must do (slice 07)

This section is not part of this slice. It shows why the provider acts as it does.

- The worker keeps a `deliveries` row with a unique key: campaign, user and channel.
- The status goes from `pending` to `sending`, then to `sent`, `failed` or `unknown`.
- `429`: wait for `Retry-After`, then retry.
- `500`, `503` or connection refused: retry with backoff. The provider did not deliver the message.
- `400`: mark `failed`. Do not retry.
- Timeout or dropped connection after the request was sent: mark `unknown`. Do not retry. The provider can have the message, and a retry can send a second copy.
- The worker sets `reference` to `<campaign id>:<user id>:<channel>`.
- The worker must not set an `Idempotency-Key` header. Go's `net/http` client treats a request with this header as idempotent and retries it by itself after a dropped connection.

## Message store

- The store is in memory. A restart deletes all messages and counters.
- The store has no size limit. A demo sends fewer than 100,000 messages.
- All handlers are safe for concurrent requests.

## Settings and clock

- `internal/config` gets a `MockProviderConfig` and a `LoadMockProvider(getenv)` function, the same pattern as `SeedConfig`. `HTTP_ADDR` keeps its default `:8081`.
- Durations use the Go format, for example `5s` or `50ms`.
- The server takes a clock and a sleep function. The rate limiter, `received_at` and all waits use them. Thus tests run with no real waits.
- The `lost` mode uses `http.Hijacker`. It works on HTTP/1.1, which the worker and `httptest` use.
