# Requirements: Trim the Vue template into `frontend/`

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

- **FR-1.** `pnpm dev` in `frontend/` starts the app with no console errors.
- **FR-2.** `/` redirects to `/campaigns/new`.
- **FR-3.** The routes `/campaigns/new`, `/campaigns` and `/campaigns/123` each show a placeholder page inside the default layout.
- **FR-4.** An unknown route shows the 404 page.
- **FR-5.** The sidebar shows the app name "Campaign Composer" and two links: "New campaign" and "Campaigns". It shows no team switcher and no user menu.
- **FR-6.** The header shows the sidebar toggle, the page search and the light/dark toggle. It shows no language switch and no theme editor.
- **FR-7.** The light/dark choice stays after a page reload.
- **FR-8.** The page search lists only the pages from FR-5.

## Non-functional requirements

- **NFR-1.** `pnpm build` passes with no type errors.
- **NFR-2.** `pnpm lint` passes.
- **NFR-3.** `pnpm test:run` passes. Delete tests for removed code. Do not delete tests for kept code.
- **NFR-4.** No file in `frontend/src` imports a removed file or a removed package. Use `grep` to find these imports before the build.
- **NFR-5.** No text in `frontend/` refers to the template's demo data: "Acme", "shadcn" user, "Evil Corp", "tasks", "billing".
- **NFR-6.** `package.json` contains:
  - `"name": "campaign-composer-web"`, `"version": "0.1.0"`, `"private": true`
  - no `author`, `homepage`, `repository`, `release` or `postinstall` entries
  - no `simple-git-hooks` or `nano-staged` blocks
- **NFR-7.** `pnpm-lock.yaml` is generated again after package removal.
- **NFR-8.** `frontend/LICENSE` keeps the original MIT notice.
- **NFR-9.** The generated files in `src/types/*.d.ts` are generated again by the build and do not refer to removed files.
- **NFR-10.** `.env.example` contains only `VITE_SERVER_API_URL=http://localhost:8080`, `VITE_SERVER_API_PREFIX=/api` and `VITE_SERVER_API_TIMEOUT=10000`.

## UI components to keep in `components/ui/`

Keep these and delete all others:

`alert`, `alert-dialog`, `avatar`, `badge`, `breadcrumb`, `button`, `card`, `chart`, `checkbox`, `collapsible`, `command`, `dialog`, `dropdown-menu`, `empty`, `field`, `form`, `input`, `label`, `popover`, `progress`, `scroll-area`, `select`, `separator`, `sheet`, `sidebar`, `skeleton`, `sonner`, `spinner`, `stepper`, `switch`, `table`, `tabs`, `textarea`, `tooltip`

Rules:

1. A kept component can import a component that is not in the list. If the build fails for this reason, keep that component too. Write its name in the summary.
2. To add a component again later, run `pnpm dlx shadcn-vue@latest add <name>`.

## Packages

Remove a package only when no kept file imports it. Candidates:

`vue-i18n`, `embla-carousel-vue`, `embla-carousel-autoplay`, `vaul-vue`, `vue-input-otp`, `universal-cookie`, `motion-v`, `@iconify/vue`, `@iconify-json/simple-icons`, `simple-git-hooks`, `nano-staged`

Check each candidate with `grep -r "<package>" src` before you remove it.

## Order of work

1. Copy `shadcn-vue-admin-main/` to `frontend/`. Do not copy `node_modules/` or `dist/`.
2. Remove the files in the "What we remove" table in `goal.md`.
3. Edit the files that refer to removed code: `plugins/index.ts`, `router/guard/index.ts`, `layouts/default.vue`, `app-sidebar/*`, `command-menu-panel/index.vue`, `lib/api-client.ts`, `constants/sidebar-data.ts`, `stores/__tests__/stores.test.ts`.
4. Add the placeholder pages.
5. Remove unused packages. Run `pnpm install`.
6. Run `pnpm build`, `pnpm lint` and `pnpm test:run`. Fix all errors.
7. Do the checks in FR-1 to FR-8 in the browser.
8. Delete `shadcn-vue-admin-main/`.

## Summary to write at the end

Write a short list for the user:

- the UI components that you kept in addition to the list (rule 1)
- the packages that you removed
- each change that is not in `goal.md`
