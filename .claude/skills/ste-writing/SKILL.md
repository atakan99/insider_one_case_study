---
name: ste-writing
description: Use when writing specs, plans, progress notes, DECISIONS.md, README, CLAUDE.md, agent prompts, subagent reports or commit messages. Applies ASD-STE100 rules. Never use it for app code, comments, UI text, API errors or the app's LLM prompts.
---

# STE writing

Write so that a reader who cannot ask a question reads one meaning only. The full rule summary is in `references/rules.md`.

## Scope

**In:** text for communication between the user and agents.

- specs and plans in `docs/`
- `tasks/progress.md`
- `CLAUDE.md` and the agent files in `.claude/agents/`
- subagent reports and replies to the user
- commit messages
- `DECISIONS.md` and `README.md`

**Out:** the app itself. Do not apply this skill to:

- code, code comments and identifiers
- UI text and API error messages
- the app's LLM prompts
- the SMS and email copy that the app generates

## Checklist

1. Write one instruction per sentence.
2. Use 20 words or fewer per instruction and 25 words or fewer per description.
3. Use the active voice in instructions.
4. Do not use semicolons.
5. Do not use the present perfect ("has been", "have sent"). Use the simple past.
6. Do not put more than 3 nouns in a cluster.
7. Use 6 sentences or fewer per paragraph.
8. Use lists for sequences and conditions.
9. Do not drop articles or verbs to save words.
10. Use each term from `references/terms.md` in its one meaning only.

**Not enforced:** the STE dictionary of approximately 900 words. We cannot check it. Use plain, common words instead.

## Examples from this repo

**Passive voice → active voice**

- Before: "Queue choice is deferred to the send slice (see Open Decisions)."
- After: "The send-queue spec decides the queue. See Open Decisions."

**Dropped verb → full sentences**

- Before: "A repo where `docker compose up` starts Postgres, the Go API and the mock provider, `make verify` passes, and every later slice has a fixed place for its spec, plan, decisions and progress."
- After: "`docker compose up` starts Postgres, the Go API and the mock provider. `make verify` passes. Each later slice has a fixed place for its spec, plan, decisions and progress."

**Vague subject and slash → clear subject and words**

- Before: "Distroless has no shell, so there's no container healthcheck for api/mock-provider."
- After: "The distroless image has no shell. Thus the api and mock-provider containers have no healthcheck."

## Before you report

Run `scripts/ste_check.py` on changed `.md` files before you report:

```bash
python3 .claude/skills/ste-writing/scripts/ste_check.py <file.md>...
```

The script prints `file:line: rule: text` for each finding. It always exits 0. Fix each finding or give a reason to keep it. The script cannot tell an instruction from a description. Thus check the 20-word limit for instructions yourself.
