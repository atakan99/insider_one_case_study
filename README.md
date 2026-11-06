# Smart Campaign Composer

A marketer writes a brief for a campaign. Claude makes a segment rule and SMS and email content from the brief. The backend validates the rule and counts the audience. The marketer edits and approves the content. Then a queue sends the messages to a mock provider, and a dashboard shows the delivery status. No real person gets a message.

## Run

```bash
docker compose up
```

## Test

```bash
make test
```

## Status

See [`tasks/backlog.json`](tasks/backlog.json).
