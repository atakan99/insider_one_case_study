# Requirements: Content generation

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

- **FR-1.** `POST /api/content/generate` with the example brief, the example rule and the default mock returns `200` with `text/event-stream`. The events are `audience`, then one or more `draft` events and one `final` event for `tr`, then `done`.
- **FR-2.** These bodies return `400` with `code: invalid_request`: a body over 16 KiB, a body that is not valid JSON, and a body with no `brief` or no `rule`. The LLM client gets no call.
- **FR-3.** A brief that fails the slice 04 checks returns `400` with `code: invalid_brief`. The LLM client gets no call.
- **FR-4.** A rule that fails the slice 04 validator returns `400` with `code: invalid_rule` and the errors in `details`. The cases include an unknown field and `"Istanbul' OR 1=1 --"` for `city`. The backend runs no query.
- **FR-5.** A rule with no reachable user returns `422` with `code: empty_audience`. The LLM client gets no call.
- **FR-6.** The backend makes one LLM call for each language in the audience. A rule with `country` in `[TR, US]` gives calls for `tr` and `en`. Each request names its language in the user message.
- **FR-7.** The variable parser accepts zero, one or more `{{first_name}}`. It rejects `{{last_name}}`, `{{ first_name }}`, `{{first_name}`, a single `{{` and a single `}}`. A table test covers each case.
- **FR-8.** The renderer replaces each `{{first_name}}` with the given name. Text with no variable stays the same. It does no HTML escaping.
- **FR-9.** The validator counts runes after rendering with the longest first name. The test uses a longest name of 20 runes. An SMS of 153 runes with one `{{first_name}}` (13 runes) renders to 160 runes and passes. An SMS of 154 runes with one variable renders to 161 and fails. The subject passes at 80 rendered runes and fails at 81. Turkish characters, for example `ş`, count as 1.
- **FR-10.** The validator rejects each case in "Validation". A table test has one or more cases for each row. The cases include these:
  - an unknown key, a missing field, and an empty field after trim
  - a `\n` in the subject, and a control character
  - an HTML tag in each field, and data after the object It returns all errors with a path.
- **FR-11.** A limit error gives the rendered length and the limit.
- **FR-12.** The audience query gives the longest first name for each language. It counts only users with consent on one or more channels. A Postgres test compares it with a value that the test calculates in Go from the seed users.
- **FR-13.** If the first answer fails validation and the second passes, the stream has `draft`, `retry`, `draft` and `final` for that language. The second mock request contains the first answer as a tool call and the errors as a tool result with the error flag.
- **FR-14.** If both answers fail, the language gets an `error` event with `code: content_invalid` and the errors in `details`.
- **FR-15.** A first answer with no `submit_content` call gets a repair round. The repair request ends with a user text message, not a tool result.
- **FR-16.** An answer with `stop_reason: max_tokens` counts as a validation failure and gets a repair round. If its tool input is not a valid JSON object, the replayed assistant turn has `{}` as the input.
- **FR-17.** An LLM client error gives an `error` event with `code: llm_error`. A timeout gives `code: llm_timeout`. A refusal gives `code: llm_refused` and no repair round. The test uses the mock and an injected timeout, with no real wait.
- **FR-18.** With two languages, an error in one language does not stop the other. The stream has a `final` event for the other language, then `done`.
- **FR-19.** The partial JSON parser gives the correct partial content for each prefix of a valid answer. The test cuts the answer at each byte, including inside a `\u` escape and inside a multi-byte rune. The parser never returns an error or a field that is longer than the final field.
- **FR-20.** The backend sends at most one `draft` event per 100 ms for each language, with an injected clock. The last `draft` before `final` has the full text.
- **FR-21.** A client disconnect cancels the LLM calls of the request. The test checks that the mock sees a cancelled context.
- **FR-22.** The system prompt contains the SMS limit, the subject limit, the variable rule and the plain text rule. It does not contain the brief. It is the same for each language and each request.
- **FR-23.** The user message contains the language, the rule description, the longest first name length and the sanitized brief between `<brief>` and `</brief>`.
- **FR-24.** The default mock content for `tr` and `en` passes validation with a longest first name of 30 runes.
- **FR-25.** On `/campaigns/new`, a segment result with one or more languages starts content generation with no second click. The page shows "Generating content…" while the request runs.
- **FR-26.** The page shows one card for each language. The `draft` text shows as a draft. On `final`, the card shows the final text and a counter for SMS and subject. On `retry`, the card shows "Fixing content…". On `error`, the card shows the message, and the other cards stay.
- **FR-27.** "Regenerate content" sends the same brief and rule again. A new "Find audience" click aborts the open content request and clears the cards.
- **FR-28.** The page shows the content as text, not as HTML. A test with `<b>x</b>` in a `draft` event shows the tags as text.
- **FR-29.** An error before the stream opens shows its message in the content area. The segment result stays.
- **FR-30.** A `fallback` block in the stream gives a `retry` event with no errors. The backend uses only the `submit_content` call after it. The language still has its one repair round.

## Non-functional requirements

- **NFR-1.** Tests use the mock LLM. No test reads `ANTHROPIC_API_KEY`. With `LLM_MODE=mock`, the api starts with no key.
- **NFR-2.** Logs contain no brief text, no content text and no API key.
- **NFR-3.** `internal/content` imports nothing from `internal/`. `internal/llm` imports nothing from `internal/` except `config`.
- **NFR-4.** The variable parser and the renderer are exported from `internal/content`. Slice 07 can use them with no change.
- **NFR-5.** No user row leaves the database in this slice. The content step uses only aggregate values from the audience query.
- **NFR-6.** The Postgres tests read `TEST_DATABASE_URL`, as in slice 02. Without it, `go test` skips them.
- **NFR-7.** Frontend tests mock the stream. `pnpm test:run`, `pnpm build` and `CI=true pnpm lint` pass.
- **NFR-8.** `make up` with no `.env` starts the stack. The example brief on `/campaigns/new` shows the audience, then a `tr` card that fills in and becomes final.
- **NFR-9.** The SSE response works with the api server settings. No server write timeout cuts a stream of 90 seconds.
- **NFR-10.** `make verify` passes.
