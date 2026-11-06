---
name: ste-check-edge-cases
description: Edge inputs that broke ste_check.py in the 00-skeleton Task 2 review (2026-11-06); re-test them after any change to the checker
metadata:
  type: project
---

Defects found in `.claude/skills/ste-writing/scripts/ste_check.py` in the 00-skeleton Task 2 review on 2026-11-06:

- Invalid UTF-8 input raised UnicodeDecodeError, which is not an OSError. The script exited 1, but its contract says it always exits 0.
- The sentence split did not split after `.)` or `."`, so two short sentences were counted as one long one.
- The `- ` list marker was counted as a word.
- Any ``` or ~~~ line toggled the fence state, so fences of different types did not match.
- Hard-wrapped sentences were counted one line at a time.
- `&amp;` and other HTML entities were flagged as semicolons.

**Why:** "always exit 0" and "warnings only" are contracts that `make lint-docs` depends on.
**How to apply:** when the checker changes, run these inputs again with `python3 -B`, and look for tests that cover them.
