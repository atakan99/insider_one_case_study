# 01 — Frontend Trim Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `frontend/` holds only the parts of the `shadcn-vue-admin` template that the campaign flow uses, and `docker compose up` serves it next to the API.

**Architecture:** Copy the template to `frontend/` and delete what `goal.md` removes. Tests for the routes, the sidebar and the page search come first, so that each later deletion is checked. The root `Dockerfile` gets a `frontend` target that builds the app with pnpm and serves `dist/` with nginx (D-008). The Makefile runs pnpm in a `node` container when pnpm is not installed, as it does for Go (D-004).

**Tech Stack:** Vue 3, Vite 8, vue-router 5 (file-based routes), Pinia, Vitest 5 with happy-dom, pnpm 11.28.8, Node 24, nginx 1.29 (unprivileged image).

**Spec:** `docs/specs/01-frontend-trim/goal.md` and `docs/specs/01-frontend-trim/requirements.md`.

## Global Constraints

- The UI text is in English only. Do not keep `vue-i18n`. Replace each `$t('…')` in a kept file with English text.
- Each campaign has one message language. This slice adds no language field.
- Keep file-based routing, `vite-plugin-vue-layouts`, `unplugin-auto-import` and `unplugin-vue-components`.
- `frontend/package.json`: `"name": "campaign-composer-web"`, `"version": "0.1.0"`, `"private": true`. No `author`, `homepage`, `repository`, `release` or `postinstall`. No `simple-git-hooks` or `nano-staged` block (NFR-6).
- `frontend/.env.example` contains only `VITE_SERVER_API_URL=http://localhost:8080`, `VITE_SERVER_API_PREFIX=/api` and `VITE_SERVER_API_TIMEOUT=10000` (NFR-10).
- Keep `frontend/LICENSE` with the original MIT notice (NFR-8).
- No text in `frontend/` refers to "Acme", the "shadcn" user, "Evil Corp", "tasks" or "billing" (NFR-5).
- The UI components to keep are the 34 names in `requirements.md`. Rule 1 there applies.

**Decisions in this plan that the spec does not make** (the user can change them in review):

1. The spec lists frontend Docker files as out of scope. The user moved them into this slice on 2026-11-06 (slice 00, D-008). Task 3 adds them.
2. Compose serves the built app with nginx on host port `3000`. Local development uses `pnpm dev` (port 5173). This closes the "Frontend dev in compose" open decision of slice 00. Log as D-010.
3. Some template files are in neither table of `goal.md`. Examples: `components/sort-select/`, `components/prop-ui/{copy,inline-tip,modal}/`, `components/global-layout/two-col*.vue`, `components/app-sidebar/validators/`, `config/site.ts`, `constants/app.ts`, `assets/icons/`. Delete each one that no kept file imports, because of the goal's rule "a file that we cannot explain costs us points". List each deleted file in the summary.
4. `components/ui/kbd` stays under rule 1, because `command-menu-panel/index.vue` imports it.
5. `pageCommands` (the flat page list for the page search) moves from `command-to-page.vue` to `constants/sidebar-data.ts`, so that a test can check FR-8.

## Review Focus

1. `pnpm install` runs the template's `postinstall: simple-git-hooks`. It writes `pre-commit` and `pre-push` hooks into the repo's `.git/hooks/`, and the hooks then block commits (this happened in slice 00). Edit `package.json` and `pnpm-workspace.yaml` before the first install. Pinned by Task 1 Step 4.
2. A browser reload on `/campaigns/123` behind nginx must load the app, not an nginx 404. Pinned by Task 3 Step 6 (curl of a deep link returns `index.html`).
3. A removed route, for example `/dashboard`, `/auth/sign-in` or `/errors/401`, must show the 404 page, not a blank page or the removed content. Pinned by Task 1 test `routes > shows 404 for removed routes`.
4. The Docker build has no `.env` file, because `.dockerignore` excludes `.env*`. Without the `VITE_*` values the env validator fails, and the app shows an "Env error" toast. Pinned by Task 3 Step 6 (`grep` for `localhost:8080` in the built JS).
5. `make test` on a machine with Docker but no Node or pnpm must still run the frontend tests (same rule as D-004). Pinned by Task 3 Step 7.

---

### Task 1: Copy the template, remove demo code, add placeholder pages

**Files:**
- Create: `frontend/` (copy of `shadcn-vue-admin-main/` without `node_modules/` and `dist/`)
- Delete in `frontend/`: each row of the "What we remove" table in `goal.md`
- Modify: `frontend/package.json`, `frontend/pnpm-workspace.yaml`, `frontend/.env.example`, `src/plugins/index.ts`, `src/router/guard/index.ts`, `src/layouts/default.vue`, `src/components/app-sidebar/{index.vue,nav-footer.vue,types.ts,data/sidebar-data.ts}`, `src/components/command-menu-panel/{index.vue,command-to-page.vue}`, `src/lib/api-client.ts`, `src/constants/sidebar-data.ts`, `src/pages/index.vue`, `src/pages/errors/data/error-pages.ts`
- Create: `src/pages/campaigns/new.vue`, `src/pages/campaigns/index.vue`, `src/pages/campaigns/[id].vue`
- Test: `src/router/__tests__/routes.test.ts`, `src/constants/__tests__/sidebar-data.test.ts`, `src/components/app-sidebar/__tests__/app-sidebar.test.ts`
- Modify tests: `src/stores/__tests__/stores.test.ts` (remove the auth case), `src/pages/errors/__tests__/error-pages.test.ts` (only 404 and 500), `src/lib/__tests__/api-client.test.ts` (remove the 401 cases)
- Delete tests: `src/composables/__tests__/use-system-config.test.ts`, `src/services/api/__tests__/example-tasks.api.test.ts`

All `src/` paths in this task are under `frontend/`.

**Interfaces:**
- Produces: routes `/campaigns/new` (name `/campaigns/new`), `/campaigns` (name `/campaigns/`), `/campaigns/[id]` (name `/campaigns/[id]`), catch-all `/[...path]`.
- Produces: `navData: NavGroup[]` with one group and two links. `pageCommands: SidebarLink[]`, the flat list that `command-to-page.vue` shows.
- Produces: `SidebarData` without `user` and `teams`.

- [x] **Step 1: Copy.** `rsync -a --exclude node_modules --exclude dist shadcn-vue-admin-main/ frontend/`. Do not run `pnpm install` yet.

- [x] **Step 2: Fix `package.json`, `pnpm-workspace.yaml` and `.env.example`.** Apply NFR-6 and NFR-10. Remove `simple-git-hooks: true` from `allowBuilds`. Remove the `simple-git-hooks` and `nano-staged` dev dependencies.

- [x] **Step 3: Write the failing tests.** Mount with `@vue/test-utils` and a router made from `setupLayouts(routes)` (`vue-router/auto-routes`, `virtual:generated-layouts`) and `createMemoryHistory()`.

  `routes.test.ts`:
  - `redirects / to /campaigns/new`: `await router.push('/')` → `router.currentRoute.value.path === '/campaigns/new'` (FR-2).
  - `resolves the campaign pages`: `router.resolve(p).name` is `/campaigns/new`, `/campaigns/` and `/campaigns/[id]` for `/campaigns/new`, `/campaigns` and `/campaigns/123`. Each matched route record has `meta.layout` unset or `'default'` (FR-3).
  - `shows 404 for unknown routes`: `/nope` resolves to name `/[...path]` (FR-4).
  - `shows 404 for removed routes`: `/dashboard`, `/tasks`, `/auth/sign-in`, `/billing` and `/settings` resolve to `/[...path]`. `resolveErrorPage('401').code === 404` (Review Focus 3).

  `sidebar-data.test.ts`:
  - `has only the campaign links`: the flat list of `navData` links equals `[{ title: 'New campaign', url: '/campaigns/new' }, { title: 'Campaigns', url: '/campaigns' }]` (compare `title` and `url` only) (FR-5).
  - `page search lists only the sidebar pages`: `pageCommands.map(c => c.url)` equals `['/campaigns/new', '/campaigns']` (FR-8).

  `app-sidebar.test.ts` (mount `AppSidebar` inside `SidebarProvider`):
  - `shows the app name and no team or user menu`: the text contains `Campaign Composer`, `New campaign` and `Campaigns`. It does not contain `Acme`, `shadcn` or `m@example.com` (FR-5).

  `error-pages.test.ts`: keep the `it.each` for `404` and `500` only. Keep the fallback cases.

- [x] **Step 4: Install, run, expect FAIL.**
  ```bash
  cd frontend && pnpm install && pnpm test:run; ls ../.git/hooks | grep -v '\.sample$'; echo "hooks_exit=$?"
  ```
  Expected: the three new test files fail (no `/campaigns/*` routes, `pageCommands` undefined, team switcher text). `hooks_exit=1` (no hook files).

- [x] **Step 5: Delete the removed files.** Each row of "What we remove" in `goal.md`. In `pages/errors/`, keep `[code].vue`, `index.vue` and the data file. Keep only `404` and `500` in `ERROR_PAGE_MAP`.

- [x] **Step 6: Edit the files that refer to removed code.** The list is in "Files" above. Rules:
  - `default.vue`: header has `SidebarTrigger`, `CommandMenuPanel` and `ToggleTheme` only (FR-6). Keep the `contentLayout` class.
  - `app-sidebar/index.vue`: the header shows a static "Campaign Composer" with one `@lucide/vue` icon. Delete `team-switcher.vue`, `nav-team-add.vue` and the footer user menu. Keep `nav-team*.vue` and `sidebar-config.ts`.
  - `navData`: one group "Campaigns" with "New campaign" (`/campaigns/new`) and "Campaigns" (`/campaigns`). Delete `settingsNavItems` and `otherPages`.
  - `api-client.ts`: remove `handleUnauthorized` and `onResponseError`. Keep `apiFetch` with `baseURL` and `timeout`.
  - `pages/index.vue`: redirect to `/campaigns/new`.

- [x] **Step 7: Add the placeholder pages.** Each page shows an `h1` and one line of text, inside the default layout:

  | File | `h1` | Line |
  |---|---|---|
  | `campaigns/new.vue` | New campaign | Write a brief. The segment, content and approval steps come in later slices. |
  | `campaigns/index.vue` | Campaigns | The list of campaigns and their status comes in a later slice. |
  | `campaigns/[id].vue` | Campaign {{ id }} | The live send dashboard comes in a later slice. |

- [x] **Step 8: Run, expect PASS.** `cd frontend && pnpm test:run && pnpm build && pnpm lint`. Expected: all pass, exit 0.

- [x] **Step 9: Commit.**
  ```bash
  git add frontend
  git commit -m "feat(frontend): trim template into frontend/, add campaign placeholder pages"
  ```

---

### Task 2: Remove unused UI components and packages

**Files:**
- Delete: `frontend/src/components/ui/<name>/` for each name not in the keep list (rule 1 applies)
- Delete: the unused files of decision 3
- Modify: `frontend/package.json`, `frontend/pnpm-lock.yaml`, `frontend/src/types/*.d.ts` (generated)

**Interfaces:**
- Consumes: the kept file set from Task 1.
- Produces: the final dependency list. Task 3 builds it in Docker with `pnpm install --frozen-lockfile`.

- [x] **Step 1: Delete the UI components that are not in the list.** Keep `kbd` (decision 4).

- [x] **Step 2: Delete the decision 3 files.** For each candidate, check `grep -rn "<path or name>" frontend/src` first. Delete only when no kept file imports it.

- [x] **Step 3: Remove the unused packages.** For each candidate in `requirements.md`, run `grep -rn "<package>" frontend/src frontend/vite.config.ts`. Remove it with `pnpm remove` when there is no match. Then run `pnpm install` (NFR-7).

- [x] **Step 4: Verify the imports and the text (NFR-4, NFR-5, NFR-9).**
  ```bash
  cd frontend && pnpm build >/dev/null && grep -rnE "vue-i18n|\\\$t\(|use-auth|stores/auth|custom-theme|language-change|team-switcher|example-(tasks|system-config)" src; echo "imports_exit=$?"; grep -rniE "acme|evil corp|m@example\.com|/tasks|/billing" . --exclude-dir=node_modules --exclude-dir=dist; echo "text_exit=$?"
  ```
  Expected: `imports_exit=1` and `text_exit=1` (no matches). `src/types/*.d.ts` are generated again by the build and are part of the first grep.

- [x] **Step 5: Run.** `cd frontend && pnpm test:run && pnpm build && pnpm lint`. Expected: exit 0.

- [x] **Step 6: Commit.**
  ```bash
  git add frontend
  git commit -m "chore(frontend): remove unused ui components and packages"
  ```

---

### Task 3: Frontend in Docker, compose and Makefile, delete the template

**Files:**
- Modify: `Dockerfile` (add targets `frontend-build` and `frontend`), `.dockerignore`, `docker-compose.yml`, `Makefile`, `DECISIONS.md`, `README.md`, `docs/architecture.md`, `tasks/backlog.json` (`01-frontend-trim.acceptance`)
- Create: `frontend/nginx.conf`
- Delete: `shadcn-vue-admin-main/`

**Interfaces:**
- Consumes: `frontend/` from Task 2. Root `Dockerfile` and Make variables from slice 00.
- Produces: compose service `frontend` (host port 3000). Make variable `PNPM`.

- [ ] **Step 1: Add the Dockerfile targets.**
  - `frontend-build`: `node:24-alpine`, `corepack enable`, copy `frontend/package.json`, `pnpm-lock.yaml` and `pnpm-workspace.yaml`, `pnpm install --frozen-lockfile`, copy `frontend/`, `pnpm build`. Build args with defaults: `VITE_SERVER_API_URL=http://localhost:8080`, `VITE_SERVER_API_PREFIX=/api`, `VITE_SERVER_API_TIMEOUT=10000`, set as `ENV` before the build.
  - `frontend`: `nginxinc/nginx-unprivileged:1.29-alpine`, copy `dist/` to `/usr/share/nginx/html` and `frontend/nginx.conf` to `/etc/nginx/conf.d/default.conf`.

- [ ] **Step 2: Write `frontend/nginx.conf`.** Listen on 8080. `try_files $uri $uri/ /index.html` (Review Focus 2). Long cache for `/assets/`, no cache for `index.html`.

- [ ] **Step 3: Add the compose service.** `frontend`: build context `.`, target `frontend`, `ports: ["3000:8080"]`, `depends_on: api` (started, not healthy, because api has no healthcheck).

- [ ] **Step 4: Update the Makefile.**
  ```make
  ifeq ($(shell command -v pnpm 2>/dev/null),)
  PNPM = docker run --rm -v "$(CURDIR)/frontend:/app" -v /app/node_modules -w /app node:24-alpine \
         sh -c 'corepack enable && pnpm install --frozen-lockfile && pnpm "$$@"' pnpm
  else
  PNPM = pnpm -C frontend
  endif
  ```
  The anonymous volume `/app/node_modules` keeps Linux packages out of the host's `frontend/node_modules`.
  - `test`: replace the `pnpm -C frontend test:run` line with `$(PNPM) test:run`.
  - `lint`: add `$(PNPM) lint`.
  - `verify`: add `$(PNPM) build` (NFR-1).
  - `up`: add a curl of `localhost:3000/` with the same retry.

- [ ] **Step 5: Update the docs.** Add two entries to `DECISIONS.md`:
  - D-009 "Frontend from the shadcn-vue-admin template, trimmed". It keeps file-based routing and auto-imports. It removes `AGENTS.md` and `.agents/`.
  - D-010 "Compose serves the built frontend with nginx" (decision 2).

  Then: `README.md`: frontend at `http://localhost:3000`, `pnpm dev` for local work. `docs/architecture.md`: add the `frontend` service. `tasks/backlog.json`: set `01-frontend-trim.acceptance` to FR-1 to FR-8 and NFR-1 to NFR-10, one short string each. Do not change `passes`.

- [ ] **Step 6: Verify compose (Review Focus 2 and 4).**
  ```bash
  make up && curl -s localhost:3000/campaigns/123 | grep -c '<div id="app">' && docker compose exec frontend sh -c 'grep -l "localhost:8080" /usr/share/nginx/html/assets/*.js | head -1'
  ```
  Expected: `make up` exits 0, then `1`, then one `.js` file path.

- [ ] **Step 7: Verify `make test` without Node, pnpm and Go.** `PATH=/usr/bin:/bin:/usr/local/bin make test`. Expected: Go and frontend tests run in containers and pass. `ls frontend/node_modules/.pnpm | head -1` is unchanged from before the run.

- [ ] **Step 8: Browser checks (FR-1 to FR-8).** Run `cd frontend && pnpm dev`. Check in the browser:
  - no console errors on load (FR-1)
  - `/` goes to `/campaigns/new` (FR-2)
  - the three pages, the 404 page, the sidebar and the header look as the FRs say (FR-3 to FR-6)
  - select dark, reload, and the page stays dark (FR-7)
  - Ctrl+K or ⌘K lists only "New campaign" and "Campaigns" (FR-8)

- [ ] **Step 9: Delete the template.** `rm -rf shadcn-vue-admin-main`. Then `make verify && make down`. Expected: exit 0.

- [ ] **Step 10: Write the summary for the user.** `requirements.md` gives its contents:
  - the extra UI components that rule 1 kept
  - the removed packages
  - the files that decision 3 deleted
  - each change that is not in `goal.md`

- [ ] **Step 11: Commit.**
  ```bash
  git add Dockerfile .dockerignore docker-compose.yml Makefile DECISIONS.md README.md docs tasks frontend
  git commit -m "chore: serve frontend in compose, pnpm in make targets, remove template"
  ```

---

## Out of scope (owned by later slices)

API calls to the Go backend, CORS on the API, SSE, a campaign language field, and all campaign features (02 to 08).
