# Decisions

## Stack

- **Go for the backend.** Workers and rate limiters need cheap concurrency, and each service builds to one static binary.
- **Postgres as the datastore.** Segment queries need SQL. Postgres also holds the delivery status and the idempotency keys.
- **`pgx` for all queries.** `sqlc` can generate the code for the fixed queries. This project uses `pgx` directly. The fixed queries are hand-written SQL, and the backend builds the segment SQL from the validated rule.

## Audience

- **Segment rules compile to parameterized SQL.** 
    The backend validates the LLM's JSON rule and builds SQL with `$n` placeholders. Thus no user value or LLM text goes into the SQL text. The fixed queries are hand-written SQL that runs through `pgx`.
- **The backend writes the rule description.** The marketer sees what the rule really does, not what Claude says it does.
- **The operator never picks a language.** The audience query returns the languages of the audience. Claude writes content for each one.


## Content
- **The backend checks the limits of the SMS and EMAIL**.
- **Each request has an array of languages, consist of all the different locales of users.**
    For each language a llm call is made and a message is constructed. 
    An alternative would be generating a source with a language and then localizing it. 
    During my prior experience with llms I have seen that my method of choice works better.
- **Claude Opus 5.5 at low effort.** The task is a short extraction, so it needs to be fast. A fallback model handles false refusals.

## Approval
- **A row lock stops a double approval.** A second approval returns the same campaign and changes nothing.
- **Database triggers freeze an approved campaign.** The content cannot change after approval, also not by manual SQL or by code added later.


## Sending

- **Redis Streams as the queue.** Consumer groups give at-least-once delivery and retries for pending entries.
- **No duplicates is more important than no missed messages.** The worker retries `429`, `5xx` and connection refused. A timeout or a lost connection after the request left gives `unknown`, and the worker does not retry it.
- **Postgres decides who sends.** The claim is one conditional `UPDATE`. A duplicate entry in Redis cannot send a second message.
- **The worker checks consent when it sends.** It reads the user again right before the request, so a change of consent is not missed.
- **The worker takes a rate token before it claims.** If the claim takes more than 100 ms, the job goes back to `queued` and takes a new token.

## Dashboard 

- **The page polls every 5 s.** SSE is only for the progress of one campaign.
- **A basic dashboard using vue and shadcn components**. For speed.
- **It made for easy ui/ux**Because ui/ux is as important as many backend decisions. 