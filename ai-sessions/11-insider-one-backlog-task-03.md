# Insider one backlog task 03

- Session: `b8a1807d-a6d6-4467-96a4-6fbd029e615a`
- Model: `claude-opus-5-5`
- Time (UTC): 2026-11-09 14:49:27 → 2026-11-09 15:27:53
- Subagent runs: 1

---

## 👤 User  <sub>2026-11-09 14:49:27</sub>

_[IDE: opened file]_ 
go on with task 03

**🔧 Bash** _Read progress log, backlog, and docs layout_
```bash
cat tasks/progress.md && cat tasks/backlog.json && ls docs docs/specs docs/plans && git diff .gitignore
```

<details><summary>Result</summary>

```
# Progress

Append one entry per session. Use the template below. Put the newest entry at the bottom.

<!-- Template:
## YYYY-MM-DD — <slice id>
- Done:
- Verified with:
- Next:
- Open questions:
-->

## 2026-11-06 — 00-skeleton
- Done: Task 1 of `docs/plans/00-skeleton.md`. Moved the case brief to `docs/case.md` and the frontend-trim spec to `docs/specs/01-frontend-trim/`. Wrote `.gitignore`, `CLAUDE.md`, `DECISIONS.md` (D-001 to D-006), `README.md`, `docs/architecture.md`, `tasks/backlog.json`, `tasks/progress.md` and `ai-sessions/.gitkeep`.
- Verified with: `git check-ignore` (prints `.env`, `.env.local`, then `exit=1`). `python3 -m json.tool tasks/backlog.json` (prints `ok`).
- Next: Task 2 (ste-writing skill, executor and reviewer subagents).
- Open questions: None.

## 2026-11-06 — 00-skeleton
- Done: Task 2 of `docs/plans/00-skeleton.md`. Added the `ste-writing` skill (`SKILL.md`, `references/rules.md`, `references/terms.md`, `scripts/ste_check.py` and its tests). Added the `executor` and `reviewer` subagents. Changed the Writing style line in `CLAUDE.md`. Deleted the template git hooks in `.git/hooks/` (user decision). The reviewer found 6 defects in `ste_check.py`. All 6 are fixed, and each one has a regression test.
- Verified with: `python3 -B -m unittest discover -s .claude/skills/ste-writing/scripts` (11 tests, OK). The session lists `executor`, `reviewer` and `ste-writing`.
- Next: Task 3 (Go module, config, health endpoints, two binaries).
- Open questions: `.gitignore` does not include `__pycache__/`. The tests write it when they run without `-B`.

## 2026-11-06 — 00-skeleton
- Done: Task 3 of `docs/plans/00-skeleton.md`. Added the Go module `campaigncomposer` (go 1.27, `pgx/v5` v5.12.0), `internal/config`, `internal/httpapi` (health endpoints), `cmd/api` and `cmd/mockprovider`. The reviewer found 3 defects in the two `main.go` files. All 3 are fixed: a 2 s timeout on the startup DB ping, no exit 0 when a listen error races a signal, and a second Ctr
… [12015 more chars truncated]
```

</details>

**🔧 Skill** ```json
{
  "skill": "superpowers:brainstorming"
}
```

<details><summary>Result</summary>

```
Launching skill: superpowers:brainstorming
```

</details>

> _[skill instructions loaded: brainstorming]_

**🔧 Bash** _Read case brief, architecture, decisions, mock provider code_
```bash
cat docs/case.md docs/architecture.md && tail -60 docs/decisions-log.md && cat DECISIONS.md | head -80 && ls docs/specs/02-seed-users && cat backend/cmd/mockprovider/main.go && ls -R backend/internal | head -50
```

<details><summary>Result</summary>

```
Content is user-generated and unverified.

Smart Campaign Composer Case

## İş bağlamı

Insider One müşterileri her gün milyonlarca kişiye SMS ve email gönderiyor. Bir kampanya hazırlamak için kitleyi tanımlamak, her kanal için metin yazmak ve karakter limitlerine uymak gerekiyor. Bu case'te pazarlamacının yalnızca ne istediğini yazdığı, gerisini Claude'un önerdiği ve son kararı insanın verdiği bir araç kurmanı istiyoruz.

## Kullanıcı akışı

Örnek brief

“Son 30 günde sepetine ürün ekleyip alışveriş yapmayan İstanbul'daki kullanıcılara %30 indirim hatırlatması gönder. Samimi ama kısa olsun.”

1. **Kitle çıkarılır** Claude brief'ten bir segment kuralı üretir, arayüz kaç kişiye gideceğini gösterir.

2. **Mesajlar üretilir** SMS ve email içeriği ekrana stream edilerek akar.

3. **İnsan onaylar** Pazarlamacı önizler, gerekirse düzenler ve onaylar.

4. **Gönderilir** Mesajlar kuyruk üzerinden mock provider'a gönderilir.

5. **Canlı izlenir** Dashboard gönderim durumunu gösterir.


Gerçek kimseye mesaj gitmez. Alıcılar senin oluşturacağın sahte kullanıcılardır, gönderim de senin yazacağın mock provider'a yapılır.

## Fonksiyonel gereksinimler

### Kitle

- Claude'un ürettiği segment kuralı backend'de doğrulanır; LLM çıktısı doğrudan SQL olarak çalıştırılmaz.
- İzni olmayan kullanıcıya o kanaldan mesaj gönderilmez.

### İçerik

- İçerik kanal limitlerine uymalı: SMS en fazla 160 karakter, email konusu en fazla 80 karakter.
- `{{first_name}}` gibi kişiselleştirme değişkenleri desteklenir.

### Gönderim

- Asenkron kuyruk, kanal başına saniyede en fazla 100 mesaj.
- Aynı kişiye aynı mesaj iki kez gitmez. Retry stratejisini sen tasarla.

### Takip

- Canlı gönderim durumu, metrics ve health check endpoint'leri.

## AI gereksinimleri

Case'in en önemli kısmı. Claude'u çağırmak yeterli değil; çıktısını güvenilir hale getirmeni bekliyoruz.

- **Doğrulama:** Claude'un çıktısı yapılandırılmış (JSON) olmalı ve backend'de kontrol edilmeli.
- **Test edilebilirlik:** testler gerçek A
… [12843 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-09 14:49:50</sub>

I've read the progress log, the case brief and the current `cmd/mockprovider` (it only has `/healthz` now). Slice 03 has no spec yet, so I'll write one first, as we did for slice 02.

If you'd rather skip my questions, say so. I'll write the spec with sensible defaults and you can correct it.

My first question: when you picture the mock provider during the demo or the review call, what do you want it to show about the send worker? For example, retries recovering from 500s, the worker slowing down after 429s, or the same message never being delivered twice.

## 👤 User  <sub>2026-11-09 15:01:37</sub>

it will act similar to the real life counterpart, generally it should work but it can sometimes error because of the provider if we send too many reqs it can 429 as well right?

## 🤖 Claude  <sub>2026-11-09 15:01:50</sub>

Here is what I understand so far:

**What it is.** A small HTTP service that acts like a real SMS and email provider, for example Twilio or SendGrid. The send worker (slice 07) calls it the same way it would call a real provider.

**Normal behaviour.** Most requests succeed. The provider accepts the message and returns a provider message ID.

**Failures from the provider.** Some requests fail for reasons that have nothing to do with us: a 5xx error, or a slow response that can hit the worker's timeout. *My guess:* a small random rate for each failure type, with defaults such as 5% 5xx and 5% slow, set by env vars.

**429 when we send too fast.** You said "if we send too many reqs". So I read this as a real rate limit that counts requests, not a random 429. *My guess:* each channel has its own limit. The default is a little above 100 messages per second, the case's limit for the worker, so a correct worker never sees 429 and a worker that bursts does. The response includes a `Retry-After` header, as real providers do.

**Tests.** *My guess:* a seed or a "failures off" setting makes the behaviour repeatable, so tests in slice 07 can force a 429 or a 500 when they need one.

Tell me what's wrong or missing in that.

Then my next question: when a message reaches the provider, do you want the provider to keep a record of it, so that the dashboard or a test can prove that no user got the same message twice?

## 👤 User  <sub>2026-11-09 15:02:29</sub>

what is your opinions on this design how should it act

## 🤖 Claude  <sub>2026-11-09 15:02:56</sub>

My recommendation is to model it on a real provider, such as Twilio for SMS and SendGrid for email, plus a few test hooks. A real provider is the best test for our worker's retry and rate-limit code.

**Endpoints**
- `POST /v1/sms` with `{to, body}` and `POST /v1/email` with `{to, subject, body}`.
- An `Idempotency-Key` header on each request. The worker sets it to campaign ID + user ID + channel.
- Success returns `202 {id, status: "accepted"}`.

**How it fails**

| Case | Response | What the worker must do |
|---|---|---|
| Too fast for the channel | `429` + `Retry-After` | Wait, then retry |
| Provider problem (random, about 3%) | `500` / `503` | Retry with backoff |
| Slow (random, about 3%) | 2–5 s delay | Time out, then retry |
| **Accepted, but the response is lost** (about 1%) | Saves the message, then returns `500` | Retry **without** a duplicate |
| Bad input | `400` (SMS > 160, subject > 80, bad phone or email) | Do not retry, mark as failed |

The "accepted, but the response is lost" row is the most important one. It is the real reason duplicates happen. Our worker cannot know that the first attempt worked. So the provider must honor `Idempotency-Key`. A repeat key returns the first response and does not deliver again.

**Rate limit:** a token bucket for each channel, default 100/s with a small burst. These come from env vars. The limit is the same as our worker's limit, so the worker sometimes gets a 429 from timing jitter. This shows the 429 path in the demo, as with a real provider.

**Proof and tests**
- The provider keeps all delivered messages in memory. `GET /v1/messages` lists them. `GET /v1/stats` shows accepted, rejected and 429 counts, duplicate keys absorbed, and **delivered more than once**. That last count must be 0.
- An `X-Mock-Fail: 429|500|slow|lost` header forces a failure. Tests in slice 07 can then hit each path exactly, with no random results. The failure rates are env vars, and `0` turns the random failures off.

**Not in this slice:** delivery-status webhooks back to the api, storage that survives a restart, and auth.

One question before I write the spec. Should the provider absorb a repeat `Idempotency-Key`, as above? The other option is to deliver it again and only count it. Then the worker must prevent duplicates by itself, which is harder but shows more of our own work.

## 👤 User  <sub>2026-11-09 15:04:48</sub>

a quick search i believe we should do the dedup idempotency logic since it does not supported everytime


Twilio supports idempotency keys for specific endpoints (such as Twilio Cohort Snapshots using the Idempotency-Key header), though core messaging endpoints like standard Messages.json creation do not natively accept an idempotency key header. [1] (https://www.twilio.com/docs/cohorts/concepts/snapshots), [2] (https://www.swytchcode.com/content/twilio-api-agent-sms)

How Idempotency Works in Supported Endpoints

• Header Name: Idempotency-Key (accepted on select APIs like Cohort Snapshots).
• Key Format: A unique string (such as a UUID) up to 128 characters long per distinct request.
• Behavior: Re-sending a request with the same key and identical payload returns the original response (including an Idempotent-Replayed: true header) without repeating the action.
• Payload Mismatch: Reusing the same key with a different request body fails with an HTTP 422 error (Error 17534).
• Retention: Keys are typically retained for at least 24 hours and scoped to your account and region. [1] (https://www.twilio.com/docs/api/errors/17534), [2] (https://www.twilio.com/docs/cohorts/concepts/snapshots)

Webhooks (Incoming from Twilio)

• Header Name: I-Twilio-Idempotency-Token is sent by Twilio in webhook requests to help distinguish retry attempts on your server. [1] (https://www.twilio.com/docs/usage/webhooks/webhooks-connection-overrides)

Handling Idempotency Where Native Support is Absent

• For core APIs like sending SMS messages where Twilio does not parse a native Idempotency-Key parameter on creation, you must build deduplication logic on your own application side (e.g., tracking message tracking IDs/SIDs, hashing payloads, or checking a local database before triggering a POST request). [1] (https://stackoverflow.com/questions/39090633/what-should-i-use-as-the-document-key-to-maintain-idempotency), [2] (https://www.swytchcode.com/content/twilio-api-agent-sms)
If you are trying to prevent duplicate requests, tell me:
• Which specific Twilio API endpoint or resource you are calling (e.g., Messaging, Voice, Cohorts, or TaskRouter)
• What language or helper library you are using
I can help you design a safe retry or deduplication strategy.

## 🤖 Claude  <sub>2026-11-09 15:05:07</sub>

Good. Then the provider acts like Twilio Messaging: it does **not** dedupe, and our worker owns the dedup logic.

**Repeat requests.** The provider delivers every request it accepts, also a repeat. It does not read `Idempotency-Key`.

**How we prove no duplicates.** Each request can carry an optional `reference` field, the same as `custom_args` in SendGrid or a client tag. The worker sets it to `campaign:user:channel`. The provider stores it and does nothing else with it. `GET /v1/stats` counts `duplicate_deliveries`, which is the number of deliveries whose `reference` the provider already has. The demo and the slice 07 tests check that this count is 0.

**The hard case.** Sometimes the provider saves the message and then returns `500`, or responds too slowly and the worker times out. Without provider dedup, a blind retry can then deliver a second copy. My recommendation is that the provider gives a lookup, `GET /v1/messages?reference=...`, as Twilio lets you list messages. Before the worker retries after an ambiguous failure, it checks whether the provider already has the message. The worker's side of this goes in slice 07. This slice only supplies the lookup. *This is my call. Tell me if you want a simpler worker rule instead, such as "never retry after a timeout".*

**Everything else stays as in my last message:** the per-channel token bucket with `429` + `Retry-After`, random `5xx` and slow responses from env vars, `400` for bad input, the `X-Mock-Fail` header for exact tests, and messages kept in memory.

**Size:** this is a project. I'll write the spec as `docs/specs/03-mock-provider/goal.md` and `requirements.md`, in the same format as slice 02. Then I'll read it as the builder would before I give it to you.

Is anything wrong or missing, or shall I write the spec?

## 👤 User  <sub>2026-11-09 15:06:39</sub>

here i also ask how sendgrid did it so we should handle it in our own logic so the mock services should not have them, i want your opinions as well


Twilio SendGrid does not support native native HTTP idempotency keys (like an Idempotency-Key header) for its core Mail Send API or marketing campaigns. If a network timeout occurs and your application retries a POST /v3/mail/send request, SendGrid will treat it as a brand-new instruction and likely send the email a second time.

How to Prevent Duplicates in SendGrid

Because native deduplication is absent, you must enforce safety constraints within your application's architecture:

1. Build a Distributed Lock or Unique Constraint

Track the outcome of outbound notifications in your own database before firing requests over to the SendGrid Mail Send API:
• State tracking: Save an explicit notification log with a status of PENDING alongside a unique business key (e.g., invoice_1042_receipt).
• Database Guard: Use a unique database constraint or a distributed cache tool like Redis to guarantee that two concurrent threads cannot process the exact same business trigger simultaneously.
• State updates: Only change the state to SENT once SendGrid answers with an HTTP 202 Accepted code. [1] (https://medium.com/@wahyubagus1910/idempotency-keys-how-to-prevent-duplicate-request-and-api-chaos-3ad6b1cdfe30)

2. Differentiate Application Retries from Network Failures

• For 5xx or Connection Drops: If the connection drops before a response header is parsed, you cannot be certain if SendGrid received the payload. To safely retry without risk, you may choose to alert an administrator or cross-reference SendGrid's Event Webhook.
• For 4xx Responses: If SendGrid returns a validation issue (like a 400 Bad Request due to duplicate recipient arrays), do not retry the identical body payload. Fix the recipient arrays programmatically first. [1] (https://stackoverflow.com/questions/53602485/prevent-duplicate-email-addresses-in-sendgrid-message)

3. Leverage SendGrid's Unique Arguments (custom_args)

While custom_args will not stop SendGrid from sending a duplicate email if you push the button twice, they provide a reliable mechanism to audit and track mistakes after the fact:
• Attach your internal transaction or tracking UUIDs to the email metadata using the custom_args field.
• Listen to SendGrid Event Webhooks. If your receiver detects two distinct delivered events carrying the exact same custom tracking UUID, you can flags the bug internally to audit your application's retry intervals.

## 🤖 Claude  <sub>2026-11-09 15:06:56</sub>

I agree. The mock provider does no dedup and has no idempotency features, the same as SendGrid and Twilio Messaging. The worker owns all of it. I also withdraw the `GET /v1/messages?reference=...` lookup that I proposed for the worker. SendGrid has no reliable lookup of that kind, so the worker must not depend on one.

What I recommend:

**Mock provider (this slice)**
- Delivers every request it accepts, also repeats.
- Accepts an optional `custom_args`-style `reference` field. It only stores this field and never acts on it, as SendGrid does.
- `5xx` / `503` means "not delivered". `slow` and `lost` mean that the provider delivered the message, but the worker gets a timeout or a dropped connection. *My call:* only these two modes are ambiguous. This keeps the mock's rules clear and testable.
- Keeps `GET /v1/messages` and `GET /v1/stats`, with `duplicate_deliveries` counted by `reference`. This plays the part of the SendGrid Event Webhook audit. The worker never calls these endpoints. Only the tests and the dashboard use them, to prove that duplicates are 0.

**Worker (slice 07). I add this here so that you can see the full design.**
- A `deliveries` row with a unique key `(campaign, user, channel)` and the status `pending → sending → sent | failed | unknown`, as your SendGrid notes describe.
- Retry with backoff only when we know that the provider did not deliver: `429` (obey `Retry-After`), `5xx`, connection refused.
- `400` → `failed`, no retry.
- Timeout or dropped connection after the request was sent → `unknown`, **no retry**. The dashboard shows these messages. The case says that the same message must never go to the same person twice, so we choose a possible missed message over a possible duplicate.

Do you agree with the "unknown, no retry" rule and with keeping `reference` + `/v1/stats` as the audit, so I can write the spec?

## 👤 User  <sub>2026-11-09 15:10:04</sub>

yes

**🔧 Skill** ```json
{
  "skill": "ste-writing"
}
```

<details><summary>Result</summary>

```
Launching skill: ste-writing
```

</details>

> _[skill instructions loaded: ste-writing]_

**🔧 Bash** _Read slice 02 spec format and compose/config_
```bash
cat docs/specs/02-seed-users/goal.md docs/specs/02-seed-users/requirements.md && cat docker-compose.yml .env.example && cat backend/internal/config/config.go | head -60 && cat backend/go.mod
```

<details><summary>Result</summary>

```
# Goal: Seed users

## Why

The case needs a few thousand fake users. Later slices use them:

- Segment rules filter them by city, language, age and shopping history.
- Templates use their names, for example `{{first_name}}`.
- The send flow checks their consent for each channel.

The data must be realistic enough for a demo. It must also contain the edge cases that the hard rules need: no consent, long names, Turkish characters.

## Input

- The case brief, section "Sahte kullanıcı verisi".
- The example brief: "users in Istanbul who added a product to the cart in the last 30 days and did not buy".
- User decisions of 2026-11-06:
  - Users are in two countries only: Turkey (`TR`) and the US (`US`).
  - Each user has a `language` and a `country` column. In this slice the language follows the country: `TR` users are `tr`, `US` users are `en`.
  - A campaign has one message language. A later slice can add campaigns with several languages and a new seed with mixed languages. See D-013.
  - Amounts are in Turkish lira (TRY) for all users.

## Output

1. A database schema with a `users` table and an `events` table. Migrations create it.
2. A Go binary `cmd/seed`. It applies the migrations, then inserts the users and their events.
3. A one-shot `seed` service in `docker-compose.yml`. The `api` service starts after `seed` completes.

This slice does not build the segment rule compiler or the audience preview. Slice 04 builds them. This slice only gives them the columns and the indexes.

## Data model

### `users`

| Column | Type | Notes |
|---|---|---|
| `id` | `bigint` identity | Primary key. Part of the idempotency key later. |
| `first_name` | `text` not null | Used by `{{first_name}}`. |
| `last_name` | `text` not null | Used by `{{last_name}}`. |
| `email` | `text` not null unique | Domain `example.com`. RFC 2606 reserves it, so no real person gets the address. |
| `phone` | `text` not null | E.164 format. `+90 5XX` for `TR`. `+1 AAA 555 01XX` for `US`. These US numb
… [11044 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:10:38</sub>

Now I'll write the spec.

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/docs/specs/03-mock-provider/goal.md`
````
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
| `202` | The provider accepted the message. | Yes | `{"id": "
… [5356 more chars truncated]
````

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/docs/specs/03-mock-provider/requirements.md`
```
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
- **FR-10.** `X-Mock-Fail: 429` returns `429` with `Retry-After: 1`. `X-Mock-Fail: none` gets no random failure. Any other value gets `400`.
- **FR-11.** With all three rates at `0`, 1,000 valid requests get `202`.
- **FR-12.** With a fixed `MOCK_RANDOM`, the count of each failure mode over 10,000 requests is within 1 percentage point of its rate.
- **FR-13.** `GET /v1/messages` filters by `channel`, `to` and `reference`. It sorts newest first and applies `limit` and `offset`. `total` counts all matches.
- **FR-14.** `GET /v1/stats` gives each counter from `goal.md` for each channel. The counters match the responses that the provider sent.
- **FR-15.** `DELETE /v1/messages` returns `204`. After it, `GET /v1/messages` gives `total: 0` and all counters are 0.
- **FR-16.** An env var that is not valid stops the process at startup with a clear error. This includes a rate outside 0 to 1, a sum of rates over 1, and a bad duration.

## Non-functional requirements

- **NFR-1.** The handlers are safe for concurrent requests. A test sends 500 requests from 50 goroutines. `delivered` is then equal to the count of `202` responses plus `lost` responses.
- **NFR-2.** Tests use `httptest` and an injected clock or sleep function. No test waits for `MOCK_SLOW_DELAY` in real time.
- **NFR-3.** Tests set the failure rates and do not depend on the defaults.
- **NFR-4.** `docker-compose.yml` passes each `MOCK_*` env var to `mock-provider`, with
… [403 more chars truncated]
```

**🔧 Bash** _Run STE checker on new spec files_
```bash
python3 .claude/skills/ste-writing/scripts/ste_check.py docs/specs/03-mock-provider/*.md
```

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/docs/specs/03-mock-provider/goal.md (file state is current in your context — no need to Read it back)
```

</details>

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/docs/specs/03-mock-provider/requirements.md (file state is current in your context — no need to Read it back)
```

</details>

<details><summary>Result</summary>

```
(Bash completed with no output)
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:11:38</sub>

The STE check found nothing. Now I'll read the spec as the builder would. I found some gaps that I can close myself:

**🔧 Bash** _Apply builder-check fixes to spec and rerun STE check_
```bash
python3 - <<'EOF'
p='docs/specs/03-mock-provider/goal.md'
s=open(p).read()
old_order=s[s.index("For each send request, the provider does these steps"):s.index("### `GET /v1/messages`")]
new_order='''For each send request, the provider does these steps in this sequence:

1. Read `X-Mock-Fail`. If the value is not in the table in "Forced failures", return `400`.
2. Rate limit. For a forced `429`, return `429` and use no token. If the channel has no token, return `429`. A `429` does not use a token.
3. Parse and validate the body. If it is not valid, return `400`.
4. Select the failure mode. See "Failure injection".
5. For `500` or `503`, return the error. Do not store the message.
6. Wait for the latency. For `slow`, wait for the slow delay instead.
7. Store the message.
8. For `lost`, close the connection with no response. Otherwise return `202`.

In steps 6 and 7, the provider ignores a client that disconnects. It stores the message also when the client timed out. This is the case that makes a retry not safe.

'''
s=s.replace(old_order,new_order)
s=s.replace("The `id` is unique for each delivery. A repeat request gets a new `id`.",
"The `id` is an opaque string, unique for each delivery. A repeat request gets a new `id`. `received_at` is UTC in RFC 3339 format.")
s=s.replace("`total` is the count of all matches, before `limit` and `offset`.",
"The filters are exact matches. `total` is the count of all matches, before `limit` and `offset`. A bad `channel`, `limit` or `offset` gets `400`.")
s=s.replace("This endpoint deletes all stored messages and sets all counters to 0. It returns `204`.",
"This endpoint deletes all stored messages and all stored references, and sets all counters to 0. It returns `204`.")
s=s.replace("""- A forced `429` comes before validation, as in step 1. The other forced modes come after validation, as in step 3.
- Any other value gets `400`.""","""- A forced `429` comes before validation, as in step 2. The other forced modes replace the random choice in step 4. Thus a forced `500` on a body that is not valid gets `400`.
- Any other value gets `400`, as in step 1.""")
s=s.replace("""- The worker sets `reference` to `<campaign id>:<user id>:<channel>`.""","""- The worker sets `reference` to `<campaign id>:<user id>:<channel>`.
- The worker must not set an `Idempotency-Key` header. Go's `net/http` client treats a request with this header as idempotent and retries it by itself after a dropped connection.""")
s=s.replace("""- All handlers are safe for concurrent requests.""","""- All handlers are safe for concurrent requests.

## Settings and clock

- `internal/config` gets a `MockProviderConfig` and a `LoadMockProvider(getenv)` function, the same pattern as `SeedConfig`. `HTTP_ADDR` keeps its default `:8081`.
- Durations use the Go format, for example `5s` or `50ms`.
- The server takes a clock and a sleep function. The rate limiter, `received_at` and all waits use them. Thus tests run with no real waits.
- The `lost` mode uses `
… [836 more chars truncated]
```

<details><summary>Result</summary>

```
16:- **FR-10.** `X-Mock-Fail: 429` returns `429` with `Retry-After: 1` and uses no token. `X-Mock-Fail: none` gets no random failure. Any other value gets `400`.
27:- **NFR-2.** Tests use `httptest` and an injected clock and sleep function. No test waits for `MOCK_SLOW_DELAY` or for the rate limiter in real time.
[This command modified 2 files you've previously read: docs/specs/03-mock-provider/goal.md, docs/specs/03-mock-provider/requirements.md. Call Read before editing.]
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:11:54</sub>

Now the decisions log entries and the backlog item:

**🔧 Bash** _Add decisions D-018 to D-021, fill backlog acceptance, run STE check_
```bash
cat >> docs/decisions-log.md <<'EOF'

## D-018 The mock provider does no dedup

- **Context:** The case says that the same message must not go to the same person twice. Twilio Messaging and SendGrid Mail Send have no `Idempotency-Key` on their send endpoints. A repeat request sends the message again.
- **Options:** The provider honors an `Idempotency-Key` and absorbs repeats. The provider delivers each repeat and the worker prevents duplicates.
- **Choice:** The provider delivers each repeat (user decision). It stores an optional `reference`, as SendGrid `custom_args` does. `GET /v1/stats` counts `duplicate_deliveries` for the audit.
- **Why:** The mock acts like the real providers. Thus the dedup logic is in our worker, where the case wants to see it. Provider dedup would hide a worker bug.

## D-019 Only a timeout or a dropped connection is ambiguous, and the worker does not retry it

- **Context:** With no provider dedup, a retry after an ambiguous failure can send a second copy.
- **Options:** Retry all failures. Check the provider before a retry. Do not retry ambiguous failures.
- **Choice:** In the mock, `500` and `503` mean "not delivered". `slow` and `lost` deliver the message, but the client gets a timeout or a dropped connection. The worker in slice 07 retries `429`, `5xx` and connection refused. It marks a timeout or a dropped connection `unknown` and does not retry (user decision).
- **Why:** The case forbids duplicates, so a possible missed message is better than a possible duplicate. SendGrid has no reliable lookup that the worker can use before a retry. Clear rules in the mock make each path testable.

## D-020 Token bucket from `golang.org/x/time/rate`

- **Context:** The mock provider needs a rate limit for each channel. The worker in slice 07 needs one too.
- **Options:** Our own token bucket, `golang.org/x/time/rate`, a Redis-based limiter.
- **Choice:** `golang.org/x/time/rate`, one limiter for each channel. The default is 100 per second with a burst of 10.
- **Why:** The package is maintained by the Go team and takes an explicit time, so tests can use a fake clock. The mock is one process, so it needs no shared state in Redis.

## D-021 `X-Mock-Fail` header for exact tests

- **Context:** The random failures make the worker tests flaky. The tests must reach each path exactly.
- **Options:** A fixed random seed only, a control endpoint that sets the next failure, a header on each request.
- **Choice:** The `X-Mock-Fail` header sets the mode for one request. The random rates and `MOCK_RANDOM` stay for the demo.
- **Why:** A header affects one request only. Thus parallel tests do not interfere. A control endpoint keeps state that a parallel test can change.
EOF
python3 - <<'EOF'
import json
p='tasks/backlog.json'
b=json.load(open(p))
import re
req=open('docs/specs/03-mock-provider/requirements.md').read()
items=[re.sub(r'\*\*(.+?)\*\*',r'\1',l[2:]).replace('`','') for l in req.splitlines() if l.startswith('- **')]
for it in b:
 
… [388 more chars truncated]
```

<details><summary>Result</summary>

```
ok
 .gitignore            |  3 +++
 docs/decisions-log.md | 28 ++++++++++++++++++++++++++++
 tasks/backlog.json    | 28 ++++++++++++++++++++++++++--
 3 files changed, 57 insertions(+), 2 deletions(-)
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:12:16</sub>

I wrote the spec for slice 03. Please review it before I write the plan. I have not written any code yet.

- [goal.md](docs/specs/03-mock-provider/goal.md): the API, the order of checks for each request, the rate limit, failure injection, the audit endpoints and what the slice 07 worker must do.
- [requirements.md](docs/specs/03-mock-provider/requirements.md): FR-1 to FR-16 and NFR-1 to NFR-7. I also copied these checks into the `03-mock-provider` item in [backlog.json](tasks/backlog.json).
- [decisions-log.md](docs/decisions-log.md): D-018 (no provider dedup), D-019 (timeout or dropped connection means `unknown`, no retry), D-020 (`x/time/rate`), D-021 (`X-Mock-Fail` header).

The STE check finds nothing in these files. I did the builder check myself and did not use a subagent. It found only small gaps, and I closed them.

These are my own calls. Please check them:

1. **Counting characters.** The provider counts runes, so `ş` is 1 character. A real SMS provider counts GSM-7 and UCS-2 segments, but the case gives its limit in characters.
2. **Defaults.**
   - The rate limit is 100/s for each channel with a burst of 10.
   - The random failures are 3% `5xx`, 3% slow and 1% lost.
   - A slow response takes 5 s, and normal latency is 0 to 50 ms.
3. **Order of checks.** A `429` comes before validation and uses no token, as with real providers. A forced `500` on a bad body still gets `400`.
4. **Client disconnects.** When a slow client times out, the provider still stores the message. This keeps the timeout case truly ambiguous.
5. **Audit endpoints.**
   - `GET /v1/messages` lists the stored messages.
   - `GET /v1/stats` gives the counters, including `duplicate_deliveries`.
   - `DELETE /v1/messages` resets them, which tests and the demo need.
6. **Gotcha for slice 07.** The worker must not send an `Idempotency-Key` header. Go's HTTP client sees that header and retries the request by itself after a dropped connection, and that retry could send a duplicate. I put this in the worker section of the spec.
7. **Config.** The settings go in `internal/config` as `MockProviderConfig`, the same way as `SeedConfig`.

`.gitignore` already had an uncommitted change (`docker-compose.override.yml`) before this session. I did not touch it.

If the spec is OK, shall I write the plan for `docs/plans/03-mock-provider.md`?

## 👤 User  <sub>2026-11-09 15:13:22</sub>

go

**🔧 Skill** ```json
{
  "skill": "superpowers:writing-plans"
}
```

<details><summary>Result</summary>

```
Launching skill: superpowers:writing-plans
```

</details>

> _[skill instructions loaded: writing-plans]_

## 🤖 Claude  <sub>2026-11-09 15:13:39</sub>

I'm using the writing-plans skill to write the plan. First, the slice 02 plan format and the existing patterns:

**🔧 Bash** _Read plan 02 format, config, Makefile, seed main_
```bash
head -80 docs/plans/02-seed-users.md && grep -n "^### Task\|^## " docs/plans/02-seed-users.md && wc -l docs/plans/02-seed-users.md docs/specs/02-seed-users/*.md && sed -n 60,200p backend/internal/config/config.go && cat Makefile | head -70 && cat backend/cmd/seed/main.go | head -50
```

<details><summary>Result</summary>

````
# 02 — Seed Users Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `docker compose up` creates the schema and loads 5,000 realistic users with their shopping events. The data is the same on each run with the same random seed.

**Architecture:** This slice adds two packages and one binary:

```
cmd/seed            applies migrations, then calls seed.Load
internal/store      embedded goose migrations, Migrate()
internal/store/storetest   test helper: one new database per test
internal/seed       Generate() (pure, no database) and Load() (pgx CopyFrom)
```

`seed` runs as a one-shot compose service. `api` starts after it completes (D-015).

**Tech Stack:** Go 1.27, `github.com/brianvoe/gofakeit/v7` v7.18.1, `github.com/pressly/goose/v3` v3.29.0, `github.com/jackc/pgx/v5` (already in `go.mod`), Postgres 17.

**Spec:** `docs/specs/02-seed-users/goal.md` and `requirements.md`.

## Global Constraints

- The language follows the country in this slice: `TR` users are `tr`, `US` users are `en` (D-013).
- All amounts are TRY. Keep money as `int64` kuruş in Go. Write it as `pgtype.Numeric{Int: big.NewInt(kurus), Exp: -2, Valid: true}`.
- Use `pgx` directly. Do not set up `sqlc` in this slice (D-016).
- `make test` must work with Docker and no Go (D-004). The Postgres for tests runs in a container too.
- Each database test creates its own database. Packages run in parallel, so a shared database causes flaky tests.
- No test reads `ANTHROPIC_API_KEY`.
- Do not run `docker compose down -v` on the default project. It deletes the user's local data. Use a separate project name for clean-volume checks.

## Open Decisions (not needed for this slice)

- **Stale seed data:** FR-3 skips the seed when users exist. On an old volume, "last 30 days" segments can become empty. The fix is `SE
… [9078 more chars truncated]
````

</details>

**🔧 Bash** _Read rest of plan 02 and config test style_
```bash
sed -n 97,259p docs/plans/02-seed-users.md && sed -n 50,120p backend/cmd/seed/main.go && cat backend/internal/config/config_test.go | head -40
```

<details><summary>Result</summary>

````
### Task 2: The pure generator

**Files:**
- Create: `backend/internal/seed/generate.go`, `backend/internal/seed/names.go`, `backend/internal/seed/generate_test.go`
- Modify: `backend/go.mod`, `backend/go.sum`

**Interfaces:**
- Consumes: nothing from Task 1.
- Produces:

```go
type Options struct {
    Random uint64    // SEED_RANDOM, default 42
    Now    time.Time // seed time; all times are offsets from it
    Users  int       // number of random users, default 5000
}

type User struct {
    ID                                         int64
    FirstName, LastName, Email, Phone          string
    Language, Country, City                    string
    BirthDate, CreatedAt                       time.Time
    EmailConsent, SMSConsent                   bool
    LastCartAddAt, LastPurchaseAt              *time.Time
    OrderCount                                 int
    TotalSpentKurus                            int64
}

type Event struct {
    ID, UserID  int64
    Type        string // "cart_add" or "purchase"
    AmountKurus *int64 // nil for cart_add
    OccurredAt  time.Time
}

func Generate(o Options) (users []User, events []Event)
```

- [ ] **Step 1: Write the name and city lists in `names.go`.** Write at least 60 Turkish first names and 60 Turkish last names. Include `ç`, `ğ`, `ı`, `İ`, `ö`, `ş`, `ü`. Write the city lists from `goal.md`.

- [ ] **Step 2: Write the failing tests.**
  - `TestGenerate_Deterministic`: two calls with the same options give equal slices (`reflect.DeepEqual`). A different `Random` gives different users.
  - `TestGenerate_Fields`: checks FR-6 to FR-9 and FR-12 on all users and events.
  - `TestGenerate_Ratios`: checks FR-10 with 5,000 users.
  - `TestGenerate_Summaries`: for each user, the summary fields are equal to the values from that user's events.
  - `TestGenerate_EventTimes`: each event is between `max(CreatedAt, Now-180d)` and `Now`.
  - `TestGenerate_EdgeCases`: the edge-case users from `goal.md` exist. One first name has exac
… [7451 more chars truncated]
````

</details>

## 🤖 Claude  <sub>2026-11-09 15:14:05</sub>

Now the plan:

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/docs/plans/03-mock-provider.md`
````
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
5. `limit=0`, `limit=1001`, `offset=-1`, `channel=push` on `GET /v1/messages` → `400`. Large limits must not return the full store. Pinned by Task 2 `TestMessag
… [12828 more chars truncated]
````

**🔧 Bash** _Run STE check on plan and compare lengths_
```bash
python3 .claude/skills/ste-writing/scripts/ste_check.py docs/plans/03-mock-provider.md; wc -l docs/plans/03-mock-provider.md docs/specs/03-mock-provider/*.md
```

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/docs/plans/03-mock-provider.md (file state is current in your context — no need to Read it back)
```

</details>

<details><summary>Result</summary>

```
docs/plans/03-mock-provider.md:168: sentence-length: 32 words: [ ] **Step 1: Write the failing tests.** Use a fake clock (CODE in a variable that the test moves) and a fake CODE that records each duration and returns at once.
     248 docs/plans/03-mock-provider.md
     201 docs/specs/03-mock-provider/goal.md
      32 docs/specs/03-mock-provider/requirements.md
     481 total
```

</details>

**🔧 Bash** _Fix STE finding and wording in plan_
```bash
python3 - <<'EOF'
p='docs/plans/03-mock-provider.md'
s=open(p).read()
s=s.replace("- [ ] **Step 1: Write the failing tests.** Use a fake clock (`now` in a variable that the test moves) and a fake `Sleep` that records each duration and returns at once.",
"- [ ] **Step 1: Write the failing tests.** Use a fake clock that the test moves. Use a fake `Sleep` that records each duration and returns at once.")
s=s.replace("`none` with all rates at `1.0` split as `ErrorRate: 1` gives `202`.","`none` with `ErrorRate: 1` gives `202`.")
open(p,'w').write(s)
EOF
python3 .claude/skills/ste-writing/scripts/ste_check.py docs/plans/03-mock-provider.md && echo clean
```

<details><summary>Result</summary>

```
clean
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:15:26</sub>

The plan is in [docs/plans/03-mock-provider.md](docs/plans/03-mock-provider.md). The STE check finds nothing in it.

It has four tasks:

1. **`config.LoadMockProvider`**: reads the `MOCK_*` env vars and gives the defaults. A bad value stops the process at startup with an error that names the var (FR-16).
2. **Normal path**: the handlers, validation (counted in runes), the in-memory store, and `/v1/messages`, `/v1/stats` and `DELETE /v1/messages`.
3. **Rate limit and failures**: one `x/time/rate` bucket for each channel on a fake clock, `X-Mock-Fail`, random modes from a seeded picker, and `lost` through `Hijack`.
4. **Wiring**: `cmd/mockprovider`, the compose env vars, `.env.example`, `architecture.md`, a curl check against a separate compose project, then `make verify`.

The Review Focus section lists five risks that the spec does not state, and each one has a test:

- When a client times out during a `slow` wait, the provider must still store the message.
- When many goroutines send the same `reference` at once, the duplicate count must stay correct.
- Bad JSON must get `400`, not `500`. A `500` would look like a random provider error to the worker.
- Messages with an empty `reference` must not count as duplicates.
- A bad `limit`, `offset` or `channel` on `GET /v1/messages` must get `400`.

The plan has no commit after each task. You commit the slice after your review, as in slices 00 to 02.

You can run the plan in one of two ways:

- **Subagent-driven:** the `executor` builds each task and the `reviewer` checks it before the next task starts. Slices 01 and 02 used this method.
- **Native:** I build all four tasks in this session. Then one reviewer checks the full slice.

I recommend **Native**. Tasks 2 and 3 both build up `server.go`, so separate agents would spend most of their context reading each other's work. It is also a mock service with small, isolated tests, so a mistake costs little.

Does the plan cover what you want, and which method do you want?

## 👤 User  <sub>2026-11-09 15:17:16</sub>

native

**🔧 Skill** ```json
{
  "skill": "superpowers:executing-plans"
}
```

<details><summary>Result</summary>

```
Launching skill: superpowers:executing-plans
```

</details>

> _[skill instructions loaded: executing-plans]_

## 🤖 Claude  <sub>2026-11-09 15:17:32</sub>

I'm using the executing-plans skill to build the plan in this session. I'll work in the current tree on `main` with no commits, as in slices 00 to 02, and log that as a ruling.

**🔧 Bash** _Set up plan workspace and check Go_
```bash
S=~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills; W=$($S/subagent-driven-development/scripts/sdd-workspace docs/plans/03-mock-provider.md); echo $W; ls $W; cat $W/progress.md 2>/dev/null; ls $S/executing-plans/scripts; which go; go version
```

<details><summary>Result</summary>

```
~/Desktop/projects/case-studies/insider_one/.superpowers/sdd/03-mock-provider
plan-path
task-done
task-start
/opt/homebrew/bin/go
go version go1.27.2 darwin/arm64
```

</details>

**🔧 Bash** _Create ledger with setup ruling and pre-flight_
```bash
W=.superpowers/sdd/03-mock-provider; cat > $W/progress.md <<'EOF'
# SDD ledger — plan: docs/plans/03-mock-provider.md
Setup: Ruling: work on main with no commits — plan Global Constraints and slices 00-02 do this; the user commits after review — cost if wrong: user moves changes to a branch before commit.
Pre-flight: Task 2 -> Task 3: Options/New/ChannelStats consumed as produced; no conflict. Task 1 -> Task 4: MockProviderConfig fields map 1:1 to Options; no conflict.
EOF
git check-ignore .superpowers && echo ignored
```

<details><summary>Result</summary>

```
.superpowers
ignored
```

</details>

**🔧 Skill** ```json
{
  "skill": "superpowers:test-driven-development"
}
```

<details><summary>Result</summary>

```
Launching skill: superpowers:test-driven-development
```

</details>

> _[skill instructions loaded: test-driven-development]_

## 🤖 Claude  <sub>2026-11-09 15:17:49</sub>

Task 1: the config tests first.

**🔧 Bash** _Write failing config tests and run them_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend && cat >> internal/config/config_test.go <<'EOF'

func TestLoadMockProvider_Defaults(t *testing.T) {
	cfg, err := LoadMockProvider(env(map[string]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPAddr != ":8081" {
		t.Errorf("HTTPAddr = %q, want :8081", cfg.HTTPAddr)
	}
	if cfg.RatePerSec != 100 || cfg.Burst != 10 {
		t.Errorf("rate = %v/%d, want 100/10", cfg.RatePerSec, cfg.Burst)
	}
	if cfg.ErrorRate != 0.03 || cfg.SlowRate != 0.03 || cfg.LostRate != 0.01 {
		t.Errorf("rates = %v %v %v, want 0.03 0.03 0.01", cfg.ErrorRate, cfg.SlowRate, cfg.LostRate)
	}
	if cfg.SlowDelay != 5*time.Second || cfg.LatencyMax != 50*time.Millisecond {
		t.Errorf("durations = %v %v, want 5s 50ms", cfg.SlowDelay, cfg.LatencyMax)
	}
}

func TestLoadMockProvider_Values(t *testing.T) {
	cfg, err := LoadMockProvider(env(map[string]string{
		"HTTP_ADDR":         ":9000",
		"MOCK_RATE_PER_SEC": "20.5",
		"MOCK_RATE_BURST":   "3",
		"MOCK_ERROR_RATE":   "0.1",
		"MOCK_SLOW_RATE":    "0.2",
		"MOCK_LOST_RATE":    "0.3",
		"MOCK_SLOW_DELAY":   "2s",
		"MOCK_LATENCY_MAX":  "0",
		"MOCK_RANDOM":       "7",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := MockProviderConfig{
		HTTPAddr: ":9000", RatePerSec: 20.5, Burst: 3,
		ErrorRate: 0.1, SlowRate: 0.2, LostRate: 0.3,
		SlowDelay: 2 * time.Second, LatencyMax: 0, Random: 7,
	}
	if cfg != want {
		t.Errorf("cfg = %+v, want %+v", cfg, want)
	}
}

func TestLoadMockProvider_Errors(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"error rate over 1", map[string]string{"MOCK_ERROR_RATE": "1.5"}, "MOCK_ERROR_RATE"},
		{"slow rate negative", map[string]string{"MOCK_SLOW_RATE": "-0.1"}, "MOCK_SLOW_RATE"},
		{"lost rate not a number", map[string]string{"MOCK_LOST_RATE": "abc"}, "MOCK_LOST_RATE"},
		{"sum over 1", map[string]string{"MOCK_ERROR_RATE": "0.5", "MOCK_SLOW_RATE": "0.4", "MOCK_LOST_RATE": "0.2"}, "sum"},
		{"slow delay without unit", map[string]string{"MOCK_SLOW_DELAY": "5"}, "MOCK_SLOW_DELAY"},
		{"latency negative", map[string]string{"MOCK_LATENCY_MAX": "-1s"}, "MOCK_LATENCY_MAX"},
		{"rate zero", map[string]string{"MOCK_RATE_PER_SEC": "0"}, "MOCK_RATE_PER_SEC"},
		{"burst zero", map[string]string{"MOCK_RATE_BURST": "0"}, "MOCK_RATE_BURST"},
		{"random negative", map[string]string{"MOCK_RANDOM": "-3"}, "MOCK_RANDOM"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadMockProvider(env(tc.env))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want error that mentions %q", err, tc.want)
			}
		})
	}
}
EOF
sed -i '' 's/^\t"testing"$/\t"testing"\n\t"time"/' internal/config/config_test.go
go test ./internal/config/ 2>&1 | tail -5
```

<details><summary>Result</summary>

```
internal/config/config_test.go:125:14: undefined: LoadMockProvider
internal/config/config_test.go:139:10: undefined: MockProviderConfig
internal/config/config_test.go:167:14: undefined: LoadMockProvider
FAIL	campaigncomposer/internal/config [build failed]
FAIL
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:18:07</sub>

RED as expected. Now `LoadMockProvider`:

**🔧 Bash** _Implement LoadMockProvider and run config tests_
```bash
sed -i '' 's/^\t"strconv"$/\t"strconv"\n\t"time"/' internal/config/config.go && cat >> internal/config/config.go <<'EOF'

// MockProviderConfig holds the settings that the mock provider reads at startup.
type MockProviderConfig struct {
	HTTPAddr   string
	RatePerSec float64
	Burst      int
	ErrorRate  float64
	SlowRate   float64
	LostRate   float64
	SlowDelay  time.Duration
	LatencyMax time.Duration
	// Random is the seed for the failure picker. It comes from MOCK_RANDOM,
	// or from the clock if MOCK_RANDOM is not set.
	Random uint64
}

// LoadMockProvider reads the mock provider config through getenv.
func LoadMockProvider(getenv func(string) string) (MockProviderConfig, error) {
	cfg := MockProviderConfig{
		HTTPAddr:   getenv("HTTP_ADDR"),
		RatePerSec: 100,
		Burst:      10,
		ErrorRate:  0.03,
		SlowRate:   0.03,
		LostRate:   0.01,
		SlowDelay:  5 * time.Second,
		LatencyMax: 50 * time.Millisecond,
		Random:     uint64(time.Now().UnixNano()),
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8081"
	}

	var errs []error
	if v := getenv("MOCK_RATE_PER_SEC"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f <= 0 {
			errs = append(errs, fmt.Errorf("MOCK_RATE_PER_SEC must be a number greater than 0, got %q", v))
		} else {
			cfg.RatePerSec = f
		}
	}
	if v := getenv("MOCK_RATE_BURST"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			errs = append(errs, fmt.Errorf("MOCK_RATE_BURST must be an integer of at least 1, got %q", v))
		} else {
			cfg.Burst = n
		}
	}
	for _, r := range []struct {
		key string
		dst *float64
	}{
		{"MOCK_ERROR_RATE", &cfg.ErrorRate},
		{"MOCK_SLOW_RATE", &cfg.SlowRate},
		{"MOCK_LOST_RATE", &cfg.LostRate},
	} {
		v := getenv(r.key)
		if v == "" {
			continue
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 || f > 1 {
			errs = append(errs, fmt.Errorf("%s must be a number from 0 to 1, got %q", r.key, v))
			continue
		}
		*r.dst = f
	}
	if sum := cfg.ErrorRate + cfg.SlowRate + cfg.LostRate; sum > 1 {
		errs = append(errs, fmt.Errorf("the sum of MOCK_ERROR_RATE, MOCK_SLOW_RATE and MOCK_LOST_RATE must be 1 or less, got %g", sum))
	}
	for _, d := range []struct {
		key string
		dst *time.Duration
	}{
		{"MOCK_SLOW_DELAY", &cfg.SlowDelay},
		{"MOCK_LATENCY_MAX", &cfg.LatencyMax},
	} {
		v := getenv(d.key)
		if v == "" {
			continue
		}
		dur, err := time.ParseDuration(v)
		if err != nil || dur < 0 {
			errs = append(errs, fmt.Errorf("%s must be a duration of 0 or more, for example 5s, got %q", d.key, v))
			continue
		}
		*d.dst = dur
	}
	if v := getenv("MOCK_RANDOM"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("MOCK_RANDOM must be an unsigned integer, got %q", v))
		} else {
			cfg.Random = n
		}
	}
	if len(errs) > 0 {
		return MockProviderConfig{}, errors.Join(errs...)
	}
	return cfg, nil
}
EOF
go test ./internal/config/ 2>&1 | tail -5
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/config	0.461s
```

</details>

**🔧 Bash** _Record Task 1 completion in ledger_
```bash
cd ~/Desktop/projects/case-studies/insider_one && S=~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/executing-plans/scripts; B=$(git rev-parse --short HEAD); $S/task-done docs/plans/03-mock-provider.md 1 $B -- sh -c 'cd backend && go test -count=1 ./internal/config/' 2>&1 | tail -3
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/config	0.112s
ledger: Task 1: complete (commits 68628df..68628df, tests: sh -c 'cd backend && go test -count=1 ./internal/config/' → ok  	campaigncomposer/internal/config	0.112s)
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:18:28</sub>

Task 2: the tests for the normal path first.

**🔧 Bash** _Write failing Task 2 tests and run them_
```bash
mkdir -p ~/Desktop/projects/case-studies/insider_one/backend/internal/mockprovider && cd $_ && cat > server_test.go <<'EOF'
package mockprovider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeClock is a clock that the test moves by hand.
type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{t: time.Date(2026, 11, 9, 10, 0, 0, 0, time.FixedZone("TRT", 3*3600))}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// normalOptions has no random failures, no latency and a rate limit that
// the tests do not reach.
func normalOptions(clock *fakeClock) Options {
	return Options{
		RatePerSec: 1e6,
		Burst:      1e6,
		Random:     1,
		Now:        clock.Now,
		Sleep:      func(time.Duration) {},
	}
}

func startServer(t *testing.T, o Options) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(New(o))
	t.Cleanup(srv.Close)
	return srv
}

func post(t *testing.T, base, path, body string, header map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, base+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func decode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	var v T
	b, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("decode %q: %v", b, err)
	}
	return v
}

type acceptBody struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type listBody struct {
	Total    int       `json:"total"`
	Messages []Message `json:"messages"`
}

func get(t *testing.T, base, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(base + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func listMessages(t *testing.T, base, query string) listBody {
	t.Helper()
	resp := get(t, base, "/v1/messages"+query)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /v1/messages%s: status %d", query, resp.StatusCode)
	}
	return decode[listBody](t, resp)
}

func stats(t *testing.T, base string) map[string]ChannelStats {
	t.Helper()
	return decode[map[string]ChannelStats](t, get(t, base, "/v1/stats"))
}

func sms(to, body, ref string) string {
	b, _ := json.Marshal(map[string]string{"to": to, "body": body, "reference": ref})
	return string(b)
}

func email(to, subject, body, ref string) string {
	b, _ := json.Marshal(map[string]stri
… [11397 more chars truncated]
```

<details><summary>Result</summary>

```
internal/mockprovider/server_test.go:354:14: undefined: ChannelStats
internal/mockprovider/server_test.go:357:14: undefined: ChannelStats
internal/mockprovider/server_test.go:357:14: too many errors
FAIL	campaigncomposer/internal/mockprovider [build failed]
FAIL
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:19:23</sub>

RED (compile error), as expected. I'll remove one useless assertion from the filter test, then write the store, the validation and the server.

**🔧 Bash** _Implement store, validation, server and run tests_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend/internal/mockprovider && python3 - <<'EOF'
p='server_test.go'
s=open(p).read()
s=s.replace('''	if l := listMessages(t, srv.URL, "?channel=push&x=1"[:0]); l.Messages == nil {
		t.Errorf("messages is null, want a JSON array")
	}
''','')
open(p,'w').write(s)
EOF
cat > options.go <<'EOF'
// Package mockprovider is a fake SMS and email provider. It acts like
// Twilio Messaging or SendGrid Mail Send: it delivers every valid request,
// also a repeat, and it sometimes fails, is slow, loses the response or
// returns 429. See docs/specs/03-mock-provider/goal.md.
package mockprovider

import "time"

// Options configures the provider. Now and Sleep let tests run with no real
// waits.
type Options struct {
	RatePerSec float64
	Burst      int

	ErrorRate, SlowRate, LostRate float64
	SlowDelay, LatencyMax         time.Duration

	// Random seeds the failure picker.
	Random uint64

	Now   func() time.Time    // nil means time.Now
	Sleep func(time.Duration) // nil means time.Sleep
}
EOF
cat > store.go <<'EOF'
package mockprovider

import (
	"fmt"
	"sync"
	"time"
)

// Channels.
const (
	channelSMS   = "sms"
	channelEmail = "email"
)

// Message is one delivered message.
type Message struct {
	ID         string    `json:"id"`
	Channel    string    `json:"channel"`
	To         string    `json:"to"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	Reference  string    `json:"reference"`
	ReceivedAt time.Time `json:"received_at"`
}

// ChannelStats holds the counters for one channel.
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

type filter struct {
	channel, to, reference string
	limit, offset          int
}

// store keeps all messages, references and counters in memory.
type store struct {
	mu       sync.Mutex
	nextID   int64
	messages []Message // oldest first
	refs     map[string]map[string]struct{}
	stats    map[string]*ChannelStats
}

func newStore() *store {
	s := &store{}
	s.resetLocked()
	return s
}

func (s *store) resetLocked() {
	s.nextID = 0
	s.messages = nil
	s.refs = map[string]map[string]struct{}{channelSMS: {}, channelEmail: {}}
	s.stats = map[string]*ChannelStats{channelSMS: {}, channelEmail: {}}
}

func (s *store) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// add stores m with a new ID and updates Delivered and DuplicateDeliveries.
func (s *store) add(m Message) Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	m.ID = fmt.Sprintf("%s_%08d", m.Channel, s.nextID)
	s.messages = append(s.messages, m)
	st := s.stats[m.Channel]
	st.Delivered++
	if m.Reference != "" {
		if _, seen := 
… [6868 more chars truncated]
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/mockprovider	0.897s
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:20:07</sub>

GREEN. I'll run it with `-race` a few times to check for flaky tests (the 2 MiB body, the concurrency test):

**🔧 Bash** _Run mockprovider tests with race detector 5 times_
```bash
go test -race -count=5 ./internal/mockprovider/ 2>&1 | tail -5 && go vet ./internal/mockprovider/ && gofmt -l .
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/mockprovider	4.400s
```

</details>

**🔧 Bash** _Record Task 2 rulings and completion_
```bash
cd ~/Desktop/projects/case-studies/insider_one && S=~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/executing-plans/scripts; cat >> .superpowers/sdd/03-mock-provider/progress.md <<'EOF'
Task 2: Ruling: added TestStore_IDs, TestMessages_EmptyIsArray and TestHealthz beyond the brief — the plan names store_test.go but no test for it; an empty list must be [] not null for the dashboard — cost if wrong: three extra small tests.
Task 2: Ruling: IDs use one counter across both channels (sms_00000001, email_00000002) — the spec says only "opaque, unique" — cost if wrong: none for clients.
EOF
$S/task-done docs/plans/03-mock-provider.md 2 68628df -- sh -c 'cd backend && go test -race -count=1 ./internal/mockprovider/' 2>&1 | tail -2
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/mockprovider	1.830s
ledger: Task 2: complete (commits 68628df..68628df, tests: sh -c 'cd backend && go test -race -count=1 ./internal/mockprovider/' → ok  	campaigncomposer/internal/mockprovider	1.830s)
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:20:38</sub>

Task 3: the rate-limit and failure tests first.

**🔧 Bash** _Write failing Task 3 tests and run them_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend/internal/mockprovider && cat > limit_test.go <<'EOF'
package mockprovider

import (
	"net/http"
	"strconv"
	"testing"
	"time"
)

func limitOptions(clock *fakeClock, perSec float64, burst int) Options {
	o := normalOptions(clock)
	o.RatePerSec = perSec
	o.Burst = burst
	return o
}

func TestRateLimit_Burst(t *testing.T) {
	srv := startServer(t, limitOptions(newFakeClock(), 100, 10))
	accepted, limited := 0, 0
	for range 200 {
		resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil)
		switch resp.StatusCode {
		case http.StatusAccepted:
			accepted++
		case http.StatusTooManyRequests:
			limited++
			ra, err := strconv.Atoi(resp.Header.Get("Retry-After"))
			if err != nil || ra < 1 {
				t.Fatalf("Retry-After = %q, want an integer of 1 or more", resp.Header.Get("Retry-After"))
			}
			if e := decode[errorBody](t, resp); e.Error.Code != "rate_limited" {
				t.Fatalf("code = %q, want rate_limited", e.Error.Code)
			}
		default:
			t.Fatalf("status = %d", resp.StatusCode)
		}
	}
	if accepted != 10 || limited != 190 {
		t.Fatalf("accepted %d, limited %d, want 10 and 190", accepted, limited)
	}
	st := stats(t, srv.URL)["sms"]
	if st.RateLimited != 190 || st.Requests != 200 || st.Delivered != 10 {
		t.Fatalf("sms = %+v", st)
	}
}

func TestRateLimit_Steady(t *testing.T) {
	clock := newFakeClock()
	srv := startServer(t, limitOptions(clock, 100, 10))
	for i := range 100 {
		clock.Advance(10 * time.Millisecond)
		if resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil); resp.StatusCode != http.StatusAccepted {
			t.Fatalf("request %d: status %d, want 202", i, resp.StatusCode)
		}
	}
}

func TestRateLimit_RetryAfterRoundsUp(t *testing.T) {
	// At 0.5 per second, the next token comes after 2 s.
	srv := startServer(t, limitOptions(newFakeClock(), 0.5, 1))
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusAccepted)
	resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil)
	wantStatus(t, resp, http.StatusTooManyRequests)
	if ra := resp.Header.Get("Retry-After"); ra != "2" {
		t.Fatalf("Retry-After = %q, want 2", ra)
	}
}

func TestRateLimit_PerChannel(t *testing.T) {
	srv := startServer(t, limitOptions(newFakeClock(), 100, 2))
	for range 2 {
		wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusAccepted)
	}
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusTooManyRequests)
	wantStatus(t, post(t, srv.URL, "/v1/email", email("ali@example.com", "s", "b", ""), nil), http.StatusAccepted)
}

func TestRateLimit_429UsesNoToken(t *testing.T) {
	srv := startServer(t, limitOptions(newFakeClock(), 100, 1))
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), map[string]string{"X-Mock-Fail": "429"}), http.StatusTooManyRequests)
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.
… [6871 more chars truncated]
```

<details><summary>Result</summary>

```
internal/mockprovider/failure_test.go:170:3: undefined: modeError503
internal/mockprovider/failure_test.go:171:3: undefined: modeSlow
internal/mockprovider/failure_test.go:172:3: undefined: modeLost
internal/mockprovider/failure_test.go:180:12: undefined: modeForced429
internal/mockprovider/failure_test.go:187:10: undefined: newPicker
internal/mockprovider/failure_test.go:187:10: too many errors
FAIL	campaigncomposer/internal/mockprovider [build failed]
FAIL
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:21:16</sub>

RED (compile error). First I'll write `failure.go` only, so the HTTP tests compile and fail at run time:

**🔧 Bash** _Add failure picker and run tests to see runtime failures_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend/internal/mockprovider && cat > failure.go <<'EOF'
package mockprovider

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// mode is what the provider does with one valid request.
type mode int

const (
	modeNone mode = iota
	modeError500
	modeError503
	modeSlow
	modeLost
	modeForced429 // only from X-Mock-Fail
)

func (m mode) String() string {
	switch m {
	case modeNone:
		return "none"
	case modeError500:
		return "500"
	case modeError503:
		return "503"
	case modeSlow:
		return "slow"
	case modeLost:
		return "lost"
	case modeForced429:
		return "429"
	}
	return fmt.Sprintf("mode(%d)", int(m))
}

// parseForced reads the X-Mock-Fail header. ok is false when the header is
// empty, so the random picker decides.
func parseForced(header string) (m mode, ok bool, err error) {
	switch header {
	case "":
		return modeNone, false, nil
	case "none":
		return modeNone, true, nil
	case "429":
		return modeForced429, true, nil
	case "500":
		return modeError500, true, nil
	case "503":
		return modeError503, true, nil
	case "slow":
		return modeSlow, true, nil
	case "lost":
		return modeLost, true, nil
	}
	return modeNone, false, fmt.Errorf("X-Mock-Fail must be one of none, 429, 500, 503, slow, lost, got %q", header)
}

// picker selects random failure modes and latencies from one seeded source.
type picker struct {
	mu                            sync.Mutex
	rnd                           *rand.Rand
	errorRate, slowRate, lostRate float64
}

func newPicker(o Options) *picker {
	return &picker{
		rnd:       rand.New(rand.NewPCG(o.Random, 0)),
		errorRate: o.ErrorRate,
		slowRate:  o.SlowRate,
		lostRate:  o.LostRate,
	}
}

// pick selects a mode by cumulative rate. 500 and 503 share the error rate.
func (p *picker) pick() mode {
	p.mu.Lock()
	defer p.mu.Unlock()
	x := p.rnd.Float64()
	switch {
	case x < p.errorRate:
		if p.rnd.IntN(2) == 0 {
			return modeError500
		}
		return modeError503
	case x < p.errorRate+p.slowRate:
		return modeSlow
	case x < p.errorRate+p.slowRate+p.lostRate:
		return modeLost
	}
	return modeNone
}

// latency returns a random duration from 0 to max.
func (p *picker) latency(max time.Duration) time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	return time.Duration(p.rnd.Int64N(int64(max) + 1))
}
EOF
cd ../.. && go test -count=1 ./internal/mockprovider/ 2>&1 | grep -E "^(--- FAIL|ok|FAIL)" | head -20
```

<details><summary>Result</summary>

```
--- FAIL: TestForced (0.01s)
--- FAIL: TestForced_Lost (0.00s)
--- FAIL: TestSlow_StoresAfterClientLeaves (0.00s)
--- FAIL: TestLatency (0.01s)
--- FAIL: TestRateLimit_Burst (0.02s)
--- FAIL: TestRateLimit_RetryAfterRoundsUp (0.00s)
--- FAIL: TestRateLimit_PerChannel (0.00s)
--- FAIL: TestRateLimit_429UsesNoToken (0.00s)
FAIL
FAIL	campaigncomposer/internal/mockprovider	1.101s
FAIL
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:21:29</sub>

The HTTP tests fail as expected. Now the limiter and the modes in `server.go`:

**🔧 Bash** _Wire limiter and failure modes into server, run tests_
```bash
go get golang.org/x/time/rate 2>&1 | tail -2 && python3 - <<'EOF'
p='internal/mockprovider/server.go'
s=open(p).read()
s=s.replace('''import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)''','''import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)''')
s=s.replace('''type server struct {
	o     Options
	store *store
}''','''type server struct {
	o        Options
	store    *store
	picker   *picker
	limiters map[string]*rate.Limiter
}''')
s=s.replace('''	s := &server{o: o, store: newStore()}''','''	s := &server{
		o:      o,
		store:  newStore(),
		picker: newPicker(o),
		limiters: map[string]*rate.Limiter{
			channelSMS:   rate.NewLimiter(rate.Limit(o.RatePerSec), o.Burst),
			channelEmail: rate.NewLimiter(rate.Limit(o.RatePerSec), o.Burst),
		},
	}''')
old=s[s.index('''	reject := func(msg string) {'''):s.index('''func (s *server) listMessages''')]
new='''	reject := func(msg string) {
		s.store.count(channel, func(st *ChannelStats) { st.Rejected++ })
		writeError(w, http.StatusBadRequest, "invalid_request", msg)
	}
	rateLimited := func(retryAfter int) {
		s.store.count(channel, func(st *ChannelStats) { st.RateLimited++ })
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		writeError(w, http.StatusTooManyRequests, "rate_limited", "too many requests for channel "+channel)
	}

	// 1. X-Mock-Fail.
	forced, isForced, err := parseForced(r.Header.Get("X-Mock-Fail"))
	if err != nil {
		reject(err.Error())
		return
	}

	// 2. Rate limit. A 429 uses no token.
	if forced == modeForced429 {
		rateLimited(1)
		return
	}
	now := s.o.Now()
	lim := s.limiters[channel]
	if !lim.AllowN(now, 1) {
		rateLimited(retryAfter(lim, now, s.o.RatePerSec))
		return
	}

	// 3. Parse and validate.
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody))
	if err := dec.Decode(req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			reject("request body is larger than 1 MiB")
			return
		}
		reject("request body is not valid JSON: " + err.Error())
		return
	}
	m, err := build()
	if err != nil {
		reject(err.Error())
		return
	}

	// 4. Failure mode.
	md := forced
	if !isForced {
		md = s.picker.pick()
	}

	// 5. Errors store nothing.
	switch md {
	case modeError500:
		s.store.count(channel, func(st *ChannelStats) { st.Errors++ })
		writeError(w, http.StatusInternalServerError, "provider_error", "internal provider error")
		return
	case modeError503:
		s.store.count(channel, func(st *ChannelStats) { st.Errors++ })
		writeError(w, http.StatusServiceUnavailable, "unavailable", "provider is unavailable, try again later")
		return
	}

	// 6. Wait. The wait and the store ignore r.Context(): if the client
	// leaves, the provider still delivers the message, like a real one.
	if md == modeSlow {
		s.store.count(channel, func(st *ChannelStats) { st.Slow++ })
		s.o.Sleep(s.o.SlowDelay)
	} else if s.o.LatencyMax > 0 {
		s.o.Sleep(s.picker.latency(s.o.LatencyMax))
	}

	//
… [879 more chars truncated]
```

<details><summary>Result</summary>

```
go: downloading golang.org/x/time v0.16.0
go: added golang.org/x/time v0.16.0
ok  	campaigncomposer/internal/mockprovider	1.016s
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:21:49</sub>

GREEN. Race detector, several runs, vet and gofmt:

**🔧 Bash** _Run race tests 5 times, vet, gofmt, check go.mod_
```bash
go test -race -count=5 ./internal/mockprovider/ 2>&1 | tail -5 && go vet ./... && gofmt -l . && grep -n "x/time" go.mod
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/mockprovider	6.190s
9:	golang.org/x/time v0.16.0
```

</details>

**🔧 Bash** _Record Task 3 rulings and completion_
```bash
cd ~/Desktop/projects/case-studies/insider_one && S=~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/executing-plans/scripts; cat >> .superpowers/sdd/03-mock-provider/progress.md <<'EOF'
Task 3: Ruling: added TestRateLimit_RetryAfterRoundsUp and TestRandom_SameSeedSameModes, and split lost and validation-first out of TestForced — the brief's round-up rule and MOCK_RANDOM repeatability had no direct test — cost if wrong: two extra tests.
Task 3: Ruling: when Hijack fails (HTTP/2), the handler panics with http.ErrAbortHandler — the spec says lost needs HTTP/1.1; an abort still gives the client no response — cost if wrong: an HTTP/2 client sees a stream reset, which is still a transport error.
Task 3: Ruling: the picker draws latency from the same seeded source as the modes — one seed makes a full demo run repeatable — cost if wrong: none.
EOF
$S/task-done docs/plans/03-mock-provider.md 3 68628df -- sh -c 'cd backend && go test -race -count=1 ./internal/mockprovider/' 2>&1 | tail -1
```

<details><summary>Result</summary>

```
ledger: Task 3: complete (commits 68628df..68628df, tests: sh -c 'cd backend && go test -race -count=1 ./internal/mockprovider/' → ok  	campaigncomposer/internal/mockprovider	2.180s)
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:22:09</sub>

Task 4: the binary, compose, `.env.example` and the docs.

**🔧 Bash** _Wire binary, compose env vars and .env.example_
```bash
python3 - <<'EOF'
p='backend/cmd/mockprovider/main.go'
s=open(p).read()
s=s.replace('''// Command mockprovider stands in for an external SMS and email provider.
// Slice 03-mock-provider adds the send endpoints, failures and 429s.
package main

import (
	"context"
	"encoding/json"
	"errors"''','''// Command mockprovider stands in for an external SMS and email provider.
// See internal/mockprovider and docs/specs/03-mock-provider/goal.md.
package main

import (
	"context"
	"errors"''')
s=s.replace('''	"time"
)''','''	"time"

	"campaigncomposer/internal/config"
	"campaigncomposer/internal/mockprovider"
)''',1)
old=s[s.index('''	addr := os.Getenv("HTTP_ADDR")'''):s.index('''	errc := make(chan error, 1)''')]
new='''	cfg, err := config.LoadMockProvider(os.Getenv)
	if err != nil {
		return err
	}
	addr := cfg.HTTPAddr
	slog.Info("mockprovider settings",
		"rate_per_sec", cfg.RatePerSec, "burst", cfg.Burst,
		"error_rate", cfg.ErrorRate, "slow_rate", cfg.SlowRate, "lost_rate", cfg.LostRate,
		"slow_delay", cfg.SlowDelay.String(), "latency_max", cfg.LatencyMax.String(),
		"random", cfg.Random)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// No WriteTimeout: slow responses wait for cfg.SlowDelay, and the server
	// must not end them itself.
	srv := &http.Server{
		Addr: addr,
		Handler: mockprovider.New(mockprovider.Options{
			RatePerSec: cfg.RatePerSec,
			Burst:      cfg.Burst,
			ErrorRate:  cfg.ErrorRate,
			SlowRate:   cfg.SlowRate,
			LostRate:   cfg.LostRate,
			SlowDelay:  cfg.SlowDelay,
			LatencyMax: cfg.LatencyMax,
			Random:     cfg.Random,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

'''
s=s.replace(old,new)
open(p,'w').write(s)

p='docker-compose.yml'
s=open(p).read()
old='''      args:
        CMD: mockprovider
    ports:'''
new='''      args:
        CMD: mockprovider
    env_file:
      - path: .env
        required: false
    # Failure injection and rate limit (docs/specs/03-mock-provider/goal.md).
    # Set the three rates to 0 to turn the random failures off.
    environment:
      MOCK_RATE_PER_SEC: ${MOCK_RATE_PER_SEC:-100}
      MOCK_RATE_BURST: ${MOCK_RATE_BURST:-10}
      MOCK_ERROR_RATE: ${MOCK_ERROR_RATE:-0.03}
      MOCK_SLOW_RATE: ${MOCK_SLOW_RATE:-0.03}
      MOCK_LOST_RATE: ${MOCK_LOST_RATE:-0.01}
      MOCK_SLOW_DELAY: ${MOCK_SLOW_DELAY:-5s}
      MOCK_LATENCY_MAX: ${MOCK_LATENCY_MAX:-50ms}
      # Empty means a seed from the clock.
      MOCK_RANDOM: ${MOCK_RANDOM:-}
    ports:'''
assert old in s
s=s.replace(old,new)
open(p,'w').write(s)
EOF
cat >> .env.example <<'EOF'

# Mock provider (the mock-provider service). Values below are the defaults.
# Rates are from 0 to 1, and their sum must be 1 or less. Set all three to 0
# to turn the random failures off. Leave MOCK_RANDOM empty for a seed from the clock.
# MOCK_RATE_PER_SEC=100
# MOCK_RATE_BURST=10
# MOCK_ERROR_RATE=0.03
# MOCK_SLOW_RATE=0.03
# MOCK_LOST_RATE=0.01
# MOCK_SLOW_DELAY=5s
# MOCK_LATENCY_MAX=50ms
… [136 more chars truncated]
```

<details><summary>Result</summary>

```
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("mockprovider stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadMockProvider(os.Getenv)
	if err != nil {
		return err
	}
	addr := cfg.HTTPAddr
	slog.Info("mockprovider settings",
		"rate_per_sec", cfg.RatePerSec, "burst", cfg.Burst,
		"error_rate", cfg.ErrorRate, "slow_rate", cfg.SlowRate, "lost_rate", cfg.LostRate,
		"slow_delay", cfg.SlowDelay.String(), "latency_max", cfg.LatencyMax.String(),
		"random", cfg.Random)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// No WriteTimeout: slow responses wait for cfg.SlowDelay, and the server
	// must not end them itself.
	srv := &http.Server{
		Addr: addr,
		Handler: mockprovider.New(mockprovider.Options{
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:22:26</sub>

Now `docs/architecture.md`:

**🔧 Bash** _Update architecture doc and check NFR-6 imports_
```bash
cd ~/Desktop/projects/case-studies/insider_one && python3 - <<'EOF'
p='docs/architecture.md'
s=open(p).read()
s=s.replace("- **mock-provider** — Go service that acts as an SMS and email provider.",
"- **mock-provider** — Go service that acts as an SMS and email provider, like Twilio or SendGrid. It sometimes returns `429`, `5xx`, slow responses and lost responses. It does no dedup (D-018).")
s=s.replace("    delivery/       queue, rate limiter, retry, idempotency\n",
"    delivery/       queue, rate limiter, retry, idempotency\n    mockprovider/   mock provider handlers, rate limit, failure modes, message store\n")
s=s.replace("""4. **Send** — A queue sends the messages to the mock provider.
   - TBD per slice.""","""4. **Send** — A queue sends the messages to the mock provider.
   - The mock provider delivers each request, also a repeat. The worker owns the dedup. It does not retry after a timeout or a dropped connection (D-018, D-019).""")
open(p,'w').write(s)
EOF
grep -n "mockprovider\|D-018" docs/architecture.md; python3 .claude/skills/ste-writing/scripts/ste_check.py docs/architecture.md
cd backend && go list -f '{{join .Imports "\n"}}' ./cmd/mockprovider ./internal/mockprovider | grep campaigncomposer
```

<details><summary>Result</summary>

```
6:- **mock-provider** — Go service that acts as an SMS and email provider, like Twilio or SendGrid. It sometimes returns `429`, `5xx`, slow responses and lost responses. It does no dedup (D-018).
19:    mockprovider/   mock SMS and email provider
28:    mockprovider/   mock provider handlers, rate limit, failure modes, message store
53:   - The mock provider delivers each request, also a repeat. The worker owns the dedup. It does not retry after a timeout or a dropped connection (D-018, D-019).
campaigncomposer/internal/config
campaigncomposer/internal/mockprovider
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:22:33</sub>

NFR-6 passes. Now the end-to-end check with a separate compose project:

**🔧 Bash** _Check whether port 8081 is in use_
```bash
cd ~/Desktop/projects/case-studies/insider_one && docker compose ps --format '{{.Service}} {{.State}}' 2>&1; lsof -iTCP:8081 -sTCP:LISTEN 2>/dev/null | head -3
```

<details><summary>Result</summary>

```
postgres running
```

</details>

**🔧 Bash** _Start mock-provider in a separate project and check it with curl_
```bash
docker compose -p cc-verify up --build -d --wait mock-provider > /tmp/claude-501/cc-verify-up.log 2>&1; echo "exit=$?"; tail -3 /tmp/claude-501/cc-verify-up.log
curl -s -o /dev/null -w '%{http_code}\n' -X POST localhost:8081/v1/sms -H 'X-Mock-Fail: none' -d '{"to":"+905321234567","body":"Merhaba Ayşe","reference":"c1:u1:sms"}'
curl -s -o /dev/null -w '%{http_code}\n' -X POST localhost:8081/v1/sms -H 'X-Mock-Fail: 503' -d '{"to":"+905321234567","body":"x"}'
curl -s -i -X POST localhost:8081/v1/sms -H 'X-Mock-Fail: 429' -d '{"to":"+905321234567","body":"x"}' | grep -i -e "^HTTP" -e retry-after
curl -s -o /dev/null -w 'lost: %{http_code} exit=' -X POST localhost:8081/v1/email -H 'X-Mock-Fail: lost' -d '{"to":"a@example.com","subject":"s","body":"b"}'; echo $?
curl -s localhost:8081/v1/stats; echo
docker compose -p cc-verify logs mock-provider | grep settings | head -1
docker compose -p cc-verify down -v 2>&1 | tail -1
```

<details><summary>Result</summary>

```
exit=0
 Container cc-verify-mock-provider-1 Started 
 Container cc-verify-mock-provider-1 Waiting 
 Container cc-verify-mock-provider-1 Healthy 
202
503
HTTP/1.1 429 Too Many Requests
Retry-After: 1
lost: 000 exit=52
{"email":{"requests":1,"delivered":1,"rejected":0,"rate_limited":0,"errors":0,"slow":0,"lost":1,"duplicate_deliveries":0},"sms":{"requests":3,"delivered":1,"rejected":0,"rate_limited":1,"errors":1,"slow":0,"lost":0,"duplicate_deliveries":0}}

mock-provider-1  | {"time":"2026-11-09T15:22:47.956037389Z","level":"INFO","msg":"mockprovider settings","rate_per_sec":100,"burst":10,"error_rate":0.03,"slow_rate":0.03,"lost_rate":0.01,"slow_delay":"5s","latency_max":"50ms","random":1794237767956030722}
 Network cc-verify_default Removed
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:22:52</sub>

Everything matches: `202`, `503`, `429` with `Retry-After: 1`, a lost response with no reply (curl exit 52), and the stats agree. Now `make verify`:

**🔧 Bash** _Run make verify and show summary_
```bash
rm -f /tmp/claude-501/cc-verify-up.log; make verify > .superpowers/sdd/03-mock-provider/verify.log 2>&1; echo "exit=$?"; grep -E "^(ok|FAIL|---|\?)" .superpowers/sdd/03-mock-provider/verify.log | head -20; tail -5 .superpowers/sdd/03-mock-provider/verify.log
```

<details><summary>Result</summary>

```
exit=0
?   	campaigncomposer/cmd/api	[no test files]
?   	campaigncomposer/cmd/mockprovider	[no test files]
?   	campaigncomposer/cmd/seed	[no test files]
ok  	campaigncomposer/internal/config	(cached)
ok  	campaigncomposer/internal/httpapi	(cached)
ok  	campaigncomposer/internal/mockprovider	0.867s
ok  	campaigncomposer/internal/seed	(cached)
ok  	campaigncomposer/internal/store	(cached)
?   	campaigncomposer/internal/store/storetest	[no test files]
dist/assets/default-ChzNtVzf.js                    182.62 kB │ gzip:  50.89 kB
dist/assets/index-BY7dSr2k.js                      309.50 kB │ gzip: 106.53 kB

✓ built in 467ms
docker compose config -q
```

</details>

**🔧 Bash** _Record Task 4 and build review package_
```bash
S=~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills; cat >> .superpowers/sdd/03-mock-provider/progress.md <<'EOF'
Task 4: NFR-6 import check: only internal/config and internal/mockprovider. Step 5: exit=0, 202, 503, 429 Retry-After 1, lost = curl exit 52, stats match.
EOF
$S/executing-plans/scripts/task-done docs/plans/03-mock-provider.md 4 68628df -- make verify 2>&1 | tail -1
$S/subagent-driven-development/scripts/review-package docs/plans/03-mock-provider.md 68628df HEAD 2>&1 | tail -3
```

<details><summary>Result</summary>

```
ledger: Task 4: complete (commits 68628df..68628df, tests: make verify → docker compose config -q)
empty commit range: 68628df..HEAD
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:23:35</sub>

There are no commits, so I'll give the reviewer the diff of the working tree. As agreed, one fresh reviewer now checks the whole slice.

**🔧 Bash** _Write working-tree diff for the reviewer_
```bash
W=.superpowers/sdd/03-mock-provider; git add -N backend/internal/mockprovider docs/specs/03-mock-provider docs/plans/03-mock-provider.md && git diff -- . ':!.gitignore' > $W/review.diff; git reset -q backend/internal/mockprovider docs/specs/03-mock-provider docs/plans/03-mock-provider.md; wc -l $W/review.diff; sed -n 1,20p ~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/requesting-code-review/code-reviewer.md
```

<details><summary>Result</summary>

````
2286 .superpowers/sdd/03-mock-provider/review.diff
# Code Reviewer Prompt Template

Use this template when dispatching a code reviewer subagent.

**Purpose:** Review completed work against requirements and code quality standards before it cascades into more work.

```
Subagent (general-purpose):
  description: "Review code changes"
  prompt: |
    You are a Senior Code Reviewer with expertise in software architecture,
    design patterns, and best practices. Your job is to review completed work
    against its plan or requirements and identify issues before they cascade.

    ## What Was Implemented

    [DESCRIPTION]

    ## Requirements / Plan
````

</details>

**🔧 Agent** **Review slice 03 mock provider** (reviewer)
```
Review the finished slice 03-mock-provider in ~/Desktop/projects/case-studies/insider_one. Read-only: do not edit any file.

What was built: a mock SMS and email provider (like Twilio Messaging / SendGrid). Package backend/internal/mockprovider (options.go, store.go, validate.go, failure.go, server.go + tests), config.LoadMockProvider in backend/internal/config, wiring in backend/cmd/mockprovider/main.go, env vars in docker-compose.yml and .env.example, docs/architecture.md, docs/decisions-log.md D-018 to D-021.

Inputs:
- Spec (binding authority): docs/specs/03-mock-provider/goal.md and requirements.md
- Plan: docs/plans/03-mock-provider.md
- Hard rules: CLAUDE.md
- Full diff of the working tree (nothing is committed): .superpowers/sdd/03-mock-provider/review.diff
- Ledger with the implementer's rulings (lines with "Ruling:"): .superpowers/sdd/03-mock-provider/progress.md. Weigh each ruling.

Evidence already gathered: `make verify` exit 0; `go test -race -count=5 ./internal/mockprovider/` passes; a compose run of mock-provider returned 202, 503, 429 with Retry-After 1, a lost response (curl exit 52), and matching /v1/stats.

Check each of these Review Focus items from the plan on purpose:
1. A client times out during a slow wait → the provider must still store the message.
2. The same reference from many goroutines at once → duplicate_deliveries must be delivered - 1.
3. JSON with wrong types or non-JSON body → 400 invalid_request, not 500.
4. An empty reference on many messages → not duplicates.
5. limit=0, limit=1001, offset=-1, channel=push on GET /v1/messages → 400.

Also look for: ordering of checks vs goal.md "Order of checks"; counter correctness in every path (requests, rejected, rate_limited, errors, slow, lost, delivered); rate limiter and Retry-After correctness with x/time/rate; hijack/lost behaviour and anything that could make the Go client auto-retry; graceful shutdown with slow requests in flight; flaky tests; spec FR/NFR items with no test.

Report: findings graded Critical / Important / Minor, each with file:line, the concrete failure scenario, and a suggested fix. Then a list "Declined to judge" for anything you set aside. Keep the report under 600 words.
```

<details><summary>Result</summary>

```
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a3a76f30ac0145c6a (internal ID - do not mention to user. Use SendMessage with to: 'a3a76f30ac0145c6a', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/b8a1807d-a6d6-4467-96a4-6fbd029e615a/tasks/a3a76f30ac0145c6a.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:23:51</sub>

All four tasks pass, and `make verify` exits 0. A reviewer is now checking the whole slice against the spec and the plan. I'll sort its findings and fix the important ones when the review comes back.

## 🤖 Claude  <sub>2026-11-09 15:26:09</sub>

I'm re-grading the findings by what they do to someone who uses the provider:

- **Fix (Important):**
  - `NaN` and `Inf` pass the config check.
  - IDs repeat after a reset.
  - The NFR-1 test never runs with failures, so that backlog acceptance check is not really verified. The reviewer said Minor, but I re-graded it.
  - The `env_file` gives `ANTHROPIC_API_KEY` to a container that has no use for it. The reviewer said Minor, but I re-graded it. That line came from my plan.
- **Defer (Minor):** trailing JSON after the body, the shutdown timeout, and two test-strength notes.

RED first:

**🔧 Bash** _Write regression tests for review findings and run them_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend && python3 - <<'EOF'
p='internal/config/config_test.go'
s=open(p).read()
s=s.replace('''		{"random negative", map[string]string{"MOCK_RANDOM": "-3"}, "MOCK_RANDOM"},''','''		{"random negative", map[string]string{"MOCK_RANDOM": "-3"}, "MOCK_RANDOM"},
		{"error rate NaN", map[string]string{"MOCK_ERROR_RATE": "NaN"}, "MOCK_ERROR_RATE"},
		{"slow rate Inf", map[string]string{"MOCK_SLOW_RATE": "Inf"}, "MOCK_SLOW_RATE"},
		{"rate per sec Inf", map[string]string{"MOCK_RATE_PER_SEC": "+Inf"}, "MOCK_RATE_PER_SEC"},
		{"rate per sec NaN", map[string]string{"MOCK_RATE_PER_SEC": "NaN"}, "MOCK_RATE_PER_SEC"},''')
open(p,'w').write(s)

p='internal/mockprovider/server_test.go'
s=open(p).read()
s=s.replace('''	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "a", "r1"), nil), http.StatusAccepted)
	if d := stats(t, srv.URL)["sms"].DuplicateDeliveries; d != 0 {
		t.Fatalf("duplicate_deliveries = %d, want 0 after reset", d)
	}
}''','''	resp = post(t, srv.URL, "/v1/sms", sms("+905321234567", "a", "r1"), nil)
	wantStatus(t, resp, http.StatusAccepted)
	if d := stats(t, srv.URL)["sms"].DuplicateDeliveries; d != 0 {
		t.Fatalf("duplicate_deliveries = %d, want 0 after reset", d)
	}
	// An id is unique for each delivery, also across a reset.
	if id := decode[acceptBody](t, resp).ID; id == "sms_00000001" {
		t.Fatalf("id %q was used before the reset", id)
	}
}''')
s=s.replace('''func TestHealthz(''','''// TestConcurrent_CountersWithFailures checks NFR-1 with all modes active:
// delivered is the count of 202 responses plus lost responses.
func TestConcurrent_CountersWithFailures(t *testing.T) {
	o := normalOptions(newFakeClock())
	o.ErrorRate, o.SlowRate, o.LostRate = 0.2, 0.2, 0.2
	srv := startServer(t, o)
	var (
		wg               sync.WaitGroup
		mu               sync.Mutex
		accepted, failed int64
		transport        int64
	)
	for g := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 10 {
				body := sms("+905321234567", fmt.Sprintf("g%d-%d", g, i), "")
				resp, err := http.Post(srv.URL+"/v1/sms", "application/json", strings.NewReader(body))
				mu.Lock()
				switch {
				case err != nil:
					transport++
				case resp.StatusCode == http.StatusAccepted:
					accepted++
				default:
					failed++
				}
				mu.Unlock()
				if err == nil {
					resp.Body.Close()
				}
			}
		}()
	}
	wg.Wait()
	st := stats(t, srv.URL)["sms"]
	if st.Lost == 0 || st.Errors == 0 || st.Slow == 0 {
		t.Fatalf("sms = %+v, want each failure mode at least once", st)
	}
	if st.Delivered != accepted+st.Lost {
		t.Fatalf("delivered %d, want 202 count %d + lost %d", st.Delivered, accepted, st.Lost)
	}
	if transport != st.Lost || failed != st.Errors || st.Requests != 500 {
		t.Fatalf("transport errors %d, other failures %d, sms = %+v", transport, failed, st)
	}
}

func TestHealthz(''')
open(p,'w').write(s)
EOF
go test -count=1 ./internal/config/ ./internal/mockprovider/ 2>&1 | grep -E "^(\s+--- F
… [47 more chars truncated]
```

<details><summary>Result</summary>

```
--- FAIL: TestLoadMockProvider_Errors (0.00s)
    --- FAIL: TestLoadMockProvider_Errors/error_rate_NaN (0.00s)
        config_test.go:173: err = <nil>, want error that mentions "MOCK_ERROR_RATE"
    --- FAIL: TestLoadMockProvider_Errors/rate_per_sec_Inf (0.00s)
        config_test.go:173: err = <nil>, want error that mentions "MOCK_RATE_PER_SEC"
    --- FAIL: TestLoadMockProvider_Errors/rate_per_sec_NaN (0.00s)
        config_test.go:173: err = <nil>, want error that mentions "MOCK_RATE_PER_SEC"
FAIL
FAIL	campaigncomposer/internal/config	0.341s
--- FAIL: TestReset (0.00s)
    server_test.go:386: id "sms_00000001" was used before the reset
FAIL
FAIL	campaigncomposer/internal/mockprovider	1.026s
FAIL
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:26:29</sub>

Three findings are RED. The new NFR-1 test passes at once, which shows that the counters were already correct with failures. Now the fixes:

**🔧 Bash** _Fix NaN/Inf validation and ID reuse, rerun tests_
```bash
python3 - <<'EOF'
p='internal/config/config.go'
s=open(p).read()
s=s.replace('''		if err != nil || f <= 0 {
			errs = append(errs, fmt.Errorf("MOCK_RATE_PER_SEC''','''		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) || f <= 0 {
			errs = append(errs, fmt.Errorf("MOCK_RATE_PER_SEC''')
s=s.replace('''		if err != nil || f < 0 || f > 1 {''','''		if err != nil || math.IsNaN(f) || f < 0 || f > 1 {''')
s=s.replace('''	"fmt"
	"strconv"''','''	"fmt"
	"math"
	"strconv"''')
open(p,'w').write(s)
p='internal/mockprovider/store.go'
s=open(p).read()
s=s.replace('''func (s *store) resetLocked() {
	s.nextID = 0
	s.messages = nil''','''// resetLocked keeps nextID, so an id is never used twice, also across a reset.
func (s *store) resetLocked() {
	s.messages = nil''')
open(p,'w').write(s)
EOF
go test -race -count=3 ./internal/config/ ./internal/mockprovider/ 2>&1 | tail -3
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/config	1.256s
ok  	campaigncomposer/internal/mockprovider	4.419s
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:26:41</sub>

GREEN. Now the `env_file` finding. First a RED check against the compose config:

**🔧 Bash** _Check whether the API key reaches mock-provider env_
```bash
cd ~/Desktop/projects/case-studies/insider_one && echo 'ANTHROPIC_API_KEY=[REDACTED]' > .superpowers/sdd/03-mock-provider/probe.env && check() { docker compose --env-file .superpowers/sdd/03-mock-provider/probe.env config --format json 2>/dev/null | python3 -c 'import json,sys; e=json.load(sys.stdin)["services"]["mock-provider"].get("environment",{}); print("RED: key reaches mock-provider" if "ANTHROPIC_API_KEY" in e else "GREEN: no key in mock-provider")'; }; check; ls .env 2>/dev/null
```

<details><summary>Result (error)</summary>

```
Exit code 1
GREEN: no key in mock-provider
```

</details>

<details><summary>Thinking</summary>

That probe was wrong: `--env-file` only feeds interpolation, while `env_file` still reads `.env` directly. Since there's no `.env` present, I'm creating a temporary one with a dummy key and deleting it in the same command.

</details>

**🔧 Bash** _Probe compose with a temporary dummy .env, then delete it_
```bash
cd ~/Desktop/projects/case-studies/insider_one && test ! -e .env && echo 'ANTHROPIC_API_KEY=[REDACTED]' > .env && docker compose config --format json 2>/dev/null | python3 -c 'import json,sys; e=json.load(sys.stdin)["services"]["mock-provider"].get("environment",{}); print("RED: key reaches mock-provider" if "ANTHROPIC_API_KEY" in e else "GREEN: no key in mock-provider")'; rm -f .env; ls .env 2>&1 | head -1
```

<details><summary>Result</summary>

```
RED: key reaches mock-provider
ls: .env: No such file or directory
```

</details>

**🔧 Bash** _Remove env_file from mock-provider and re-probe_
```bash
python3 - <<'EOF'
p='docker-compose.yml'
s=open(p).read()
old='''        CMD: mockprovider
    env_file:
      - path: .env
        required: false
    # Failure injection and rate limit (docs/specs/03-mock-provider/goal.md).
    # Set the three rates to 0 to turn the random failures off.'''
new='''        CMD: mockprovider
    # Failure injection and rate limit (docs/specs/03-mock-provider/goal.md).
    # Set the three rates to 0 to turn the random failures off. Compose reads
    # the values from .env. No env_file: the provider must not get the API key.'''
assert old in s
s=s.replace(old,new)
open(p,'w').write(s)
p='docs/plans/03-mock-provider.md'
s=open(p).read()
s=s.replace("  - `mock-provider` gets `env_file` with `path: .env` and `required: false`, as `api` has.\n","  - `mock-provider` gets no `env_file`. Compose reads the `${...}` values from `.env`, and the provider must not get `ANTHROPIC_API_KEY`.\n")
open(p,'w').write(s)
EOF
test ! -e .env && printf 'ANTHROPIC_API_KEY=[REDACTED]' > .env && docker compose config --format json 2>/dev/null | python3 -c 'import json,sys; e=json.load(sys.stdin)["services"]["mock-provider"]["environment"]; print("RED" if "ANTHROPIC_API_KEY" in e else "GREEN: no key", "| MOCK_ERROR_RATE =", e["MOCK_ERROR_RATE"])'; rm -f .env; ls .env 2>&1 | head -1
```

<details><summary>Result</summary>

```
GREEN: no key | MOCK_ERROR_RATE = 0.5
ls: .env: No such file or directory
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:27:06</sub>

GREEN, and values from `.env` still reach the provider. Now the full suite:

**🔧 Bash** _Run make verify after the fix pass_
```bash
make verify > .superpowers/sdd/03-mock-provider/verify2.log 2>&1; echo "exit=$?"; grep -E "^(ok|FAIL|---)" .superpowers/sdd/03-mock-provider/verify2.log; grep -E "Tests +[0-9]+ passed" .superpowers/sdd/03-mock-provider/verify2.log; python3 .claude/skills/ste-writing/scripts/ste_check.py docs/plans/03-mock-provider.md
```

<details><summary>Result</summary>

```
exit=0
ok  	campaigncomposer/internal/config	0.244s
ok  	campaigncomposer/internal/httpapi	(cached)
ok  	campaigncomposer/internal/mockprovider	1.011s
ok  	campaigncomposer/internal/seed	(cached)
ok  	campaigncomposer/internal/store	(cached)
      Tests  55 passed (55)
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:27:24</sub>

`make verify` exits 0. Now I'll record the work in `tasks/progress.md` and clean up the workspace:

**🔧 Bash** _Write ledger, progress entry, delete workspace, show status_
```bash
cat >> .superpowers/sdd/03-mock-provider/progress.md <<'EOF'
Final review: reviewer subagent (opus). 0 Critical, 2 Important, 6 Minor.
Final: fixed NaN/Inf accepted by LoadMockProvider — TestLoadMockProvider_Errors NaN/Inf cases RED→GREEN, make verify exit 0
Final: fixed id reuse after DELETE /v1/messages — TestReset id check RED→GREEN, make verify exit 0
Final: fixed (re-graded Minor→Important) env_file gave ANTHROPIC_API_KEY to mock-provider — compose config probe RED→GREEN, make verify exit 0
Final: fixed (re-graded Minor→Important) NFR-1 untested with failures — TestConcurrent_CountersWithFailures added; passed at once, code was correct
Final: minor (deferred): trailing data after the JSON object is accepted, also 2 MiB of trailing spaces
Final: minor (deferred): shutdownTimeout 10s is too short if MOCK_SLOW_DELAY > 10s
Final: minor (deferred): TestSlow_StoresAfterClientLeaves does not wait for r.Context().Done()
Final: minor (deferred): TestRateLimit_429UsesNoToken does not check the bucket after the third request
EOF
grep "Ruling:" .superpowers/sdd/03-mock-provider/progress.md | wc -l
cat >> tasks/progress.md <<'EOF'

## 2026-11-09 — 03-mock-provider
- Done: spec `docs/specs/03-mock-provider/`, plan `docs/plans/03-mock-provider.md` and Tasks 1 to 4. Added `config.LoadMockProvider`, the package `internal/mockprovider` (send endpoints, a token bucket for each channel, random and forced failures, an in-memory store, `/v1/messages` and `/v1/stats`) and the `MOCK_*` env vars in compose. User decisions: the provider does no dedup, as Twilio and SendGrid (D-018). After a timeout or a dropped connection, the worker marks `unknown` and does not retry (D-019). Added D-020 and D-021. The final review found 2 Important defects, and both are fixed: `NaN` passed the rate check, and IDs were used again after a reset. I re-graded 2 Minor findings to Important and fixed them: `env_file` gave the API key to `mock-provider`, and NFR-1 had no test with failures.
- Verified with: `make verify` exit 0. `go test -race -count=5 ./internal/mockprovider/` passes. A compose run with `-p cc-verify` gave `202`, `503`, `429` with `Retry-After: 1`, a lost response (curl exit 52) and matching stats.
- Next: user review of slice 03. The user sets `03-mock-provider.passes`. Then slice 04-segment-rules.
- Open questions: four Minor items stay open. The decoder accepts data after the JSON object. The 10 s shutdown timeout is too short for `MOCK_SLOW_DELAY` over 10 s. Two tests can be stronger (slow disconnect, bucket after a `429`). Slice 07 must not send `Idempotency-Key` or `X-Idempotency-Key`.
EOF
python3 .claude/skills/ste-writing/scripts/ste_check.py tasks/progress.md | grep -n "03-mock\|2026-11-09" ; python3 .claude/skills/ste-writing/scripts/ste_check.py tasks/progress.md | tail -5
rm -rf .superpowers/sdd/03-mock-provider && git status --short
```

<details><summary>Result</summary>

```
6
tasks/progress.md:26: sentence-length: 33 words: All 3 are fixed: a 2 s timeout on the startup DB ping, no exit 0 when a listen error races a signal, and a second Ctrl-C that stops the process at once.
tasks/progress.md:32: sentence-length: 28 words: The plan was updated for these decisions and for a new Step 5 check, because Compose v5 prints CODE for CODE on a port that is not published.
tasks/progress.md:32: sentence-length: 40 words: All are fixed: CODE under Make 3.81, a gofmt parse error that CODE ignored, CODE on the socket during first init, an unquoted volume path, an ARG that split the module cache, the Step 5 check, and the plan snippet.
tasks/progress.md:50: sentence-length: 32 words: Fixed 6: the pnpm store in the host folder, root-owned files on Linux, the nginx redirect port, a double Cache-Control header, missing files that returned the app, and the NFR-3 acceptance text.
tasks/progress.md:62: sentence-length: 30 words: Added CODE, the package CODE (send endpoints, a token bucket for each channel, random and forced failures, an in-memory store, CODE and CODE) and the CODE env vars in compose.
 M .claude/agent-memory/reviewer/MEMORY.md
 M .env.example
 M .gitignore
 M backend/cmd/mockprovider/main.go
 M backend/go.mod
 M backend/go.sum
 M backend/internal/config/config.go
 M backend/internal/config/config_test.go
 M docker-compose.yml
 M docs/architecture.md
 M docs/decisions-log.md
 M tasks/backlog.json
 M tasks/progress.md
?? .claude/agent-memory/reviewer/http-config-validation-defects.md
?? backend/internal/mockprovider/
?? docs/plans/03-mock-provider.md
?? docs/specs/03-mock-provider/
```

</details>

**🔧 Bash** _Fix long sentence in the new progress entry_
```bash
python3 - <<'EOF'
p='tasks/progress.md'
s=open(p).read()
s=s.replace("Added `config.LoadMockProvider`, the package `internal/mockprovider` (send endpoints, a token bucket for each channel, random and forced failures, an in-memory store, `/v1/messages` and `/v1/stats`) and the `MOCK_*` env vars in compose.",
"Added `config.LoadMockProvider` and the `MOCK_*` env vars in compose. Added the package `internal/mockprovider`. It has the send endpoints, a token bucket for each channel, random and forced failures, an in-memory store, `/v1/messages` and `/v1/stats`.")
open(p,'w').write(s)
EOF
python3 .claude/skills/ste-writing/scripts/ste_check.py tasks/progress.md | grep -c ":6[0-9]:"
```

<details><summary>Result</summary>

```
0
```

</details>

## 🤖 Claude  <sub>2026-11-09 15:27:53</sub>

Slice 03 is built, and `make verify` passes (exit 0). I did not commit anything, and I did not set `03-mock-provider.passes`. Both wait for your review.

**What the provider does now**
- `POST /v1/sms` and `POST /v1/email` deliver every valid request, also a repeat. The provider does no dedup, like Twilio and SendGrid.
- Each channel has a rate limit of 100/s. Over the limit, it returns `429` with `Retry-After`.
- It sometimes fails at random: `500`/`503` (nothing is delivered), slow (delivered after a delay), or lost (delivered, then the connection closes with no response).
- The `X-Mock-Fail` header forces one of these results for one request, so tests can hit each path exactly.
- Characters are counted as runes, so `ş` is 1. An SMS body over 160 or an email subject over 80 gets `400`.
- `GET /v1/messages` and `GET /v1/stats` are the audit, including `duplicate_deliveries`. `DELETE /v1/messages` resets them.

**Checks**
- The tests passed five times in a row with Go's race detector.
- I started `mock-provider` alone in a separate compose project. curl got `202`, `503`, `429` with `Retry-After: 1`, and a lost response with no reply. The stats matched these responses.

**Review.** A fresh reviewer read the whole slice and found 0 Critical, 2 Important and 6 Minor defects. For each fix I first wrote a test that failed, then made it pass:
- A rate of `NaN` or `Inf` let the process start. It now stops with an error.
- Message IDs were used again after `DELETE /v1/messages`. They now stay unique.
- I moved two of the reviewer's Minor findings up to Important:
  - The `mock-provider` container got `ANTHROPIC_API_KEY` from `.env`. I removed its `env_file` line. To check it, I made a temporary `.env` with a dummy key and deleted it right after.
  - The concurrency requirement (NFR-1) had no test with failures turned on. I added one, and it passed at once, so the counting code was already correct.

**Calls I made**
- I worked on `main` with no commits, as in slices 00 to 02.
- I added five tests that the plan did not list: ID format, an empty list returns `[]`, `/healthz`, `Retry-After` rounds up, and the same seed gives the same failures.
- IDs use one counter for both channels, for example `sms_00000001`, then `email_00000002`.
- If the "lost" mode cannot close the connection directly (HTTP/2), it aborts the response. The client still gets no response.
- The same random seed sets both the failure modes and the latency, so one `MOCK_RANDOM` repeats a whole demo run.

**Minor findings left open** (also in `tasks/progress.md`)
- The provider accepts extra data after the JSON object.
- The 10 s shutdown timeout is too short if you set `MOCK_SLOW_DELAY` above 10 s.
- Two tests could be stronger: the slow-disconnect test and the bucket check after a `429`.

**Other changes in your tree**
- The `.gitignore` change was there before this session.
- The reviewer wrote a note for itself in `.claude/agent-memory/reviewer/`.

Do you want me to commit the slice?


---

## Subagent transcripts


### Subagent: Review slice 03 mock provider (reviewer)

#### 👤 User  <sub>2026-11-09 15:23:49</sub>

Review the finished slice 03-mock-provider in ~/Desktop/projects/case-studies/insider_one. Read-only: do not edit any file.

What was built: a mock SMS and email provider (like Twilio Messaging / SendGrid). Package backend/internal/mockprovider (options.go, store.go, validate.go, failure.go, server.go + tests), config.LoadMockProvider in backend/internal/config, wiring in backend/cmd/mockprovider/main.go, env vars in docker-compose.yml and .env.example, docs/architecture.md, docs/decisions-log.md D-018 to D-021.

Inputs:
- Spec (binding authority): docs/specs/03-mock-provider/goal.md and requirements.md
- Plan: docs/plans/03-mock-provider.md
- Hard rules: CLAUDE.md
- Full diff of the working tree (nothing is committed): .superpowers/sdd/03-mock-provider/review.diff
- Ledger with the implementer's rulings (lines with "Ruling:"): .superpowers/sdd/03-mock-provider/progress.md. Weigh each ruling.

Evidence already gathered: `make verify` exit 0; `go test -race -count=5 ./internal/mockprovider/` passes; a compose run of mock-provider returned 202, 503, 429 with Retry-After 1, a lost response (curl exit 52), and matching /v1/stats.

Check each of these Review Focus items from the plan on purpose:
1. A client times out during a slow wait → the provider must still store the message.
2. The same reference from many goroutines at once → duplicate_deliveries must be delivered - 1.
3. JSON with wrong types or non-JSON body → 400 invalid_request, not 500.
4. An empty reference on many messages → not duplicates.
5. limit=0, limit=1001, offset=-1, channel=push on GET /v1/messages → 400.

Also look for: ordering of checks vs goal.md "Order of checks"; counter correctness in every path (requests, rejected, rate_limited, errors, slow, lost, delivered); rate limiter and Retry-After correctness with x/time/rate; hijack/lost behaviour and anything that could make the Go client auto-retry; graceful shutdown with slow requests in flight; flaky tests; spec FR/NFR items with no test.

Report: findings graded Critical / Important / Minor, each with file:line, the concrete failure scenario, and a suggested fix. Then a list "Declined to judge" for anything you set aside. Keep the report under 600 words.

**🔧 Bash** _Read spec, diff size, rulings_
```bash
cd ~/Desktop/projects/case-studies/insider_one; cat docs/specs/03-mock-provider/goal.md docs/specs/03-mock-provider/requirements.md; wc -l .superpowers/sdd/03-mock-provider/review.diff; grep -n "Ruling" .superpowers/sdd/03-mock-provider/progress.md
```

<details><summary>Result</summary>

````
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
| `to` | string | R
… [12339 more chars truncated]
````

</details>

**🔧 Bash** _Read mockprovider sources_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend; cat -n internal/mockprovider/server.go internal/mockprovider/failure.go internal/mockprovider/store.go internal/mockprovider/validate.go internal/mockprovider/options.go
```

**🔧 Bash** _Read main and config diff_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend; cat -n cmd/mockprovider/main.go; git diff internal/config/config.go
```

<details><summary>Result</summary>

```
1	package mockprovider
     2	
     3	import (
     4		"encoding/json"
     5		"errors"
     6		"math"
     7		"net/http"
     8		"strconv"
     9		"time"
    10	
    11		"golang.org/x/time/rate"
    12	)
    13	
    14	const maxRequestBody = 1 << 20
    15	
    16	type server struct {
    17		o        Options
    18		store    *store
    19		picker   *picker
    20		limiters map[string]*rate.Limiter
    21	}
    22	
    23	// New returns the provider handler with all routes from the spec,
    24	// including GET /healthz.
    25	func New(o Options) http.Handler {
    26		if o.Now == nil {
    27			o.Now = time.Now
    28		}
    29		if o.Sleep == nil {
    30			o.Sleep = time.Sleep
    31		}
    32		s := &server{
    33			o:      o,
    34			store:  newStore(),
    35			picker: newPicker(o),
    36			limiters: map[string]*rate.Limiter{
    37				channelSMS:   rate.NewLimiter(rate.Limit(o.RatePerSec), o.Burst),
    38				channelEmail: rate.NewLimiter(rate.Limit(o.RatePerSec), o.Burst),
    39			},
    40		}
    41	
    42		mux := http.NewServeMux()
    43		mux.HandleFunc("POST /v1/sms", func(w http.ResponseWriter, r *http.Request) {
    44			var req smsRequest
    45			s.send(w, r, channelSMS, &req, func() (Message, error) {
    46				if err := validateSMS(req); err != nil {
    47					return Message{}, err
    48				}
    49				return Message{Channel: channelSMS, To: req.To, Body: req.Body, Reference: req.Reference}, nil
    50			})
    51		})
    52		mux.HandleFunc("POST /v1/email", func(w http.ResponseWriter, r *http.Request) {
    53			var req emailRequest
    54			s.send(w, r, channelEmail, &req, func() (Message, error) {
    55				if err := validateEmail(req); err != nil {
    56					return Message{}, err
    57				}
    58				return Message{Channel: channelEmail, To: req.To, Subject: req.Subject, Body: req.Body, Reference: req.Reference}, nil
    59			})
    60		})
    61		mux.HandleFunc("GET /v1/messages", s.listMessages)
    62		mux.HandleFunc("DELETE /v1/mes
… [15758 more chars truncated]
```

</details>

<details><summary>Result</summary>

```
1	// Command mockprovider stands in for an external SMS and email provider.
     2	// See internal/mockprovider and docs/specs/03-mock-provider/goal.md.
     3	package main
     4	
     5	import (
     6		"context"
     7		"errors"
     8		"log/slog"
     9		"net/http"
    10		"os"
    11		"os/signal"
    12		"syscall"
    13		"time"
    14	
    15		"campaigncomposer/internal/config"
    16		"campaigncomposer/internal/mockprovider"
    17	)
    18	
    19	const shutdownTimeout = 10 * time.Second
    20	
    21	func main() {
    22		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    23		slog.SetDefault(logger)
    24	
    25		if err := run(); err != nil {
    26			slog.Error("mockprovider stopped", "err", err)
    27			os.Exit(1)
    28		}
    29	}
    30	
    31	func run() error {
    32		cfg, err := config.LoadMockProvider(os.Getenv)
    33		if err != nil {
    34			return err
    35		}
    36		addr := cfg.HTTPAddr
    37		slog.Info("mockprovider settings",
    38			"rate_per_sec", cfg.RatePerSec, "burst", cfg.Burst,
    39			"error_rate", cfg.ErrorRate, "slow_rate", cfg.SlowRate, "lost_rate", cfg.LostRate,
    40			"slow_delay", cfg.SlowDelay.String(), "latency_max", cfg.LatencyMax.String(),
    41			"random", cfg.Random)
    42	
    43		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    44		defer stop()
    45	
    46		// No WriteTimeout: slow responses wait for cfg.SlowDelay, and the server
    47		// must not end them itself.
    48		srv := &http.Server{
    49			Addr: addr,
    50			Handler: mockprovider.New(mockprovider.Options{
    51				RatePerSec: cfg.RatePerSec,
    52				Burst:      cfg.Burst,
    53				ErrorRate:  cfg.ErrorRate,
    54				SlowRate:   cfg.SlowRate,
    55				LostRate:   cfg.LostRate,
    56				SlowDelay:  cfg.SlowDelay,
    57				LatencyMax: cfg.LatencyMax,
    58				Random:     cfg.Random,
    59			}),
    60			ReadHeaderTimeout: 5 * time.Second,
    61		}
    62	
    63		errc := make(c
… [4081 more chars truncated]
```

</details>

**🔧 Bash** _List tests and x/time version_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend; ls internal/mockprovider; grep -n "^func Test\|t.Run(\"" internal/mockprovider/*_test.go internal/config/config_test.go | head -80; grep -n "x/time" go.mod
```

<details><summary>Result</summary>

```
failure.go
failure_test.go
limit_test.go
options.go
server.go
server_test.go
store.go
store_test.go
validate.go
internal/mockprovider/store_test.go:8:func TestStore_IDs(t *testing.T) {
internal/mockprovider/failure_test.go:30:func TestForced(t *testing.T) {
internal/mockprovider/failure_test.go:98:func TestForced_Lost(t *testing.T) {
internal/mockprovider/failure_test.go:113:func TestForced_ValidationFirst(t *testing.T) {
internal/mockprovider/failure_test.go:119:func TestSlow_StoresAfterClientLeaves(t *testing.T) {
internal/mockprovider/failure_test.go:150:func TestRandom_ZeroRates(t *testing.T) {
internal/mockprovider/failure_test.go:159:func TestRandom_Distribution(t *testing.T) {
internal/mockprovider/failure_test.go:185:func TestRandom_SameSeedSameModes(t *testing.T) {
internal/mockprovider/failure_test.go:195:func TestLatency(t *testing.T) {
internal/mockprovider/limit_test.go:17:func TestRateLimit_Burst(t *testing.T) {
internal/mockprovider/limit_test.go:47:func TestRateLimit_Steady(t *testing.T) {
internal/mockprovider/limit_test.go:58:func TestRateLimit_RetryAfterRoundsUp(t *testing.T) {
internal/mockprovider/limit_test.go:69:func TestRateLimit_PerChannel(t *testing.T) {
internal/mockprovider/limit_test.go:78:func TestRateLimit_429UsesNoToken(t *testing.T) {
internal/mockprovider/server_test.go:143:func TestSend_SMSAccepted(t *testing.T) {
internal/mockprovider/server_test.go:166:func TestSend_EmailAccepted(t *testing.T) {
internal/mockprovider/server_test.go:185:func TestSend_Invalid(t *testing.T) {
internal/mockprovider/server_test.go:215:func TestSend_RuneCount(t *testing.T) {
internal/mockprovider/server_test.go:222:func TestSend_BadJSON(t *testing.T) {
internal/mockprovider/server_test.go:240:func TestSend_RepeatDelivers(t *testing.T) {
internal/mockprovider/server_test.go:259:func TestSend_EmptyReferenceNotDuplicate(t *testing.T) {
internal/mockprovider/server_test.go:269:func TestSend_DuplicatePerChannel(t *testing.T) {
internal/mockprovider/server_t
… [1370 more chars truncated]
```

</details>

**🔧 Bash** _Read server tests_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend; cat -n internal/mockprovider/server_test.go | sed -n 1,142p; cat -n internal/mockprovider/server_test.go | sed -n 185,430p
```

**🔧 Bash** _Read failure and limit tests_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend; cat -n internal/mockprovider/failure_test.go internal/mockprovider/limit_test.go
```

<details><summary>Result</summary>

```
1	package mockprovider
     2	
     3	import (
     4		"context"
     5		"net/http"
     6		"strings"
     7		"sync"
     8		"testing"
     9		"time"
    10	)
    11	
    12	// sleepRecorder is a fake Sleep that records each duration and returns at once.
    13	type sleepRecorder struct {
    14		mu    sync.Mutex
    15		calls []time.Duration
    16	}
    17	
    18	func (s *sleepRecorder) Sleep(d time.Duration) {
    19		s.mu.Lock()
    20		defer s.mu.Unlock()
    21		s.calls = append(s.calls, d)
    22	}
    23	
    24	func (s *sleepRecorder) Calls() []time.Duration {
    25		s.mu.Lock()
    26		defer s.mu.Unlock()
    27		return append([]time.Duration(nil), s.calls...)
    28	}
    29	
    30	func TestForced(t *testing.T) {
    31		const slowDelay = 5 * time.Second
    32		cases := []struct {
    33			header        string
    34			status        int
    35			code          string
    36			wantDelivered int64
    37			check         func(t *testing.T, st ChannelStats, sleeps []time.Duration)
    38		}{
    39			{"500", 500, "provider_error", 0, func(t *testing.T, st ChannelStats, _ []time.Duration) {
    40				if st.Errors != 1 {
    41					t.Errorf("errors = %d, want 1", st.Errors)
    42				}
    43			}},
    44			{"503", 503, "unavailable", 0, func(t *testing.T, st ChannelStats, _ []time.Duration) {
    45				if st.Errors != 1 {
    46					t.Errorf("errors = %d, want 1", st.Errors)
    47				}
    48			}},
    49			{"slow", 202, "", 1, func(t *testing.T, st ChannelStats, sleeps []time.Duration) {
    50				if st.Slow != 1 {
    51					t.Errorf("slow = %d, want 1", st.Slow)
    52				}
    53				if len(sleeps) != 1 || sleeps[0] != slowDelay {
    54					t.Errorf("sleeps = %v, want [%v]", sleeps, slowDelay)
    55				}
    56			}},
    57			{"429", 429, "rate_limited", 0, func(t *testing.T, st ChannelStats, _ []time.Duration) {
    58				if st.RateLimited != 1 {
    59					t.Errorf("rate_limited = %d, want 1", st.RateLimited)
    60				}
    61			}},
    62			{"none
… [9779 more chars truncated]
```

</details>

<details><summary>Result</summary>

```
1	package mockprovider
     2	
     3	import (
     4		"encoding/json"
     5		"fmt"
     6		"io"
     7		"net/http"
     8		"net/http/httptest"
     9		"strings"
    10		"sync"
    11		"testing"
    12		"time"
    13	)
    14	
    15	// fakeClock is a clock that the test moves by hand.
    16	type fakeClock struct {
    17		mu sync.Mutex
    18		t  time.Time
    19	}
    20	
    21	func newFakeClock() *fakeClock {
    22		return &fakeClock{t: time.Date(2026, 11, 9, 10, 0, 0, 0, time.FixedZone("TRT", 3*3600))}
    23	}
    24	
    25	func (c *fakeClock) Now() time.Time {
    26		c.mu.Lock()
    27		defer c.mu.Unlock()
    28		return c.t
    29	}
    30	
    31	func (c *fakeClock) Advance(d time.Duration) {
    32		c.mu.Lock()
    33		defer c.mu.Unlock()
    34		c.t = c.t.Add(d)
    35	}
    36	
    37	// normalOptions has no random failures, no latency and a rate limit that
    38	// the tests do not reach.
    39	func normalOptions(clock *fakeClock) Options {
    40		return Options{
    41			RatePerSec: 1e6,
    42			Burst:      1e6,
    43			Random:     1,
    44			Now:        clock.Now,
    45			Sleep:      func(time.Duration) {},
    46		}
    47	}
    48	
    49	func startServer(t *testing.T, o Options) *httptest.Server {
    50		t.Helper()
    51		srv := httptest.NewServer(New(o))
    52		t.Cleanup(srv.Close)
    53		return srv
    54	}
    55	
    56	func post(t *testing.T, base, path, body string, header map[string]string) *http.Response {
    57		t.Helper()
    58		req, err := http.NewRequest(http.MethodPost, base+path, strings.NewReader(body))
    59		if err != nil {
    60			t.Fatal(err)
    61		}
    62		req.Header.Set("Content-Type", "application/json")
    63		for k, v := range header {
    64			req.Header.Set(k, v)
    65		}
    66		resp, err := http.DefaultClient.Do(req)
    67		if err != nil {
    68			t.Fatalf("POST %s: %v", path, err)
    69		}
    70		t.Cleanup(func() { resp.Body.Close() })
    71		return resp
    72	}
    73	
    74	func decode[
… [12660 more chars truncated]
```

</details>

**🔧 Bash** _Config tests, compose diff, STE check_
```bash
cd ~/Desktop/projects/case-studies/insider_one; sed -n 105,200p backend/internal/config/config_test.go; git diff docker-compose.yml .env.example; python3 .claude/skills/ste-writing/scripts/ste_check.py docs/architecture.md docs/decisions-log.md docs/plans/03-mock-provider.md docs/specs/03-mock-provider/*.md 2>&1 | tail -30
```

<details><summary>Result</summary>

```
func TestLoadMockProvider_Defaults(t *testing.T) {
	cfg, err := LoadMockProvider(env(map[string]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPAddr != ":8081" {
		t.Errorf("HTTPAddr = %q, want :8081", cfg.HTTPAddr)
	}
	if cfg.RatePerSec != 100 || cfg.Burst != 10 {
		t.Errorf("rate = %v/%d, want 100/10", cfg.RatePerSec, cfg.Burst)
	}
	if cfg.ErrorRate != 0.03 || cfg.SlowRate != 0.03 || cfg.LostRate != 0.01 {
		t.Errorf("rates = %v %v %v, want 0.03 0.03 0.01", cfg.ErrorRate, cfg.SlowRate, cfg.LostRate)
	}
	if cfg.SlowDelay != 5*time.Second || cfg.LatencyMax != 50*time.Millisecond {
		t.Errorf("durations = %v %v, want 5s 50ms", cfg.SlowDelay, cfg.LatencyMax)
	}
}

func TestLoadMockProvider_Values(t *testing.T) {
	cfg, err := LoadMockProvider(env(map[string]string{
		"HTTP_ADDR":         ":9000",
		"MOCK_RATE_PER_SEC": "20.5",
		"MOCK_RATE_BURST":   "3",
		"MOCK_ERROR_RATE":   "0.1",
		"MOCK_SLOW_RATE":    "0.2",
		"MOCK_LOST_RATE":    "0.3",
		"MOCK_SLOW_DELAY":   "2s",
		"MOCK_LATENCY_MAX":  "0",
		"MOCK_RANDOM":       "7",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := MockProviderConfig{
		HTTPAddr: ":9000", RatePerSec: 20.5, Burst: 3,
		ErrorRate: 0.1, SlowRate: 0.2, LostRate: 0.3,
		SlowDelay: 2 * time.Second, LatencyMax: 0, Random: 7,
	}
	if cfg != want {
		t.Errorf("cfg = %+v, want %+v", cfg, want)
	}
}

func TestLoadMockProvider_Errors(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"error rate over 1", map[string]string{"MOCK_ERROR_RATE": "1.5"}, "MOCK_ERROR_RATE"},
		{"slow rate negative", map[string]string{"MOCK_SLOW_RATE": "-0.1"}, "MOCK_SLOW_RATE"},
		{"lost rate not a number", map[string]string{"MOCK_LOST_RATE": "abc"}, "MOCK_LOST_RATE"},
		{"sum over 1", map[string]string{"MOCK_ERROR_RATE": "0.5", "MOCK_SLOW_RATE": "0.4", "MOCK_LOST_RATE": "0.2"}, "sum"},
		{"slow delay without unit", map[string]string{"MOCK_SLOW_DELAY": "5"}, "MOCK_SLOW_DELAY"
… [2171 more chars truncated]
```

</details>

**🔧 Bash** _Check env, compose api section, docs diff_
```bash
cd ~/Desktop/projects/case-studies/insider_one; cat .env.example | head -12; sed -n 1,80p docker-compose.yml | grep -n "env_file\|HTTP_ADDR\|api:\|environment" ; python3 .claude/skills/ste-writing/scripts/ste_check.py docs/architecture.md docs/decisions-log.md docs/plans/03-mock-provider.md docs/specs/03-mock-provider/goal.md docs/specs/03-mock-provider/requirements.md | wc -l; git diff docs/architecture.md docs/decisions-log.md | grep "^+" | head -60
```

<details><summary>Result</summary>

```
# Copy this file to .env. docker compose up also works without .env (mock LLM).

# Claude API key. Required only when LLM_MODE=anthropic. Never commit a real key.
ANTHROPIC_API_KEY=

# LLM backend: mock (no API key, used by all tests) or anthropic (real Claude API).
LLM_MODE=mock

# Seed (the seed service). Values below are the defaults. Uncomment to change them.
# SEED_USERS=5000
# SEED_RANDOM=42
# WARNING: SEED_RESET=true deletes and reloads all users on each start.
4:    environment:
31:  api:
37:    env_file:
40:    environment:
61:    env_file:
64:    environment:
79:    env_file:
       0
+++ b/docs/architecture.md
+- **mock-provider** — Go service that acts as an SMS and email provider, like Twilio or SendGrid. It sometimes returns `429`, `5xx`, slow responses and lost responses. It does no dedup (D-018).
+    mockprovider/   mock provider handlers, rate limit, failure modes, message store
+   - The mock provider delivers each request, also a repeat. The worker owns the dedup. It does not retry after a timeout or a dropped connection (D-018, D-019).
+++ b/docs/decisions-log.md
+
+## D-018 The mock provider does no dedup
+
+- **Context:** The case says that the same message must not go to the same person twice. Twilio Messaging and SendGrid Mail Send have no `Idempotency-Key` on their send endpoints. A repeat request sends the message again.
+- **Options:** The provider honors an `Idempotency-Key` and absorbs repeats. The provider delivers each repeat and the worker prevents duplicates.
+- **Choice:** The provider delivers each repeat (user decision). It stores an optional `reference`, as SendGrid `custom_args` does. `GET /v1/stats` counts `duplicate_deliveries` for the audit.
+- **Why:** The mock acts like the real providers. Thus the dedup logic is in our worker, where the case wants to see it. Provider dedup would hide a worker bug.
+
+## D-019 Only a timeout or a dropped connection is ambiguous, and the worker does not retry it
+
+- **Context:** With no pr
… [1815 more chars truncated]
```

</details>

**🔧 Bash** _Read compose services and HTTP_ADDR usage_
```bash
cd ~/Desktop/projects/case-studies/insider_one; sed -n 28,100p docker-compose.yml; grep -rn "HTTP_ADDR" backend --include=*.go | grep -v _test | head; grep -n "D-019\|passive\|is maintained" docs/decisions-log.md | head
```

<details><summary>Result</summary>

```
retries: 15
    # No ports, for the same reason as Postgres (host port 6379).

  api:
    build:
      context: .
      target: go-service
      args:
        CMD: api
    env_file:
      - path: .env
        required: false
    environment:
      DATABASE_URL: postgres://campaign:campaign@postgres:5432/campaign?sslmode=disable
      LLM_MODE: ${LLM_MODE:-mock}
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      # The api starts after the seed applies the migrations (D-015).
      seed:
        condition: service_completed_successfully

  # One-shot job. It applies the migrations, loads the users, then exits.
  # If users exist, it keeps them and exits 0. With SEED_RESET=true it
  # deletes all users and loads them again on each start.
  seed:
    build:
      context: .
      target: go-service
      args:
        CMD: seed
    env_file:
      - path: .env
        required: false
    environment:
      DATABASE_URL: postgres://campaign:campaign@postgres:5432/campaign?sslmode=disable
    restart: "no"
    depends_on:
      postgres:
        condition: service_healthy

  # Distroless images have no shell, so api and mock-provider have no
  # container healthcheck. `make up` checks them with curl.
  mock-provider:
    build:
      context: .
      target: go-service
      args:
        CMD: mockprovider
    env_file:
      - path: .env
        required: false
    # Failure injection and rate limit (docs/specs/03-mock-provider/goal.md).
    # Set the three rates to 0 to turn the random failures off.
    environment:
      MOCK_RATE_PER_SEC: ${MOCK_RATE_PER_SEC:-100}
      MOCK_RATE_BURST: ${MOCK_RATE_BURST:-10}
      MOCK_ERROR_RATE: ${MOCK_ERROR_RATE:-0.03}
      MOCK_SLOW_RATE: ${MOCK_SLOW_RATE:-0.03}
      MOCK_LOST_RATE: ${MOCK_LOST_RATE:-0.01}
      MOCK_SLOW_DELAY: ${MOCK_SLOW_DELAY:-5s}
      MOCK_LATENCY_MAX: ${MOCK_LATENCY_MAX:-50ms}
      # Empty means a seed from the clock.
      MOCK_RANDOM: ${MOCK_RANDOM:
… [417 more chars truncated]
```

</details>

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/.claude/agent-memory/reviewer/http-config-validation-defects.md`
```
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
```

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/.claude/agent-memory/reviewer/http-config-validation-defects.md (file state is current in your context — no need to Read it back)
```

</details>

**🔧 Bash** _Add memory index entry_
```bash
cd ~/Desktop/projects/case-studies/insider_one/.claude/agent-memory/reviewer; echo "- [HTTP and config validation defects](http-config-validation-defects.md) — NaN in float checks, Decoder trailing data, reset reuses IDs, env_file leaks" >> MEMORY.md
```

<details><summary>Result</summary>

```
(Bash completed with no output)
```

</details>

**🔧 SubagentHandback** ```json
{
  "message": "I found no Critical defects in slice 03-mock-provider. I found 2 Important and 6 Minor defects.\n\nI did not run the code. I found these by reading the code and the diff. I used your evidence (`make verify` and the race test both passed). The edge cases below are from Go's documented behaviour. I did not try them against the server.\n\nThe five Review Focus items work as the spec says:\n1. A slow wait stores the message after the client times out. `server.go:146-155` does not read `r.Context()`. `TestSlow_StoresAfterClientLeaves` tests this. See Minor 5 about that test.\n2. Many goroutines with the same reference give 499 duplicates out of 500. `store.add` does this under one lock. `TestConcurrent_DuplicateCount` tests it.\n3. A wrong JSON type or a body that is not JSON gets 400 `invalid_request`. `TestSend_BadJSON` tests it.\n4. An empty reference does not count as a duplicate (`store.go:80`, `TestSend_EmptyReferenceNotDuplicate`).\n5. `limit=0`, `limit=1001`, `offset=-1` and `channel=push` all get 400 (`TestMessages_BadQuery`).\n\nThe order of checks matches goal.md. The counters are correct in each path. Retry-After uses `TokensAt` correctly. A failed `AllowN` uses no token in x/time/rate v0.16. A POST with no `Idempotency-Key` is not replayable, so the Go client does not retry it by itself.\n\n## Critical\nNone.\n\n## Important\n1. `backend/internal/config/config.go` (the rate loop and `MOCK_RATE_PER_SEC`) — `strconv.ParseFloat` accepts `NaN` and `Inf`.\n   - Failure: `MOCK_ERROR_RATE=NaN` passes the check `f < 0 || f > 1`, and the sum check also passes. The process starts, and `x < NaN` is always false, so it never returns an error. FR-16 says the process must not start.\n   - Fix: reject `math.IsNaN(f) || math.IsInf(f, 0)`. Add the test cases `NaN` and `Inf`.\n2. `backend/internal/mockprovider/store.go:59` — `resetLocked` sets `nextID` to 0.\n   - Failure: after `DELETE /v1/messages`, the next SMS gets `sms_00000001` again. The spec says an `id` is unique for each delivery. The demo resets the provider but the deliveries table keeps its rows. Thus slice 07 can find two rows with the same provider id.\n   - Fix: do not reset `nextID`. Or add a prefix for each reset.\n\n## Minor\n1. `server.go:110-111` — `Decode` reads only one JSON value.\n   - Failure: `{\"to\":\"+905321234567\",\"body\":\"hi\"} junk` gets 202. A small object followed by 2 MiB of spaces gets 202, but the spec says a body over 1 MiB gets 400.\n   - Fix: after `Decode`, require `dec.Decode(&struct{}{}) == io.EOF`.\n2. `docker-compose.yml:79-81` — the `env_file: .env` line for mock-provider is not necessary.\n   - Failure: the container gets `ANTHROPIC_API_KEY`. If `.env` sets `HTTP_ADDR`, the provider listens on the wrong port and the `8081:8081` mapping fails.\n   - Fix: remove `env_file`. The `${...}` values in compose already read `.env`.\n3. `cmd/mockprovider/main.go:19` — `shutdownTimeout` is 10 s, and Docker's default stop grace period is also 10 s.\n  
… [2097 more chars truncated]
```

<details><summary>Result</summary>

```
{"success":true,"message":"Report delivered to your caller."}
```

</details>

