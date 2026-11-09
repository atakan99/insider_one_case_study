# Requirements: Mock SMS and email provider

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

- **FR-1.** `POST /v1/sms` and `POST /v1/email` with a valid body return `202` with a new `id`. The provider stores the message.
- **FR-2.** A request that is not valid returns `400` with `code: invalid_request`. The provider stores nothing. This includes an SMS `body` over 160 characters, a `subject` over 80 characters, a bad `to` and a missing required field.
- **FR-3.** The provider counts characters as runes. An SMS `body` of 160 Turkish characters, for example `ş`, gets `202`. A `body` of 161 characters gets `400`.
- **FR-4.** A repeat request with the same body and `reference` gets `202` and a new `id`. The provider stores both. `duplicate_deliveries` goes up by 1.
- **FR-5.** When a channel is over its rate limit, the provider returns `429` with a `Retry-After` header of 1 or more. A `429` on `sms` does not use an `email` token.
- **FR-6.** With `MOCK_RATE_PER_SEC=100` and the default burst, 200 requests sent at once to one channel get at least one `429`. 100 requests sent at a steady 100 per second get no `429`.
- **FR-7.** `X-Mock-Fail: 500` and `X-Mock-Fail: 503` return that status. The provider stores nothing.
- **FR-8.** `X-Mock-Fail: slow` returns `202` after `MOCK_SLOW_DELAY`. The provider stores the message.
- **FR-9.** `X-Mock-Fail: lost` closes the connection with no response. The provider stores the message. The client gets a transport error, not an HTTP status.
- **FR-10.** `X-Mock-Fail: 429` returns `429` with `Retry-After: 1` and uses no token. `X-Mock-Fail: none` gets no random failure. Any other value gets `400`.
- **FR-11.** With all three rates at `0`, 1,000 valid requests get `202`.
- **FR-12.** With a fixed `MOCK_RANDOM`, the count of each failure mode over 10,000 requests is within 1 percentage point of its rate.
- **FR-13.** `GET /v1/messages` filters by `channel`, `to` and `reference`. It sorts newest first and applies `limit` and `offset`. `total` counts all matches.
- **FR-14.** `GET /v1/stats` gives each counter from `goal.md` for each channel. The counters match the responses that the provider sent.
- **FR-15.** `DELETE /v1/messages` returns `204`. After it, `GET /v1/messages` gives `total: 0` and all counters are 0.
- **FR-16.** An env var that is not valid stops the process at startup with a clear error. This includes a rate outside 0 to 1, a sum of rates over 1, and a bad duration.

## Non-functional requirements

- **NFR-1.** The handlers are safe for concurrent requests. A test sends 500 requests from 50 goroutines. `delivered` is then equal to the count of `202` responses plus `lost` responses.
- **NFR-2.** Tests use `httptest` and an injected clock and sleep function. No test waits for `MOCK_SLOW_DELAY` or for the rate limiter in real time.
- **NFR-3.** Tests set the failure rates and do not depend on the defaults.
- **NFR-4.** `docker-compose.yml` passes each `MOCK_*` env var to `mock-provider`, with the defaults from `goal.md`. `.env.example` lists them as comments.
- **NFR-5.** `make up` starts `mock-provider`. `curl -X POST localhost:8081/v1/sms` with a valid body and `X-Mock-Fail: none` returns `202`.
- **NFR-6.** `cmd/mockprovider` imports nothing from `internal/` except `config` and `mockprovider`. `internal/mockprovider` imports nothing from `internal/`.
- **NFR-7.** `make verify` passes.
