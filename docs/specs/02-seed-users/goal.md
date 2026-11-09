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
| `phone` | `text` not null | E.164 format. `+90 5XX` for `TR`. `+1 AAA 555 01XX` for `US`. These US numbers are reserved for fiction. |
| `language` | `text` not null | `tr` or `en`. A check constraint enforces the values. It is a separate column, so that a later seed can mix languages in one country. |
| `country` | `text` not null | `TR` or `US`. A check constraint enforces the values. |
| `city` | `text` not null | One value from a fixed list. See "City lists". |
| `birth_date` | `date` not null | Ages 18 to 75. |
| `email_consent` | `boolean` not null | Consent for the email channel. |
| `sms_consent` | `boolean` not null | Consent for the SMS channel. |
| `created_at` | `timestamptz` not null | Sign-up time, up to 3 years before the seed time. |
| `last_cart_add_at` | `timestamptz` null | Summary of `events`. Null if the user has no cart event. |
| `last_purchase_at` | `timestamptz` null | Summary of `events`. Null if the user has no purchase. |
| `order_count` | `integer` not null | Summary of `events`. |
| `total_spent` | `numeric(12,2)` not null | Summary of `events`. In TRY. |

### `events`

| Column | Type | Notes |
|---|---|---|
| `id` | `bigint` identity | Primary key. |
| `user_id` | `bigint` not null | Foreign key to `users.id`. |
| `type` | `text` not null | `cart_add` or `purchase`. |
| `amount` | `numeric(10,2)` null | Set for `purchase` only. In TRY. |
| `occurred_at` | `timestamptz` not null | Up to 180 days before the seed time. |

### Why summary columns

The segment allowlist uses only columns of `users`. Thus the compiler writes a simple `WHERE` clause with no joins. The example brief becomes this rule:

`city = Istanbul AND last_cart_add_at >= now - 30d AND (last_purchase_at IS NULL OR last_purchase_at < now - 30d)`

The `events` table keeps the real history. The seed calculates the summary columns from it. See D-014.

### Campaign language

Each campaign has one message language. The segment compiler in slice 04 adds `language = <campaign language>` to each rule. Thus a Turkish campaign never reaches an `en` user. The index on `users.language` keeps this filter fast.

## Distribution

| Property | Value |
|---|---|
| Users | 5,000 (env `SEED_USERS`) |
| Country | 70% `TR`, 30% `US` |
| Language for `TR` users | 100% `tr` |
| Language for `US` users | 100% `en` |
| Email consent | 70% |
| SMS consent | 50% |
| Users with at least one event | 60% |

Names follow the language. The phone and the city follow the country.

## City lists

The seed stores the English name of each city, with no Turkish characters. This gives one value per city for the allowlist and for Claude.

- `TR`: Istanbul (40% of `TR` users), Ankara, Izmir, Bursa, Antalya, Adana, Konya.
- `US`: New York, Los Angeles, Chicago, Houston, Miami, Seattle, Boston.

## Name sources

- `en` names come from `gofakeit`.
- `tr` names come from fixed lists in the seed package. `gofakeit` has no Turkish name data.
- The lists include names with `ç`, `ğ`, `ı`, `İ`, `ö`, `ş`, `ü`.

## Edge-case users

The seed adds these users with fixed emails, after the random users:

1. A user with a first name of 40 characters. Templates near the SMS limit must fail the length check for this user.
2. A user with Turkish characters in the first and last name, for example `Çağrı Şükrüoğlu`.
3. A user with no consent on both channels.
4. A user with SMS consent but no email consent, and the opposite.
5. A `tr` user in Istanbul who matches the example brief.

## Repeatable data

- The env `SEED_RANDOM` sets the random seed. The default is `42`.
- All times are offsets from the seed time ("12 days before"). Thus "last 30 days" segments are never empty.
- Tests can set the seed time. The same random seed and seed time give the same rows.
