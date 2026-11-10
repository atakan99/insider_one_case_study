# Final fix report, slice 04-segment-rules

## 1. Button label
- Change: `brief-audience.vue` label is now "Find audience".
- Test: new case 'labels the button Find audience' in `brief-audience.test.ts`.
- Evidence: `pnpm test:run` 66 passed (11 files).

## 2. architecture.md Flow step 1
- Change: sentence rewritten (FILTER columns in `CountAudience`; `Recipients` only returns user rows and always adds consent; slices 05 to 07 use only `Recipients`).
- Evidence: `ste_check.py docs/architecture.md` printed no findings.

## 3. Unknown key and decoder text
- Change: `validate.go` uses `truncate(k, 40)` in the unknown-key path. Decoder error text replaced by "not valid JSON".
- Test: `TestParseRule_LongUnknownKeyIsTruncated` (100-char key; path has at most 40 runes of key text). No existing test asserted the old text.
- Evidence: `go test ./internal/segment` ok.

## 4. Brief tag regexp
- Change: `(?i)<\s*/?\s*brief\s*>`.
- Test: rows `</brief >`, `< /brief>`, `<BRIEF >` in TestSanitizeBrief.

## 5. Empty Where
- Change: new `ErrEmptyWhere` ("segment: empty where"). `Recipients` and `CountAudience` check `w.sql == ""` before any query.
- Test: `TestEmptyWhereIsRejected` with a nil pool.

## 6. CORS default
- Change: default adds `http://127.0.0.1:3000` and `http://127.0.0.1:5173` in `config.go`, `docker-compose.yml`, `.env.example`. D-028 default sentence updated.
- Test: config test `TestLoad_LLMSettings` expects four origins.

## Commands
- `go test ./...`: all packages ok (segment ok, config ok, httpapi ok).
- `go vet ./...`: ok. `gofmt -l .`: empty.
- `pnpm test:run`: 11 files, 66 tests passed. `CI=true pnpm lint`: eslint clean.
- `make verify`: tail "built in 451ms" then `docker compose config -q`.
- `make verify` exit code: 0
