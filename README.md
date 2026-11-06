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

## Test

```bash
make test
```

## Status

See [`tasks/backlog.json`](tasks/backlog.json).
