# Goal: Trim the Vue template into `frontend/`

## Why

We use the `shadcn-vue-admin` template as the base of the frontend. The template contains many demo pages, a mock login, two UI languages and a marketing site. The case does not need these parts.

Reviewers will read the repo and ask "why is this here?" about each file. A file that we cannot explain costs us points. Thus the frontend must contain only the parts that the campaign flow uses.

## Input

- Source: `shadcn-vue-admin-main/` (template, MIT license).
- The case flow: brief → segment and audience count → streamed SMS and email content → edit and approve → send → live dashboard.

## Output

1. A new folder `frontend/` at the repo root. It contains the trimmed template.
2. The folder `shadcn-vue-admin-main/` is deleted after `frontend/` builds.
3. An app shell that starts with `pnpm dev` and shows:
   - the sidebar, the header and the light/dark toggle
   - three empty placeholder pages for our flow
   - a 404 page

This task does not build any campaign feature. It only prepares a clean base.

## What we keep

| Area | Keep | Reason |
|---|---|---|
| Build config | `vite.config.ts`, `tsconfig*.json`, `index.html`, `components.json`, `eslint.config.mjs`, `pnpm-workspace.yaml`, `.gitignore`, `LICENSE` | Build and lint work without changes. `components.json` lets us add removed shadcn-vue components again with the CLI. MIT requires the license notice. |
| Routing | File-based routing, `vite-plugin-vue-layouts`, `src/router/` | Removal changes every page file. We keep it and explain it in `DECISIONS.md`. |
| Auto-imports | `unplugin-auto-import`, `unplugin-vue-components` | Same reason as routing. |
| Styles | `src/assets/*.css` | Theme and chart colors. |
| Layouts | `layouts/default.vue`, `layouts/blank.vue` | App shell and error pages. |
| Shell components | `app-sidebar/`, `global-layout/`, `toggle-theme.vue`, `loading.vue`, `custom-error.vue`, `confirm-dialog.vue`, `command-menu-panel/` | Sidebar, page header, approve/send confirmation, page search. |
| Data display | `data-table/`, `prop-ui/status-badge/` | Recipient list and delivery status list. |
| UI primitives | `components/ui/` — only the components in the list in `requirements.md` | Each one has a planned use. |
| Plugins | pinia, router, tanstack-vue-query, dayjs, nprogress, auto-animate | Server state, routing, times on the dashboard. |
| Stores | `theme.ts`, `sidebar-config.ts` | The layout and the sidebar read them. |
| Lib | `lib/utils.ts`, `lib/api-client.ts` (without the 401 logic) | One HTTP client for our Go API. |
| Env | `utils/env.ts`, `validators/env.validator.ts`, `.env.example` | Typed API URL. |
| Tests | Vitest setup, `data-table` test, `utils` tests | `pnpm test` must run from day one. |

## What we remove

| Area | Remove | Reason |
|---|---|---|
| Demo pages | `pages/` dashboard, tasks, users, apps, ai-talk, billing, settings, help-center, prop-components, marketing, auth | Demo content. Our pages replace them. |
| Error pages | `errors/` codes other than 404 and 500 | Only 404 and 500 can occur without auth. |
| Login | `stores/auth.ts`, `composables/use-auth.ts`, `config/auth.ts`, `router/guard/auth-guard.ts`, 401 handler in `api-client.ts`, sign-in and sign-up buttons | The case has no users or login. A fake login is a feature that we must explain and test for no value. |
| UI languages | `plugins/i18n/`, `language-change.vue`, `vue-i18n` | The UI text is in English only (user decision). Message language is a different feature: each campaign has one target language, for example Turkish for users in Turkey or English for users in the US. The backend and the LLM prompt handle the message language. The UI shows it as one field on the campaign. Thus the UI does not need `vue-i18n`. |
| Marketing | `layouts/marketing.vue`, `components/marketing*/`, `inspira-ui/` | Landing page. Not in scope. |
| Theme editor | `custom-theme/` (theme popover, radius, color) | Light/dark toggle is sufficient. |
| Team switcher | `team-switcher.vue`, `nav-team-add.vue`, team data, user menu in `nav-footer.vue` | No teams and no users. Replace with a static app name. |
| System config demo | `use-system-config*.ts`, `services/api/example-*.ts` | Demo API code. |
| Repo tooling | `.github/`, `.agents/`, `AGENTS.md`, `README-CN.md`, `vercel.json`, `public/_redirects`, `simple-git-hooks` and `nano-staged` config | GitHub reads `.github/` only at repo root. We write our own CI and docs later. The pre-push hook runs a full build on each push. |
| Packages | Each dependency that no kept file imports | Smaller install and less to explain. |

## New placeholder pages

Each page shows only a title and one line of text. Later tasks fill them.

| Route | Sidebar label | Later purpose |
|---|---|---|
| `/campaigns/new` | New campaign | Brief, segment, streamed content, approve |
| `/campaigns` | Campaigns | List of campaigns and their status |
| `/campaigns/[id]` | (not in sidebar) | Live send dashboard |

`/` redirects to `/campaigns/new`.

## Decisions

User decisions:

1. The UI text is in English only.
2. Each campaign has one message language. The audience decides the language, for example Turkish users or US users. One campaign does not mix languages.

Decisions in this plan that the user did not make:

1. The target folder is `frontend/` at the repo root.
2. We keep file-based routing and auto-imports.
3. We remove the template's `AGENTS.md` and `.agents/` skills.

## Out of scope

- Any API call to the Go backend.
- SSE or streaming code.
- Docker files for the frontend.
- New UI design or new colors.
