# Requirements: Segment rules

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

- **FR-1.** `POST /api/segments/preview` with the example brief and the default mock returns `200`. The response has the rule, the description and an audience with `email` and `sms` of 1 or more.
- **FR-2.** Sanitization removes control characters except `\n` and `\t`, format characters and `<brief>` tags, with case ignored. It normalizes to NFC and trims the text. A table test covers each step.
- **FR-3.** A brief with no letter, or with more than 1,000 runes after sanitization, returns `400` with `code: invalid_brief`. The LLM client gets no call. A brief of 1,000 Turkish runes, for example `ş`, is valid.
- **FR-4.** A body over 16 KiB, a body that is not valid JSON, or a body with no `brief` returns `400` with `code: invalid_request`.
- **FR-5.** The validator rejects each case in "Validation". A table test has one or more cases for each item. The cases include these:
  - an unknown field, an unknown operator, and a wrong operator for a field
  - a string `"30"` for `within_days`, and a city that is not on the list
  - `between` with low over high, and an `in` list of 21 values
  - depth 4, and 21 conditions
  - an unknown key, and data after the object
- **FR-6.** The validator returns all errors with a path, not only the first.
- **FR-7.** SQL injection in a value fails validation. Examples: `"Istanbul' OR 1=1 --"` for `city`, and `"city; DROP TABLE users"` as a field name.
- **FR-8.** For each field and operator, the compiler gives the expected SQL text and parameters. No value from the rule is in the SQL text.
- **FR-9.** With a fixed clock, `within_days`, `older_than_days` and `age` give the expected time and date parameters.
- **FR-10.** An empty root group compiles to `TRUE` and counts all users. An empty nested group fails validation.
- **FR-11.** If the first LLM answer fails validation and the second is valid, the response is `200`. The second request to the mock contains the first answer as a tool call. It also contains the validation errors as a tool result with the error flag.
- **FR-12.** If both LLM answers fail validation, the response is `422` with `code: segment_invalid` and the errors in `details`. The backend runs no count query.
- **FR-13.** An `unsupported` answer returns `422` with `code: segment_unsupported` and the reason. The backend runs no count query.
- **FR-14.** An `ok` answer with `unmapped` returns `200` with the `unmapped` list. More than 5 items, or an item over 200 runes, fails validation.
- **FR-15.** Against Postgres, the audience counts are equal to the counts that a test calculates in Go from the seeded users. The test covers the example rule, an `or` group, an `age between` rule and an empty root.
- **FR-16.** `audience.languages` has one entry for each language in the audience, sorted by `reachable` from high to low, then by `code`. A language with `reachable` of 0 is not in the list. A rule with `country in [TR, US]` gives two entries. The example rule gives only `tr`.
- **FR-17.** A rule that selects no users returns `200` with all counts at 0 and an empty `languages` list.
- **FR-18.** The same rule always gives the same description. Each field and each operator has a test.
- **FR-19.** An LLM client error returns `502` with `code: llm_error`. A timeout returns `504` with `code: llm_timeout`. A refusal returns `502` with `code: llm_refused` and no repair round. The test uses the mock and an injected timeout, with no real wait.
- **FR-20.** The system prompt contains each allowlist field, each operator and each city. A test builds the prompt and checks this.
- **FR-21.** The user message puts the sanitized brief between `<brief>` and `</brief>`. The system prompt does not contain the brief.
- **FR-22.** On `/campaigns/new`, the operator enters a brief and clicks "Find audience". The page shows "Generating segment…" while the request runs, then the description, the counts, the languages and the `unmapped` notes.
- **FR-23.** On an API error, the page shows the API message and keeps the brief. The page shows `unmapped` and `reason` as text, not as HTML.
- **FR-24.** For each user in the result of the recipients function for `sms`, `sms_consent` is true. The same is true for `email`. The test uses a rule that matches the seed users with no consent, and they are not in either result.
- **FR-25.** `no_consent` is equal to `matched` minus `reachable`, in the top level and in each language entry.
- **FR-26.** A first answer with no `submit_segment` call gets a repair round. The repair request ends with a user text message, not a tool result.

## Non-functional requirements

- **NFR-1.** The compiler accepts only the validated rule type. No function in `internal/segment` takes raw JSON and returns SQL.
- **NFR-2.** The allowlist is in one table. The validator, the compiler, the description, the prompt and the tool schema read it. A test adds a fake field to a copy of the table and checks that all five use it.
- **NFR-3.** The count query has a 2 second statement timeout. A timeout returns `503` with `code: db_unavailable`.
- **NFR-4.** Tests use the mock LLM. No test reads `ANTHROPIC_API_KEY`. With `LLM_MODE=mock`, the api starts with no key.
- **NFR-5.** Logs contain no brief text and no API key.
- **NFR-6.** `internal/segment` imports nothing from `internal/`. `internal/llm` imports nothing from `internal/` except `config`.
- **NFR-7.** The Postgres tests read `TEST_DATABASE_URL`, as in slice 02. Without it, `go test` skips them.
- **NFR-8.** Frontend tests mock the API. `pnpm test:run`, `pnpm build` and `CI=true pnpm lint` pass.
- **NFR-9.** `make up` with no `.env` starts the stack. The example brief on `/campaigns/new` shows an audience.
- **NFR-10.** `make verify` passes.
- **NFR-11.** The recipients function is the only function in `internal/segment` that returns user rows. Its channel argument is a type with two values. No argument turns off the consent condition.
