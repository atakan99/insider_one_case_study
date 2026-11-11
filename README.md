# Smart Campaign Composer

A marketer writes a brief for a campaign. Claude makes a segment rule and SMS and email content from the brief. The backend validates the rule and counts the audience. The marketer edits and approves the content. Then a queue sends the messages to a mock provider, and a dashboard shows the delivery status. No real person gets a message.

## Run

```bash
docker compose up
```

- Frontend: http://localhost:3000
- API: http://localhost:8080 (`/healthz`, `/readyz`)
- Mock provider: http://localhost:8081

For frontend work with hot reload, run `pnpm dev` in `frontend/` (http://localhost:5173).

## Mock mode and Claude

The default start uses a mock LLM. It needs no API key, and all tests use it.

The mock does not read the brief. For each brief, it returns the same answers:

- The segment rule of the example brief. It selects users in Istanbul who added to the cart in the last 30 days. It excludes users who bought in the last 30 days.
- A fixed cart reminder with a 30% discount, in Turkish or English. Other languages get the English text.

All other steps are real in mock mode: the rule validation, the audience count, the limit checks, the review, the send queue and the dashboard.

To see Claude make the segment rule and the content from your brief:

1. Copy `.env.example` to `.env`.
2. In `.env`, set `LLM_MODE=anthropic`.
3. In `.env`, set `ANTHROPIC_API_KEY` to your key.
4. Run `docker compose up` again.

The default model is `claude-opus-5-5`. To change it, set `ANTHROPIC_MODEL` in `.env`. Only the `api` service gets the key. The provider stays a mock in both modes, so no real person gets a message.

## Test

```bash
make test
```

## Status

See [`tasks/backlog.json`](tasks/backlog.json).
