# Goal: Review and approve

## Why

The case flow has five steps. Step 3 says that the marketer previews the content, edits it if necessary, and approves it. Claude suggests, and the human makes the last decision.

This slice builds step 3. When Claude finishes the content, the backend saves the campaign as a draft. The marketer opens the draft in a review window, checks each message with a sample user, edits the text, and approves. A second confirm popup comes before the approval. Slice 07 sends the approved campaign.

## Input

- The case brief, section "Kullanıcı akışı" step 3, and the sections "Kitle" and "İçerik".
- Slice 04: the brief sanitization, the rule validator, the audience query. See `docs/specs/04-segment-rules/goal.md`.
- Slice 05: the content endpoint, the content types, the variable parser, the renderer, the validator, and the content cards on `/campaigns/new`. See `docs/specs/05-content-generation/goal.md`.
- The mockup that the user accepted on 2026-11-10. A copy is in `docs/specs/06-review-approve/review-mockup.html`.
- User decisions of 2026-11-10:
  - The review is one window. The marketer reviews, edits and approves in it.
  - SMS and Email are tabs. The Email tab shows the subject and the body.
  - One button "Approve & send" approves the campaign. It opens a second confirm popup first.
  - The design is clean, uses design tokens, works on a phone and is easy to use.
  - If the content of a language failed in slice 05, the marketer can still approve. The campaign skips that language.
  - A later slice can add a retry for one failed language. This slice does not.
  - Campaigns have drafts. When Claude finishes the content, the backend creates a draft.
  - A click on a draft opens its review window. The marketer approves the draft there, and then the campaign goes out.

## Output

1. A migration that adds the campaign tables.
2. A Go package `internal/campaign`. It creates, reads, updates and approves campaigns.
3. A change to `POST /api/content/generate`: at the end of the stream, the backend saves a draft and sends a `campaign` event.
4. New endpoints in `internal/httpapi`:
   - `GET /api/campaigns` lists the campaigns.
   - `GET /api/campaigns/{id}` returns one campaign.
   - `PUT /api/campaigns/{id}/contents` saves the edits of a draft.
   - `POST /api/campaigns/{id}/approve` saves the last edits and approves the draft.
   - `GET /api/campaigns/{id}/sample` gives sample first names for the preview.
5. A review window component. `/campaigns/new` and `/campaigns` both open it.
6. On `/campaigns/new`: a "Review & approve" button.
7. On `/campaigns`: the campaign list. A click on a draft opens the review window.
8. On `/campaigns/{id}`: a read-only view of an approved campaign.

This slice does not send messages. Slice 07 reads the approved campaigns and sends them. Slice 08 builds the live dashboard on `/campaigns/{id}`.

## Flow

### New campaign

1. The marketer writes the brief and clicks "Find audience". Slices 04 and 05 run as before.
2. When all languages are final or failed, the backend saves a draft. The stream sends a `campaign` event with the draft id, then `done`.
3. The "Review & approve" button comes on.
4. The marketer clicks "Review & approve". The review window opens with the draft.
5. The marketer checks and edits the content, then clicks "Approve & send". The confirm popup opens.
6. The marketer clicks "Send now". The page sends the approval request.
7. On success, the page goes to `/campaigns/{id}`. The page shows the campaign with the status "Approved".

### Existing draft

1. The marketer opens `/campaigns`. The list shows each campaign with its status.
2. The marketer clicks a draft. The review window opens with that draft.
3. Steps 5 to 7 of "New campaign" follow.

The marketer can close the window at any step. The draft stays in the list.

### Regenerate

- "Regenerate content" on `/campaigns/new` sends the draft id with the content request. The backend replaces the content of that draft. It does not create a second draft.
- A new "Find audience" click starts a new campaign. The old draft stays in the list.

## Draft creation

The content endpoint of slice 05 gets these changes:

- The request has an optional `campaign_id`. The page sends it for "Regenerate content".
- When all languages have a `final` or an `error` event, and one or more languages are final, the backend saves the draft. It then sends the `campaign` event and the `done` event.
- The draft has the sanitized brief, the validated rule, the description, the final content of each language, the counts and the `first_name_runes` of each language. It also lists the failed languages as skipped.
- With a `campaign_id`, the backend replaces the content, the counts and the skipped languages of that draft in one transaction. The brief and the rule must be the same as in the draft. If not, or if the campaign is not a draft, the backend sends an `error` event with no language and `code: campaign_conflict`. The stream does not open a new draft in this case.
- If all languages fail, the backend saves no draft. It sends no `campaign` event.
- If the save fails, the backend sends an `error` event with no language and `code: draft_not_saved`. The cards keep their content. "Review & approve" stays off.
- After a client disconnect, the backend saves no draft. The marketer did not see the result.

New event:

| Event | Data | When |
|---|---|---|
| `campaign` | `{"id": 12, "status": "draft"}` | After the backend saves the draft, before `done`. |

The `error` event can now have no `language` key. The frontend shows such an error in the content area.

## Review window

The review window follows the mockup. It uses the shadcn-vue components of the app: `Dialog`, `Tabs`, `Textarea`, `Input`, `Select`, `Button` and `AlertDialog`.

When it opens, the window loads the campaign with `GET /api/campaigns/{id}` and the sample names with `GET /api/campaigns/{id}/sample`. If the campaign is not a draft, the window shows the content read-only, with no edit and no approve.

### Layout

- The header has the title "Review campaign" and the text "Check and edit the messages. Nothing is sent until you confirm."
- A toolbar has two tab groups:
  - Channel tabs: "SMS" and "Email". Each tab shows the recipient count of the selected language for that channel.
  - Language tabs: one tab for each language of the draft. The tab label is the language name.
- The body has two columns: the editor on the left and the preview on the right.
- The footer has a summary line and the buttons "Cancel", "Save draft" and "Approve & send".
- The footer lists each skipped language, for example "English skipped: content failed."

### SMS tab

- The editor has one text area for the SMS.
- A counter shows the rendered length and the limit, for example `121 / 160`.
- The preview shows the SMS in a message bubble, with the sample first name filled in.

### Email tab

- The editor has one input for the subject and one text area for the body.
- The subject has a counter with the limit 80. The body has a counter with the limit 2,000.
- The preview shows an email with the sender, the subject and the body, with the sample first name filled in.

### Counters and checks in the browser

- The counter renders the text with a name of `first_name_runes` runes, from the campaign. It then counts the code points. Thus it gives the same number as the backend.
- The counter is orange above 90% of the limit and red above the limit.
- Below each field, a short line says why the field fails, for example "Too long after the name is filled in."
- The browser also checks the variable rule, an empty field and a line break in the subject.
- If one field of one language fails, "Save draft" and "Approve & send" are off. A mark on the language tab shows which language has the error.
- These checks help the marketer only. The backend makes the decision.

### Preview as

- A select above the preview lists the sample first names of the selected language.
- The first entry is the longest first name. Its label has "(longest)" after it.
- The preview shows the text as plain text, not as HTML.
- If the sample request fails, the preview uses the name "Alex" and shows a short note. The approval still works.
- The select shows first names only. The mockup shows surnames and a "To:" address. The build leaves them out, because only first names reach the browser.
- The email preview shows the sender as "Your brand". Slice 07 owns the real sender. This slice adds no sender setting.

### Save and close

- "Save draft" sends `PUT /api/campaigns/{id}/contents` with the content of all languages. On success, a toast says "Draft saved."
- "Cancel" or the close button closes the window. If the window has unsaved edits, a confirm popup asks "Discard your changes?" first.
- `{{first_name}}` shows as written in the editor.

### Confirm popup

- "Approve & send" opens an `AlertDialog` with the title "Send this campaign?".
- The text says: "Messages go out now. Only users who gave consent for a channel get that channel. You cannot undo this."
- A summary shows the SMS total, the email total and the languages. The totals include only the languages of the draft.
- If the campaign skips a language, the summary says so.
- "Back" closes the popup and keeps the review window open.
- "Send now" sends `POST /api/campaigns/{id}/approve` with the content of all languages. While the request runs, both buttons are off and "Send now" shows a spinner.
- On an error, the popup closes and the review window shows the message. For `content_invalid`, each error shows under its field.
- An error with no field, for example a language with no reachable users left, shows in a banner at the top of the window.

### Mobile

- Below 720 px width, the window fills the screen.
- The editor is above the preview.
- The tab groups and the footer buttons use the full width.
- The page has no horizontal scroll at 360 px width.

### Design tokens

- All colours, radii and spaces come from the theme tokens in `src/assets/index.css`. No component has a hard-coded colour.
- The slice adds one token `--warning` for the orange counter, in the light theme and in the dark theme.
- The window works in the light theme and in the dark theme.

## Campaign list

On `/campaigns`:

- The page calls `GET /api/campaigns`. The newest campaign is first.
- Each row shows the campaign number, the rule description, the languages, a status badge ("Draft" or "Approved") and the creation time.
- A click on a draft opens the review window on the list page.
- A click on an approved campaign goes to `/campaigns/{id}`.
- With no campaigns, the page shows "No campaigns yet" and a link to `/campaigns/new`.
- On a phone, the rows become stacked cards. The page has no horizontal scroll at 360 px width.
- The sidebar has a "Campaigns" link to this page.

## Campaign page

On `/campaigns/{id}`:

- The page calls `GET /api/campaigns/{id}`.
- The header shows "Campaign #12", a status badge and the rule description.
- The page shows one card for each language, with the SMS, the email subject and the email body as read-only text.
- Each card shows the SMS and email counts at approval.
- A short line says that sending starts in a later step. Slice 07 replaces this line.
- For a draft, the page shows a "Review & approve" button that opens the review window.
- A `404` shows "Campaign not found" and a link to `/campaigns`.

## Data

The migration adds two tables.

`campaigns`:

| Column | Type | Note |
|---|---|---|
| `id` | `bigint` identity | Primary key. |
| `brief` | `text` | The sanitized brief. |
| `rule` | `jsonb` | The validated rule. |
| `description` | `text` | The rule description from slice 04. |
| `status` | `text` | Not null. This slice writes `draft` and `approved`. Slice 07 adds more values. |
| `skipped_languages` | `text[]` | The languages with failed content. |
| `created_at` | `timestamptz` | Not null. |
| `updated_at` | `timestamptz` | Not null. |
| `approved_at` | `timestamptz` | Null for a draft. |

`campaign_contents`:

| Column | Type | Note |
|---|---|---|
| `campaign_id` | `bigint` | References `campaigns`. |
| `language` | `text` | `tr` or `en`. Primary key with `campaign_id`. |
| `sms` | `text` | Trimmed. |
| `email_subject` | `text` | Trimmed. |
| `email_body` | `text` | Trimmed. |
| `sms_count` | `bigint` | The SMS recipient count at the last save. |
| `email_count` | `bigint` | The email recipient count at the last save. |
| `first_name_runes` | `integer` | The longest first name at the last save. |

- The tables store no user row and no recipient list. Slice 07 gets the recipients with `segment.Recipients` at send time.
- The counts are a snapshot for the view. Slice 07 decides which recipients get a message.
- Only a draft can change. After approval, the content and the status of this slice cannot change. A Postgres test proves this. The plan decides between a trigger and a check in the store code.

## Validation

The backend checks each save and each approval with these rules. It returns all errors with a path, as in slice 05.

| Item | Rule |
|---|---|
| body | Valid JSON, 64 KiB or less, with `contents`. |
| campaign | Exists and has the status `draft`. |
| `contents` | Exactly one entry for each language of the draft. No other language. |
| `contents[i]` | Exactly the keys `language`, `sms`, `email_subject` and `email_body`. |
| text fields | Pass the slice 05 validator with the `first_name_runes` of that language. |

- The paths have the index, for example `contents[1].sms`.
- On each save and each approval, the backend validates the stored rule again and runs the audience query again. It updates the counts and `first_name_runes`. It then checks the content with the new `first_name_runes`.
- The audience can change between generation and approval. The check uses the audience at that time.
- If a language of the draft has no reachable users now, the backend returns `content_invalid` with the path `contents[i].language`. It does not skip the language silently.
- The marketer cannot add or remove a language in the review window.

## Double submit

The case says that a person must not get the same message two times. A double click or a repeated request must not approve a campaign two times.

- The approval locks the campaign row, checks the status, writes the content and sets `approved`, all in one transaction.
- If the campaign is already approved, the backend returns `200` with the campaign and changes nothing. It does not check the body.
- Two approvals of one draft at the same time give one approval. A Postgres test covers this.
- A save on an approved campaign returns `409 campaign_not_draft`.

## API

All errors use the slice 04 error shape.

### `GET /api/campaigns`

Response `200`: `{"campaigns": [ ... ]}`. Each entry has `id`, `status`, `description`, `languages`, `created_at` and `approved_at`. The newest is first. The list has the newest 100 campaigns. The list has no content text.

### `GET /api/campaigns/{id}`

Response `200`:

```json
{
  "id": 12,
  "status": "draft",
  "description": "...",
  "brief": "...",
  "skipped_languages": ["en"],
  "created_at": "2026-11-10T12:00:00Z",
  "updated_at": "2026-11-10T12:00:00Z",
  "approved_at": null,
  "contents": [
    { "language": "tr", "sms": "...", "email_subject": "...", "email_body": "...",
      "sms_count": 980, "email_count": 1212, "first_name_runes": 14 }
  ]
}
```

- An `id` that is not a positive integer returns `400 invalid_request`.
- An unknown `id` returns `404 not_found`.

### `PUT /api/campaigns/{id}/contents`

Request: `{"contents": [ {"language": "tr", "sms": "...", "email_subject": "...", "email_body": "..."} ]}`.

### `POST /api/campaigns/{id}/approve`

Request: the same body as `PUT /api/campaigns/{id}/contents`.

Both endpoints return the campaign on success. The approval returns `200` also when the campaign was already approved.

| Status | Code | When |
|---|---|---|
| `400` | `invalid_request` | The body is not valid, or the id is not valid. |
| `404` | `not_found` | No campaign has this id. |
| `409` | `campaign_not_draft` | A save on a campaign that is not a draft. |
| `422` | `content_invalid` | One or more entries fail. `details` has the errors with paths. |
| `503` | `db_unavailable` | A query failed or took more than 2 seconds. |

### `GET /api/campaigns/{id}/sample`

Response `200`:

```json
{ "languages": [ { "code": "tr", "first_names": ["Abdurrahmangazi", "Ayşe", "Mehmet"] } ] }
```

- The backend uses the stored rule. The browser sends no rule.
- For each language of the campaign, the backend returns up to 5 distinct first names of reachable users.
- The first name is the longest. If two names have the same length, the backend sorts them by the name.
- The other names come from the users with the lowest `id`. Thus the result is stable.
- The query reads only `first_name`, `language`, `id`, the consent columns and the columns of the rule. No other personal data leaves the database.
- The errors are `400 invalid_request`, `404 not_found` and `503 db_unavailable`.

### Logs

For each save and each approval, the backend logs the campaign id, the languages, the counts and whether the approval was a repeat. It does not log the brief, the content text or the API key.

## Non-goals

- Sending the campaign. Slice 07 does this.
- A retry for one failed language. A later slice adds it.
- A delete of a draft.
- An edit after approval.
- An edit of the brief or the rule of a draft.
- Autosave of edits.
- Scheduled sending.

## Calls by Claude

The user did not decide these points. Claude chose them. Each one gets an entry in `docs/decisions-log.md`.

- The content endpoint saves the draft. Thus the stored content is always content that passed the validator.
- "Regenerate content" replaces the content of the same draft. A new "Find audience" makes a new draft.
- No draft is saved after a client disconnect, or when all languages fail.
- The browser sends no rule after the draft exists. The backend uses the stored rule and validates it again on each save.
- Edits are saved with "Save draft" or with the approval. There is no autosave. A close with unsaved edits asks first.
- The approval locks the row and checks the status in one transaction. A repeated approval returns the campaign and changes nothing. This replaces an idempotency key.
- The sample names are the longest first name and the names of the users with the lowest ids. Only first names reach the browser.
- The campaign stores the counts as a snapshot, but not the recipients.
- The marketer cannot add or remove a language. The only skipped languages are the failed languages.
- If a language has no reachable users at approval, the approval fails with an error. The backend does not drop the language silently.
- The preview shows first names only and the sender "Your brand".

## Settings

- No new env var.
- The save and approval body limit is 64 KiB. Each language has up to 2,240 runes of text.
