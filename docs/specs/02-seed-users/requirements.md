# Requirements: Seed users

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

- **FR-1.** `seed` applies all `goose` migrations before it inserts data.
- **FR-2.** On an empty database, `seed` inserts `SEED_USERS` random users plus the edge-case users from `goal.md`.
- **FR-3.** If `users` contains rows, `seed` makes no change and exits with code 0.
- **FR-4.** With `SEED_RESET=true`, `seed` deletes all users and events, then inserts them again.
- **FR-5.** The same `SEED_RANDOM` and seed time give the same rows.
- **FR-6.** Each `language` is `tr` or `en`. Each `country` is `TR` or `US`. Check constraints reject other values. Each `TR` user is `tr` and each `US` user is `en`.
- **FR-7.** Each `email` is unique and ends with `@example.com`.
- **FR-8.** Each `phone` matches `^\+905\d{9}$` for `TR` or `^\+1\d{3}55501\d{2}$` for `US`.
- **FR-9.** Each `city` is in the list for the user's country.
- **FR-10.** The country, language and consent ratios are within 3 percentage points of the table in `goal.md`.
- **FR-11.** For each user, the summary columns are equal to the values calculated from `events`.
- **FR-12.** Each `purchase` event has an `amount` greater than 0. Each `cart_add` event has a null `amount`.
- **FR-13.** The example brief's rule from `goal.md`, with `language = tr`, selects at least 1 user.

## Non-functional requirements

- **NFR-1.** `seed` inserts 5,000 users in less than 10 seconds. Use `pgx` `CopyFrom`.
- **NFR-2.** The migrations add indexes on `users.city`, `users.language`, `users.last_cart_add_at`, `users.last_purchase_at` and `events.user_id`.
- **NFR-3.** `docker-compose.yml` has a `seed` service with `restart: "no"`. `api` depends on it with `condition: service_completed_successfully`.
- **NFR-4.** `make up` starts the stack on an empty volume. The `users` table then contains the seeded rows.
- **NFR-5.** The generator is a pure function of the random seed and the seed time. Tests run it with no database.
- **NFR-6.** Tests for FR-1 to FR-4 and FR-11 run against Postgres. They read `TEST_DATABASE_URL`. `make test` starts a temporary Postgres container and sets this variable. Without the variable, `go test` skips these tests.
- **NFR-7.** `make verify` passes.
- **NFR-8.** The seed binary imports nothing from `internal/` except `config`, `store` and `seed`.
