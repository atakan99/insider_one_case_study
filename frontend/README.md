# Campaign Composer web app

The Vue 3 frontend of Campaign Composer. See the root `README.md` for the full project.

## Commands

```bash
pnpm install
pnpm dev        # http://localhost:5173
pnpm test:run
pnpm lint
pnpm build
```

Copy `.env.example` to `.env` to set the API URL.

## Origin

This app started from [shadcn-vue-admin](https://github.com/Whbbit1999/shadcn-vue-admin) (MIT, see `LICENSE`). We removed the demo pages, the mock login, the UI languages, the marketing site and the theme editor. D-009 in the root `DECISIONS.md` gives the reasons.

To add a removed shadcn-vue component again, run `pnpm dlx shadcn-vue@latest add <name>`.
