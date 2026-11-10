# Requirements: Review and approve

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

### Draft creation

- **FR-1.** `POST /api/content/generate` with the example brief, the example rule and the default mock ends with a `campaign` event, then `done`. The database has one campaign with `status = draft` and one `campaign_contents` row for `tr`.
- **FR-2.** The draft has the sanitized brief, the rule, the description, the final content, the counts and `first_name_runes`. The counts match `CountAudience` for the same rule.
- **FR-3.** With two languages, one final and one failed, the draft has one content row. `skipped_languages` has the failed language.
- **FR-4.** If all languages fail, the stream has no `campaign` event and the database has no new campaign.
- **FR-5.** If the draft save fails, the stream has an `error` event with no `language` and `code: draft_not_saved`, then `done`.
- **FR-6.** After a client disconnect, the backend saves no draft. The test checks that the database has no new campaign.
- **FR-7.** A request with the `campaign_id` of a draft and the same brief and rule replaces the content of that draft. The database still has one campaign.
- **FR-8.** A request with the `campaign_id` of an approved campaign, or with a different brief or rule, gives an `error` event with no `language` and `code: campaign_conflict`. The campaign does not change.

### Save and approve

- **FR-9.** `PUT /api/campaigns/{id}/contents` with valid edits returns `200`. `GET` then returns the edits, trimmed.
- **FR-10.** `POST /api/campaigns/{id}/approve` with valid content returns `200`. The campaign has `status = approved`, `approved_at` set, and the content of the request.
- **FR-11.** A second approval returns `200` and the same campaign. It changes nothing, also with a different body.
- **FR-12.** Ten approvals of one draft at the same time give one approval. `approved_at` has one value. A Postgres test covers this.
- **FR-13.** A save on an approved campaign returns `409 campaign_not_draft`. A direct SQL update of the content of an approved campaign fails, or the store blocks it. A Postgres test covers the chosen way.
- **FR-14.** These bodies return `400 invalid_request` for save and approve: a body over 64 KiB, a body that is not valid JSON, and a body with no `contents`. An id of `abc`, `0` or `-1` also returns `400`. An unknown id returns `404 not_found`.
- **FR-15.** These `contents` return `422 content_invalid` with the errors and their paths in `details`. A table test has one case for each:
  - a missing language of the draft
  - a language that is not in the draft
  - the same language two times
  - an unknown key in an entry
  - an SMS of 161 runes after rendering with the `first_name_runes` of that language
  - a subject of 81 runes after rendering, and a subject with `\n`
  - `{{last_name}}` in a field, and an HTML tag in a field
- **FR-16.** An error path has the index, for example `contents[1].sms`. With two bad entries, `details` has the errors of both.
- **FR-17.** Each save and each approval runs the audience query again and updates the counts and `first_name_runes`. A test adds a user with a longer first name between the draft and the approval. An SMS that fit before then fails.
- **FR-18.** If a language of the draft has no reachable users at approval, the approval returns `422 content_invalid` with the path `contents[i].language`. The campaign stays a draft.
- **FR-19.** If a query fails or takes more than 2 seconds, save and approve return `503 db_unavailable` and change nothing.

### Read

- **FR-20.** `GET /api/campaigns/{id}` returns the campaign in the shape of `goal.md`.
- **FR-21.** `GET /api/campaigns` returns the campaigns, the newest first, at most 100. The entries have no content text.
- **FR-22.** `GET /api/campaigns/{id}/sample` returns up to 5 distinct first names for each language of the campaign. The first name is the longest. A Postgres test compares the result with a value that the test calculates in Go from the seed users.
- **FR-23.** The sample result is the same for two calls.
- **FR-24.** The sample query reads no column except `first_name`, `language`, `id`, the consent columns and the columns of the rule. A test checks the SQL text.

### Frontend

- **FR-25.** On `/campaigns/new`, "Review & approve" is off while the content stream runs. It comes on after a `campaign` event. It stays off when the stream has no `campaign` event.
- **FR-26.** "Regenerate content" sends the `campaign_id` of the current draft. A new "Find audience" sends no `campaign_id`.
- **FR-27.** An `error` event with no `language` shows in the content area. The cards stay.
- **FR-28.** The review window has the channel tabs "SMS" and "Email", and one language tab for each language of the draft.
- **FR-29.** The SMS tab shows the SMS editor, the counter and the SMS preview. The Email tab shows the subject and body editors, their counters and the email preview.
- **FR-30.** The counter renders with a name of `first_name_runes` runes and counts code points. A test with `first_name_runes: 14` and an SMS of 146 runes plus one `{{first_name}}` shows `160 / 160`. One more rune shows `161 / 160` in red, and "Save draft" and "Approve & send" are off.
- **FR-31.** The counter is orange above 90% of the limit. It uses the `--warning` token.
- **FR-32.** A field with `{{last_name}}`, an empty field or a subject with a line break shows an error line. "Save draft" and "Approve & send" are off. The language tab shows an error mark.
- **FR-33.** The preview renders `{{first_name}}` with the selected sample name. The first sample is the longest name, with "(longest)" in its label. The preview shows `<b>x</b>` as text, not as HTML.
- **FR-34.** If the sample request fails, the preview uses "Alex" and shows a note. "Approve & send" still works.
- **FR-35.** The footer and the confirm popup name each skipped language. The confirm totals add only the languages of the draft.
- **FR-36.** "Save draft" sends one `PUT` with all languages and shows "Draft saved." on success.
- **FR-37.** "Cancel" with unsaved edits asks "Discard your changes?". With no edits, it closes at once.
- **FR-38.** "Approve & send" opens the confirm popup. "Back" closes it and keeps the edits. "Send now" sends one approval with the edited content of all languages.
- **FR-39.** While the approval runs, "Send now" and "Back" are off. A double click sends one request.
- **FR-40.** On `200`, the page goes to `/campaigns/{id}`.
- **FR-41.** On `422 content_invalid`, each error shows under its field, on the correct language tab. An error with no field shows in a banner.
- **FR-42.** For an approved campaign, the review window shows the content read-only, with no "Save draft" and no "Approve & send".
- **FR-43.** `/campaigns` lists the campaigns with the number, the description, the languages, the status badge and the time. A click on a draft opens the review window. A click on an approved campaign goes to `/campaigns/{id}`. With no campaigns, the page shows "No campaigns yet".
- **FR-44.** `/campaigns/{id}` shows the campaign number, the status badge, the description and one read-only card for each language with its counts. A draft also shows "Review & approve". A `404` shows "Campaign not found" and a link to `/campaigns`.

## Non-functional requirements

- **NFR-1.** Tests use the mock LLM. No test reads `ANTHROPIC_API_KEY`.
- **NFR-2.** Logs contain no brief text, no content text and no API key.
- **NFR-3.** Save and approve use `content.Validate` and `content.Render` from slice 05. The browser counter has a unit test with the same cases as FR-9 of slice 05.
- **NFR-4.** `internal/campaign` imports nothing from `internal/` except `content`, `segment` and `store`.
- **NFR-5.** The campaign tables store no user row. Only the sample endpoint sends first names to the browser. No endpoint sends a surname, an email address or a phone number.
- **NFR-6.** The review window and the list have no hard-coded colour. All colours come from the theme tokens. They work in the light theme and in the dark theme.
- **NFR-7.** At 360 px width, the review window fills the screen, and `/campaigns` and the window have no horizontal scroll. A headless Chrome check at 360 px and at 1280 px shows this.
- **NFR-8.** The tabs, the dialog and the confirm popups work with the keyboard. The focus stays in the open dialog.
- **NFR-9.** The Postgres tests read `TEST_DATABASE_URL`, as in slice 02. Without it, `go test` skips them.
- **NFR-10.** Frontend tests mock the API. `pnpm test:run`, `pnpm build` and `CI=true pnpm lint` pass.
- **NFR-11.** `make up` with no `.env` starts the stack. The example brief on `/campaigns/new` makes a draft. The draft shows on `/campaigns`. A click opens the review window. "Approve & send", then "Send now", opens `/campaigns/{id}` with the status "Approved".
- **NFR-12.** `make verify` passes.
