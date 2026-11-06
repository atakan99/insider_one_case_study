---
name: frontend-trim-defects
description: Recurring defects when code is removed from the Vue frontend (stale d.ts, leftover 401 logic, vacuous asserts, stale config globs, dead exports, peer deps)
metadata:
  type: feedback
---

Check these after each removal of code in `frontend/`:

1. `unplugin-auto-import` appends to `src/types/auto-import.d.ts` and does not remove entries. `pnpm build` does not clean it. `skipLibCheck: true` hides the broken `typeof import(...)` of removed files. Grep the d.ts for removed paths. The fix is to delete the d.ts and build again.
2. Removed auth logic can stay in other files. In slice 01 Task 1, `plugins/tanstack-vue-query/setup.ts` still skipped the toast for 401.
3. A text assert such as `toContain('Campaigns')` can pass because of another label (the sidebar group title). Check that each expected string is unique in the mounted text.

4. A deleted folder can stay in config globs. In 01 Task 2, `vite.config.ts` coverage `include` still listed `src/services/**` after `services/` was deleted.
5. A deleted feature can leave dead exports in a kept file. In 01 Task 2, `constants/themes.ts` kept `THEME_PRIMARY_COLORS` and `CONTENT_LAYOUTS` of the removed theme editor. Grep each export of a kept constants file for a consumer.
6. `pnpm-workspace.yaml` has `autoInstallPeers: false`. Thus a peer dependency with no import (`universal-cookie`, `@unovis/ts`) is not unused. Check peers before you report a package as unused.

**Why:** Items 1 to 3 all passed `pnpm test:run`, `pnpm build` and `pnpm lint` in the 01 Task 1 review (2026-11-06).

**How to apply:** In each frontend review, grep the generated d.ts files and search for the removed feature by its behaviour (status codes, store names), not only by its file names.
