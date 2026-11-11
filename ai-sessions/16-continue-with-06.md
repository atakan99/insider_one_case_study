# Continue with 06

- Session: `c7414ebc-9dab-4372-82ec-2a9965506baa`
- Model: `claude-opus-5-5`
- Time (UTC): 2026-11-10 11:18:03 → 2026-11-10 12:12:23
- Subagent runs: 2

---

## 👤 User  <sub>2026-11-10 11:18:03</sub>

_[IDE: opened file]_ 
go on with 06

**🔧 Bash** _Read progress log, backlog, spec and plan lists_
```bash
tail -80 tasks/progress.md && cat tasks/backlog.json && ls docs/specs docs/plans
```

<details><summary>Result</summary>

```
<persisted-output>
Output too large (32.7KB). Full output saved to: ~/.claude/projects/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/tool-results/b4in1tkgv.txt

Preview (first 2KB):
# Progress

Append one entry per session. Use the template below. Put the newest entry at the bottom.

<!-- Template:
## YYYY-MM-DD — <slice id>
- Done:
- Verified with:
- Next:
- Open questions:
-->

## 2026-11-06 — 00-skeleton
- Done: Task 1 of `docs/plans/00-skeleton.md`. Moved the case brief to `docs/case.md` and the frontend-trim spec to `docs/specs/01-frontend-trim/`. Wrote `.gitignore`, `CLAUDE.md`, `DECISIONS.md` (D-001 to D-006), `README.md`, `docs/architecture.md`, `tasks/backlog.json`, `tasks/progress.md` and `ai-sessions/.gitkeep`.
- Verified with: `git check-ignore` (prints `.env`, `.env.local`, then `exit=1`). `python3 -m json.tool tasks/backlog.json` (prints `ok`).
- Next: Task 2 (ste-writing skill, executor and reviewer subagents).
- Open questions: None.

## 2026-11-06 — 00-skeleton
- Done: Task 2 of `docs/plans/00-skeleton.md`. Added the `ste-writing` skill (`SKILL.md`, `references/rules.md`, `references/terms.md`, `scripts/ste_check.py` and its tests). Added the `executor` and `reviewer` subagents. Changed the Writing style line in `CLAUDE.md`. Deleted the template git hooks in `.git/hooks/` (user decision). The reviewer found 6 defects in `ste_check.py`. All 6 are fixed, and each one has a regression test.
- Verified with: `python3 -B -m unittest discover -s .claude/skills/ste-writing/scripts` (11 tests, OK). The session lists `executor`, `reviewer` and `ste-writing`.
- Next: Task 3 (Go module, config, health endpoints, two binaries).
- Open questions: `.gitignore` does not include `__pycache__/`. The tests write it when they run without `-B`.

## 2026-11-06 — 00-skeleton
...
</persisted-output>
```

</details>

**🔧 Bash** _Show recent progress and pending backlog items_
```bash
tail -40 tasks/progress.md; python3 -c "
import json;d=json.load(open('tasks/backlog.json'))
items=d if isinstance(d,list) else d.get('slices',d)
for s in items: print(json.dumps(s,indent=1)[:3000] if not s.get('passes') else s.get('id'), s.get('passes'))
"; ls docs/specs docs/plans; git branch -a
```

<details><summary>Result</summary>

```
- Next: Task 2 (unused UI components, packages and files). The template `frontend/README.md` must be replaced, because it names removed code (NFR-5).
- Open questions: `__pycache__/` in `.gitignore` (from slice 00).

## 2026-11-06 — 01-frontend-trim
- Done: Task 2 of `docs/plans/01-frontend-trim.md`. Deleted 29 UI components (kept the 34 on the list and `kbd`). Deleted the decision 3 files that no kept file uses. Removed 10 packages. `universal-cookie` stays, because `useCookies` in `default.vue` needs it as a peer. Replaced the template `frontend/README.md` and the `index.html` title (NFR-5). The reviewer found 2 defects. Both are fixed: a coverage glob for the deleted `src/services/`, and theme-editor exports in `constants/themes.ts`.
- Verified with: `pnpm test:run` (55 tests), `pnpm build`, `CI=true pnpm lint`. The NFR-4 import grep has no matches. The NFR-5 text grep matches only the negative assertions in `__tests__`. `pnpm peers check` shows only the two issues that the template also has.
- Next: Task 3 (frontend in Docker and compose, pnpm in the Makefile, delete the template).
- Open questions: the Task 2 Step 4 text grep conflicts with the Task 1 tests. It needs `--exclude-dir=__tests__`. `__pycache__/` in `.gitignore` (from slice 00).

## 2026-11-06 — 01-frontend-trim
- Done: Task 3 of `docs/plans/01-frontend-trim.md`. The root `Dockerfile` has targets `frontend-build` and `frontend` (nginx). Compose serves the frontend on port 3000. The Makefile runs pnpm in a container when pnpm is missing. Added D-009 and D-010. Deleted `shadcn-vue-admin-main/`. User decision: the env validator uses the `.env.example` values as defaults, so `pnpm dev` works without `.env`. The reviewer found 7 defects. Fixed 6: the pnpm store in the host folder, root-owned files on Linux, the nginx redirect port, a double Cache-Control header, missing files that returned the app, and the NFR-3 acceptance text.
- Verified with: headless Chrome checks of FR-1 to FR-8 on `pnpm dev` (no `.
… [6666 more chars truncated]
```

</details>

**🔧 Skill** ```json
{
  "skill": "superpowers:brainstorming"
}
```

<details><summary>Result</summary>

```
Launching skill: superpowers:brainstorming
```

</details>

> _[skill instructions loaded: brainstorming]_

**🔧 Bash** _Read case brief, architecture and recent decisions_
```bash
cat docs/case.md; echo ----; cat docs/architecture.md; echo ----; ls docs/specs/05-content-generation; grep -n "^## \|^### " docs/decisions-log.md | tail -20; git log main --oneline | head -3
```

<details><summary>Result</summary>

```
Content is user-generated and unverified.

Smart Campaign Composer Case

## İş bağlamı

Insider One müşterileri her gün milyonlarca kişiye SMS ve email gönderiyor. Bir kampanya hazırlamak için kitleyi tanımlamak, her kanal için metin yazmak ve karakter limitlerine uymak gerekiyor. Bu case'te pazarlamacının yalnızca ne istediğini yazdığı, gerisini Claude'un önerdiği ve son kararı insanın verdiği bir araç kurmanı istiyoruz.

## Kullanıcı akışı

Örnek brief

“Son 30 günde sepetine ürün ekleyip alışveriş yapmayan İstanbul'daki kullanıcılara %30 indirim hatırlatması gönder. Samimi ama kısa olsun.”

1. **Kitle çıkarılır** Claude brief'ten bir segment kuralı üretir, arayüz kaç kişiye gideceğini gösterir.

2. **Mesajlar üretilir** SMS ve email içeriği ekrana stream edilerek akar.

3. **İnsan onaylar** Pazarlamacı önizler, gerekirse düzenler ve onaylar.

4. **Gönderilir** Mesajlar kuyruk üzerinden mock provider'a gönderilir.

5. **Canlı izlenir** Dashboard gönderim durumunu gösterir.


Gerçek kimseye mesaj gitmez. Alıcılar senin oluşturacağın sahte kullanıcılardır, gönderim de senin yazacağın mock provider'a yapılır.

## Fonksiyonel gereksinimler

### Kitle

- Claude'un ürettiği segment kuralı backend'de doğrulanır; LLM çıktısı doğrudan SQL olarak çalıştırılmaz.
- İzni olmayan kullanıcıya o kanaldan mesaj gönderilmez.

### İçerik

- İçerik kanal limitlerine uymalı: SMS en fazla 160 karakter, email konusu en fazla 80 karakter.
- `{{first_name}}` gibi kişiselleştirme değişkenleri desteklenir.

### Gönderim

- Asenkron kuyruk, kanal başına saniyede en fazla 100 mesaj.
- Aynı kişiye aynı mesaj iki kez gitmez. Retry stratejisini sen tasarla.

### Takip

- Canlı gönderim durumu, metrics ve health check endpoint'leri.

## AI gereksinimleri

Case'in en önemli kısmı. Claude'u çağırmak yeterli değil; çıktısını güvenilir hale getirmeni bekliyoruz.

- **Doğrulama:** Claude'un çıktısı yapılandırılmış (JSON) olmalı ve backend'de kontrol edilmeli.
- **Test edilebilirlik:** testler gerçek A
… [8217 more chars truncated]
```

</details>

**🔧 Bash** _Read slice 05 goal and check campaign storage_
```bash
cat docs/specs/05-content-generation/goal.md; ls backend/internal/store/migrations; ls frontend/src/pages/campaigns* -R 2>/dev/null | head; grep -rn "campaign" backend/internal/store/migrations | head
```

<details><summary>Result</summary>

```
# Goal: Content generation

## Why

The case flow has five steps. Step 2 says that Claude writes the SMS and the email, and that the content streams onto the page. The case also gives two limits: 160 characters for SMS and 80 characters for the email subject. The content supports `{{first_name}}` variables.

This slice builds step 2. Claude writes a draft for each language of the audience. The backend checks the draft against the limits after it fills in the variables. The operator sees the text arrive while Claude writes it, and sees the final text only after the backend accepts it.

## Input

- The case brief, sections "Kullanıcı akışı" step 2 and "İçerik".
- Slice 04: the brief sanitization, the rule validator, the rule description, the audience count, `internal/llm` and the panel on `/campaigns/new`. See `docs/specs/04-segment-rules/goal.md`.
- D-022: the segment response has a `languages` array. This slice uses it.
- User decisions of 2026-11-10:
  - The steps run in sequence. Content generation needs the result of the segment step.
  - The operator clicks one time. Content generation starts automatically when the segment step gives an audience.
  - The tone and style instructions in the brief go into the content prompt. There is no separate tone step or tone field. Example: "sıcak ve samimi".
  - The email body is plain text in this slice. A later slice can add formatting.

## Output

1. A Go package `internal/content`. It has the content types, the variable parser, the renderer, the validator, the prompt, the tool schema and the mock answers.
2. A streaming method in `internal/llm`, for the Anthropic client and the mock client.
3. The endpoint `POST /api/content/generate` in `internal/httpapi`. It streams Server-Sent Events (SSE).
4. On `/campaigns/new`: one content card for each language. The card shows the SMS, the email subject and the email body. The text fills in while Claude writes it.

This slice does not store content. Slice 06 adds the campaign recor
… [14211 more chars truncated]
```

</details>

**🔧 Bash** _Inspect existing campaign pages_
```bash
cd ~/Desktop/projects/case-studies/insider_one/frontend/src/pages/campaigns && wc -l *.vue && cat index.vue '[id].vue' | head -80
```

<details><summary>Result</summary>

```
14 [id].vue
      10 index.vue
      17 new.vue
      41 total
<template>
  <div class="space-y-2">
    <h1 class="text-2xl font-bold tracking-tight">
      Campaigns
    </h1>
    <p class="text-muted-foreground">
      The list of campaigns and their status comes in a later slice.
    </p>
  </div>
</template>
<script setup lang="ts">
const route = useRoute('/campaigns/[id]')
</script>

<template>
  <div class="space-y-2">
    <h1 class="text-2xl font-bold tracking-tight">
      Campaign {{ route.params.id }}
    </h1>
    <p class="text-muted-foreground">
      The live send dashboard comes in a later slice.
    </p>
  </div>
</template>
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:18:29</sub>

Slice 05 is marked done, so next up is slice 06, Review and approve. It has no spec yet. From what slice 05 left for it, it should cover a stored campaign record, a preview with a sample user, editing, and the approval itself. If you'd rather skip the questions and have me write the spec straight away, just say so.

Otherwise: once the drafts in the cards are final, what does the marketer do, step by step, up to the point where they click approve?

## 👤 User  <sub>2026-11-10 11:25:20</sub>

they should see a window where they can review, do some editing and approve and send with a button button can show a second approval popup as well thats it

## 🤖 Claude  <sub>2026-11-10 11:25:24</sub>

This might be easier if I show you. Want me to open a browser tab with some mockups?

## 👤 User  <sub>2026-11-10 11:25:57</sub>

sure

**🔧 Bash** _Read visual companion instructions_
```bash
cat ~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/brainstorming/visual-companion.md; ls ~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/brainstorming/
```

<details><summary>Result</summary>

```
# Visual Companion Guide

Browser-based visual brainstorming companion for showing mockups, diagrams, and options.

## When to Use

Decide per-question, not per-session. The test: **would the user understand this better by seeing it than reading it?**

**Use the browser** when the content itself is visual:

- **UI mockups** — wireframes, layouts, navigation structures, component designs
- **Architecture diagrams** — system components, data flow, relationship maps
- **Side-by-side visual comparisons** — comparing two layouts, two color schemes, two design directions
- **Design polish** — when the question is about look and feel, spacing, visual hierarchy
- **Spatial relationships** — state machines, flowcharts, entity relationships rendered as diagrams

**Use the terminal** when the content is text or tabular:

- **Requirements and scope questions** — "what does X mean?", "which features are in scope?"
- **Conceptual A/B/C choices** — picking between approaches described in words
- **Tradeoff lists** — pros/cons, comparison tables
- **Technical decisions** — API design, data modeling, architectural approach selection
- **Clarifying questions** — anything where the answer is words, not a visual preference

A question *about* a UI topic is not automatically a visual question. "What kind of wizard do you want?" is conceptual — use the terminal. "Which of these wizard layouts feels right?" is visual — use the browser.

## How It Works

The server watches a directory for HTML files and serves the newest one to the browser. You write HTML content to `screen_dir`, the user sees it in their browser and can click to select options. Selections are recorded to `state_dir/events` that you read on your next turn.

**Content fragments vs full documents:** If your HTML file starts with `<!DOCTYPE` or `<html`, the server serves it as-is (just injects the helper script). Otherwise, the server automatically wraps your content in the frame template — adding the header, CSS theme, conne
… [11393 more chars truncated]
```

</details>

**🔧 Bash** _Start the visual companion server_
```bash
cd ~/Desktop/projects/case-studies/insider_one && grep -n superpowers .gitignore; bash ~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/brainstorming/scripts/start-server.sh --project-dir ~/Desktop/projects/case-studies/insider_one --open
```

<details><summary>Result</summary>

```
{"type":"server-started","port":57723,"host":"127.0.0.1","url_host":"localhost","url":"http://localhost:57723/?key=2dd64a2170ba82295026ae3120c3ddb14d89fcb310671b57c81345162c475ef6","screen_dir":"~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/content","state_dir":"~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/state","idle_timeout_ms":14400000}
```

</details>

**🔧 Bash** _Check that brainstorm files are git-ignored_
```bash
git check-ignore -v .superpowers/brainstorm/x
```

<details><summary>Result</summary>

```
.git/info/exclude:18:.superpowers/	.superpowers/brainstorm/x
```

</details>

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/content/review-layout.html`
```
<h2>Review window: which layout feels right?</h2>
<p class="subtitle">Both open from a "Review & approve" button after all drafts are final. Both end with the same confirm popup (bottom).</p>

<div class="options">
  <div class="option" data-choice="a" onclick="toggleSelect(this)">
    <div class="letter">A</div>
    <div class="content">
      <h3>Large dialog over the campaign page</h3>
      <p>Edit fields on the left, preview with a sample user on the right. Tabs switch language.</p>
    </div>
  </div>
  <div class="option" data-choice="b" onclick="toggleSelect(this)">
    <div class="letter">B</div>
    <div class="content">
      <h3>Separate review page</h3>
      <p>/campaigns/:id/review. Same content, full page width, a back link to the brief.</p>
    </div>
  </div>
</div>

<div class="split">
  <div class="mockup">
    <div class="mockup-header">A — Dialog</div>
    <div class="mockup-body" style="background:#0003;padding:16px">
      <div style="background:var(--bg-primary,#fff);border-radius:8px;padding:12px;border:1px solid #8884">
        <div style="display:flex;justify-content:space-between"><b>Review campaign</b><span>✕</span></div>
        <div style="margin:8px 0"><span class="mock-button" style="padding:2px 10px">Türkçe · 1,212 email · 980 SMS</span> <span style="opacity:.6;padding:2px 10px">English · 194 · 150</span></div>
        <div style="display:flex;gap:10px">
          <div style="flex:1">
            <div class="label">SMS <span style="float:right">121 / 160</span></div>
            <div class="mock-input" style="height:48px">{{first_name}}, sepetindeki ürünler seni bekliyor! %30 indirim…</div>
            <div class="label">Email subject <span style="float:right">47 / 80</span></div>
            <div class="mock-input">{{first_name}}, sepetin için %30 indirim</div>
            <div class="label">Email body</div>
            <div class="mock-input" style="height:70px">Merhaba {{first_name}}, …</div>
          </div>
          <div style="flex:1">
            <div class="label">Preview as: <u>Ayşe Yılmaz ▾</u></div>
            <div class="placeholder" style="height:50px">📱 Ayşe, sepetindeki ürünler seni bekliyor!…</div>
            <div class="placeholder" style="height:90px">✉ Ayşe, sepetin için %30 indirim<br>Merhaba Ayşe, …</div>
          </div>
        </div>
        <div style="text-align:right;margin-top:8px"><span style="opacity:.7;margin-right:8px">Cancel</span><button class="mock-button">Approve &amp; send</button></div>
      </div>
    </div>
  </div>
  <div class="mockup">
    <div class="mockup-header">B — Page</div>
    <div class="mockup-body">
      <div class="mock-nav">← Back to brief &nbsp;|&nbsp; Campaign #12 · draft</div>
      <div style="margin:8px 0"><span class="mock-button" style="padding:2px 10px">Türkçe</span> <span style="opacity:.6">English</span></div>
      <div style="display:flex;gap:10px">
        <div style="flex:1">
          <div class="label">SMS 121/160</div><div class="mock-
… [964 more chars truncated]
```

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/content/review-layout.html (file state is current in your context — no need to Read it back)
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:26:29</sub>

I've opened the browser tab: http://localhost:57723/?key=2dd64a2170ba82295026ae3120c3ddb14d89fcb310671b57c81345162c475ef6

It shows two layouts for the review window:
- **A:** a large dialog that opens over the campaign page.
- **B:** a separate review page.

Both have the same parts: language tabs, edit fields with live character counters, and a preview that fills in a sample user's name. Below them is the confirm popup that the "Approve & send" button opens.

Which of the two feels closer to what you had in mind? You can click one to select it.

## 👤 User  <sub>2026-11-10 11:27:53</sub>

this design looks bad use tabs to show sms and show email elements as well, be careful about the design as well generate a token based simple design make it look clean, mobile friendly and usable ui

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/content/review-v2.html`
```
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Review campaign</title>
<style>
  /* ---------- Tokens (same names as shadcn-vue in the app) ---------- */
  :root {
    --background: #ffffff;
    --foreground: #0a0a0a;
    --muted: #f4f4f5;
    --muted-foreground: #71717a;
    --border: #e4e4e7;
    --input: #e4e4e7;
    --primary: #18181b;
    --primary-foreground: #fafafa;
    --accent: #f4f4f5;
    --destructive: #dc2626;
    --success: #16a34a;
    --warning: #d97706;
    --overlay: rgb(0 0 0 / .45);
    --ring: #a1a1aa;
    --bubble: #e9e9eb;

    --radius: 10px;
    --radius-sm: 6px;
    --space-1: 4px; --space-2: 8px; --space-3: 12px; --space-4: 16px; --space-6: 24px;
    --text-xs: 12px; --text-sm: 14px; --text-base: 15px; --text-lg: 18px;
    --shadow: 0 10px 30px rgb(0 0 0 / .12), 0 2px 6px rgb(0 0 0 / .06);
    --font: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --background: #0b0b0d; --foreground: #fafafa; --muted: #18181b;
      --muted-foreground: #a1a1aa; --border: #27272a; --input: #27272a;
      --primary: #fafafa; --primary-foreground: #18181b; --accent: #1f1f23;
      --overlay: rgb(0 0 0 / .65); --bubble: #2a2a2e;
    }
  }

  * { box-sizing: border-box; }
  body { margin: 0; font-family: var(--font); font-size: var(--text-base); color: var(--foreground); background: var(--muted); }

  /* Page behind the dialog */
  .page { padding: var(--space-6); max-width: 960px; margin: 0 auto; opacity: .5; }
  .page h1 { font-size: 22px; margin: 0 0 var(--space-2); }
  .page .stub { height: 80px; border-radius: var(--radius); background: var(--background); border: 1px solid var(--border); margin-bottom: var(--space-3); }

  .overlay { position: fixed; inset: 0; background: var(--overlay); display: grid; place-items: center; padding: var(--space-4); }

  /* Dialog */
  .dialog {
    width: min(920px, 100%); max-height: calc(100dvh - 32px);
    background: var(--background); border: 1px solid var(--border); border-radius: var(--radius);
    box-shadow: var(--shadow); display: flex; flex-direction: column; overflow: hidden;
  }
  .dialog-header { padding: var(--space-4) var(--space-6); border-bottom: 1px solid var(--border); display: flex; gap: var(--space-3); align-items: flex-start; }
  .dialog-header h2 { margin: 0; font-size: var(--text-lg); font-weight: 600; }
  .dialog-header p { margin: 2px 0 0; color: var(--muted-foreground); font-size: var(--text-sm); }
  .icon-btn { margin-left: auto; border: 0; background: none; color: var(--muted-foreground); font-size: 20px; line-height: 1; cursor: pointer; padding: var(--space-1); border-radius: var(--radius-sm); }
  .icon-btn:hover { background: var(--accent); color: var(--foreground); }

  .toolbar { display: flex; flex-wrap: wrap; gap: var(--space-3); align-items: center; justify-content: space-between; padding: 
… [13290 more chars truncated]
```

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/content/review-v2.html (file state is current in your context — no need to Read it back)
```

</details>

**🔧 Bash** _Screenshot the mockup at desktop and phone widths_
```bash
cd /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad 2>/dev/null || mkdir -p /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad && cd $_; C="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"; F=~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/content/review-v2.html; "$C" --headless --disable-gpu --screenshot=$PWD/desk.png --window-size=1280,860 file://$F 2>/dev/null; "$C" --headless --disable-gpu --screenshot=$PWD/mob.png --window-size=390,844 file://$F 2>/dev/null; ls $PWD
```

<details><summary>Result</summary>

```
desk.png
mob.png
```

</details>

**🔧 Read** `/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad/desk.png`

<details><summary>Result</summary>

```
[image]
```

</details>

**🔧 Read** `/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad/mob.png`

<details><summary>Result</summary>

```
[image]
```

</details>

**🔧 Bash** _Fix grid overflow and re-screenshot at phone width_
```bash
cd /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad; F=~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/content/review-v2.html; sed -i '' 's/grid-template-columns: 1fr 1fr;/grid-template-columns: minmax(0,1fr) minmax(0,1fr);/; s/.body { grid-template-columns: 1fr; gap/.body { grid-template-columns: minmax(0,1fr); gap/' $F; "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new --disable-gpu --screenshot=$PWD/mob2.png --window-size=390,844 file://$F 2>/dev/null; echo ok
```

<details><summary>Result</summary>

```
ok
[This command modified 1 file you've previously read: ../../../../../..~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/content/review-v2.html. Call Read before editing.]
```

</details>

**🔧 Read** `/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad/mob2.png`

<details><summary>Result</summary>

```
[image]
```

</details>

**🔧 Bash** _Render mockup in a 390px iframe to test phone layout_
```bash
cd /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad; F=~/Desktop/projects/case-studies/insider_one/.superpowers/brainstorm/12471-1794309965/content/review-v2.html; echo "<body style='margin:0'><iframe src='file://$F' width=390 height=844 style='border:0'></iframe>" > wrap.html; "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new --allow-file-access-from-files --screenshot=$PWD/mob3.png --window-size=800,844 file://$PWD/wrap.html 2>/dev/null; echo ok
```

<details><summary>Result</summary>

```
ok
```

</details>

**🔧 Read** `/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad/mob3.png`

<details><summary>Result</summary>

```
[image]
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:29:24</sub>

I've put a new version in the browser tab: http://localhost:57723/?key=2dd64a2170ba82295026ae3120c3ddb14d89fcb310671b57c81345162c475ef6

What changed:
- **Tabs:** the SMS / Email tabs switch the channel, and each tab shows its recipient count. A second set of tabs switches the language (Türkçe / English).
- **SMS tab:** the message editor on the left, and a phone-style preview on the right.
- **Email tab:** a subject field and a body field on the left, and an email preview on the right with from, to and the subject.
- **Editing:** this mockup is clickable. When you type, the counter (for example `101 / 160`) and the preview update. The counter turns orange near the limit and red over it. Over the limit, "Approve & send" turns off.
- **Preview as:** you pick a sample user and the preview shows their name.
- **Second approval:** "Approve & send" opens a confirm popup with the SMS count, the email count and the languages, then Back / Send now.
- **Design:** one set of colour, spacing and radius tokens, with the same names the app already uses (shadcn-vue), plus light and dark mode. On a phone the window fills the whole screen, the editor sits above the preview, and the buttons are full width. I checked it at 390 px wide and at desktop width.

What looks wrong or is missing?

## 👤 User  <sub>2026-11-10 11:29:38</sub>

yes this looks more acceptable

## 🤖 Claude  <sub>2026-11-10 11:29:52</sub>

Good. Here's how I understand the screen so far. Your words are marked as yours, and the rest is my guess.

**What it is (yours):** a review window that opens after the drafts are final. The marketer checks the content, edits it if needed, and approves. One "Approve & send" button does both. Before anything happens, it opens a second confirm popup.

**Layout (yours):** SMS and Email are tabs. The Email tab shows the subject and the body. The design is clean, built on tokens, works on a phone, and is easy to use, like the mockup.

**My guesses:**
- A "Review & approve" button on `/campaigns/new` opens the window. It turns on only when every language has a final draft or an error.
- Edits get the same checks as the Claude drafts: 160 for SMS and 80 for the subject, counted after the longest first name is filled in, and only `{{first_name}}` is allowed. The browser shows the counters. The backend checks the content again on approve and rejects anything that breaks a rule, so a changed browser can't get past the limits.
- "Preview as" picks from a few real users in the audience. Only their first names reach the browser.
- When "Send now" is confirmed, the backend saves the campaign with the brief, the rule and the final content for each language, and marks it approved. The page then goes to `/campaigns/:id`. Slice 07 makes the queue send it, and slice 08 builds the live dashboard on that page. So in this slice, "send" stops at a saved, approved campaign that waits for the queue.
- After approval, nobody can edit the content.

What's wrong or missing here? And one specific point: if Claude fails for one language (for example, English gets an error card), should the marketer be able to approve and send the other languages anyway, or must every language succeed first?

## 👤 User  <sub>2026-11-10 11:34:26</sub>

_[IDE: opened file]_ 
for now skip it marketer send it like this but in the future we should give the marketer an ability to retry that language so that it generated

**🔧 Bash** _Read memory notes, slice 05 requirements and latest decisions_
```bash
cd ~/Desktop/projects/case-studies/insider_one && cat ~/.claude/projects/-Users-USER-Desktop-projects-case-studies-insider-one/memory/*.md; cat docs/specs/05-content-generation/requirements.md; sed -n 200,260p docs/decisions-log.md
```

<details><summary>Result</summary>

```
- [Campaign language model](campaign-language-model.md) — UI English only; operator never picks language; segment returns languages array from audience (D-022)
- [Docs format](docs-format-ste.md) — specs in docs/specs/NN, plans in docs/plans/NN; STE for docs and replies, not marketing copy
- [Decisions split](decisions-split.md) — DECISIONS.md is the user's; Claude logs to docs/decisions-log.md
---
name: campaign-language-model
description: Insider One case — UI English only; users have language + country; operator never picks language; segment response has a languages array from the audience (D-022)
metadata:
  node_type: memory
  type: project
  originSessionId: 4e4c53cc-f227-4387-83cb-25ee0df21573
  modified: 2026-11-10T09:26:03.074Z
---

UI text is English only (no vue-i18n). Users have `language` (`tr`, `en`) and `country` (`TR`, `US`) as separate columns.

On 2026-11-10 (slice 04) the user decided: the operator only types a brief and never picks a language. The backend fills a `languages` array in the segment response from a `GROUP BY language` on the audience. The compiler adds NO language filter (this replaces the campaign-language part of D-013). The user wants slice 05 to be able to craft SMS + email per language in that array, but said "don't worry about 05 now" — confirm the slice 05 scope when it starts.

History: 2026-11-06 per-language variants, 2026-11-09 one language per campaign, 2026-11-10 languages array from the audience.

**Why:** the operator writes only a prompt, and the design must grow to multi-language without schema or API change.

**How to apply:** never add a language picker or a forced `language =` filter. See D-022 in docs/decisions-log.md and [[docs-format-ste]].
---
name: decisions-split
description: "Root DECISIONS.md belongs to the user (short sentences, case order); Claude logs its choices in docs/decisions-log.md"
metadata:
  node_type: memory
  type: feedback
  originSessionId: af9be78e-499c-440e-bfe7-1031b65314cf
  modified: 2
… [14887 more chars truncated]
```

</details>

**🔧 Bash** _Create slice branch and survey backend packages_
```bash
git checkout -b 06-review-approve && ls backend/internal/* frontend/src/components frontend/src/pages/campaigns; cat backend/internal/store/migrations/00001_users_events.sql | head -60; grep -n "func " backend/internal/store/*.go backend/internal/content/*.go backend/internal/segment/*.go | grep -v _test | head -60; grep -n "HandleFunc\|Handle(" -r backend/internal/httpapi backend/cmd/api | head
```

<details><summary>Result</summary>

```
Switched to a new branch '06-review-approve'
backend/internal/config:
config.go
config_test.go

backend/internal/content:
content.go
mock.go
mock_test.go
partial.go
partial_test.go
prompt.go
prompt_test.go
validate.go
validate_test.go
vars.go
vars_test.go

backend/internal/httpapi:
body.go
content.go
content_test.go
cors.go
cors_test.go
errors.go
health.go
health_test.go
segments.go
segments_test.go
sse.go
sse_test.go

backend/internal/llm:
anthropic.go
anthropic_test.go
llm.go
mock.go
mock_test.go

backend/internal/mockprovider:
failure.go
failure_test.go
limit_test.go
options.go
server.go
server_test.go
store.go
store_test.go
validate.go

backend/internal/seed:
generate.go
generate_test.go
load.go
load_test.go
names.go

backend/internal/segment:
allowlist.go
allowlist_test.go
audience.go
audience_test.go
brief.go
brief_test.go
compile.go
compile_test.go
describe.go
describe_test.go
prompt.go
prompt_test.go
rule.go
validate.go
validate_test.go

backend/internal/store:
migrate.go
migrate_test.go
migrations
storetest

frontend/src/components:
app-sidebar
campaign
command-menu-panel
confirm-dialog.vue
custom-error.vue
data-table
global-layout
loading.vue
prop-ui
toggle-theme.vue
ui

frontend/src/pages/campaigns:
[id].vue
index.vue
new.vue
-- +goose Up
CREATE TABLE users (
    id               bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    first_name       text NOT NULL,
    last_name        text NOT NULL,
    email            text NOT NULL UNIQUE,
    phone            text NOT NULL,
    language         text NOT NULL CHECK (language IN ('tr', 'en')),
    country          text NOT NULL CHECK (country IN ('TR', 'US')),
    city             text NOT NULL,
    birth_date       date NOT NULL,
    email_consent    boolean NOT NULL,
    sms_consent      boolean NOT NULL,
    created_at       timestamptz NOT NULL,
    last_cart_add_at timestamptz,
    last_purchase_at timestamptz,
    order_count      integer NOT NULL CHECK (order_count >= 0),
    total_spent      n
… [6853 more chars truncated]
```

</details>

**🔧 Bash** _Read audience, content and frontend campaign code_
```bash
sed -n 1,200p backend/internal/segment/audience.go; sed -n 1,60p backend/internal/content/validate.go; cat backend/internal/content/content.go | head -60; ls frontend/src/components/campaign; cat frontend/src/pages/campaigns/new.vue; cat frontend/src/components/confirm-dialog.vue | head -40
```

<details><summary>Result</summary>

```
package segment

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// statementTimeout limits each query. Tests change it.
var statementTimeout = "2s"

// Counts holds the audience numbers for one scope.
type Counts struct {
	Email     int64 `json:"email"`
	SMS       int64 `json:"sms"`
	Reachable int64 `json:"reachable"`
	Matched   int64 `json:"matched"`
	NoConsent int64 `json:"no_consent"`
}

// LanguageCounts holds the counts for one language.
type LanguageCounts struct {
	Code string `json:"code"`
	Counts
	// MaxFirstName is the longest first name, in characters, among the
	// reachable users. The preview response does not include it.
	MaxFirstName int64 `json:"-"`
}

// Audience holds the totals and the counts for each language.
type Audience struct {
	Counts
	Languages []LanguageCounts `json:"languages"`
}

// buildAudience drops the languages with no reachable user, sorts the rest
// and adds the totals. The totals include the dropped languages.
func buildAudience(rows []LanguageCounts) Audience {
	a := Audience{Languages: []LanguageCounts{}}
	for _, r := range rows {
		a.Email += r.Email
		a.SMS += r.SMS
		a.Reachable += r.Reachable
		a.Matched += r.Matched
		if r.Reachable > 0 {
			a.Languages = append(a.Languages, r)
		}
	}
	a.NoConsent = a.Matched - a.Reachable
	sort.Slice(a.Languages, func(i, j int) bool {
		x, y := a.Languages[i], a.Languages[j]
		if x.Reachable != y.Reachable {
			return x.Reachable > y.Reachable
		}
		return x.Code < y.Code
	})
	return a
}

// inTx runs fn in a read-only transaction with the statement timeout.
func inTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = '"+statementTimeout+"'"); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
… [6828 more chars truncated]
```

</details>

**🔧 Bash** _Check UI components, tokens and content card code_
```bash
cd ~/Desktop/projects/case-studies/insider_one/frontend && ls src/components/ui; grep -rn "\-\-warning\|\-\-success\|\-\-destructive:" src --include=*.css | head; sed -n 1,80p src/components/campaign/content-cards.vue; ls src/components/campaign/__tests__ src/composables 2>/dev/null; ls src
```

<details><summary>Result</summary>

```
alert
alert-dialog
avatar
badge
breadcrumb
button
card
chart
checkbox
collapsible
command
dialog
dropdown-menu
empty
field
form
input
kbd
label
popover
progress
scroll-area
select
separator
sheet
sidebar
skeleton
sonner
spinner
stepper
switch
table
tabs
textarea
tooltip
(eval):1: no matches found: --include=*.css
<script setup lang="ts">
import { onBeforeUnmount, reactive, ref, watch } from 'vue'

import type { Content, ContentEvent, ContentLimits } from '@/lib/content'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ContentRequestError, generateContent } from '@/lib/content'

interface LangCard {
  code: string
  email: number
  sms: number
  draft?: Content
  final?: { content: Content, limits: ContentLimits }
  retrying: boolean
  error?: string
  closed: boolean
}

const props = defineProps<{ brief: string, rule: unknown, runId: number }>()
const emit = defineEmits<{ 'update:running': [running: boolean] }>()

const NAMES: Record<string, string> = { tr: 'Turkish', en: 'English' }
const CLOSED_MESSAGE = 'The connection closed early. Try again.'

const cards = ref<LangCard[]>([])
const running = ref(false)
const requestError = ref('')
let controller: AbortController | null = null
let token = 0

function langName(code: string) {
  return NAMES[code] ?? code.toUpperCase()
}

function setRunning(v: boolean) {
  running.value = v
  emit('update:running', v)
}

function cardFor(code: string) {
  return cards.value.find(c => c.code === code)
}

function handle(e: ContentEvent) {
  if (e.event === 'audience') {
    cards.value = e.data.languages.map(l => reactive({
      code: l.code,
      email: l.email,
      sms: l.sms,
      retrying: false,
      closed: false,
    }))
    return
  }
  if (e.event === 'done')
    return
  const card = cardFor(e.data.language)
  if (!car
… [712 more chars truncated]
```

</details>

**🔧 Bash** _Find theme token definitions in frontend_
```bash
grep -rln "\-\-destructive" src/assets | head; grep -n "\-\-[a-z-]*:" $(grep -rln "\-\-destructive" src/assets | head -1) | head -40; grep -n "char\|rune\|length" src/lib/content.ts | head
```

<details><summary>Result</summary>

```
src/assets/index.css
src/assets/themes.css
7:  --background: oklch(1 0 0);
8:  --foreground: oklch(0.145 0 0);
9:  --card: oklch(1 0 0);
10:  --card-foreground: oklch(0.145 0 0);
11:  --popover: oklch(1 0 0);
12:  --popover-foreground: oklch(0.145 0 0);
13:  --primary: oklch(0.205 0 0);
14:  --primary-foreground: oklch(0.985 0 0);
15:  --secondary: oklch(0.97 0 0);
16:  --secondary-foreground: oklch(0.205 0 0);
17:  --muted: oklch(0.97 0 0);
18:  --muted-foreground: oklch(0.556 0 0);
19:  --accent: oklch(0.97 0 0);
20:  --accent-foreground: oklch(0.205 0 0);
21:  --destructive: oklch(0.577 0.245 27.325);
22:  --border: oklch(0.922 0 0);
23:  --input: oklch(0.922 0 0);
24:  --ring: oklch(0.708 0 0);
30:  --radius: 0.625rem;
31:  --sidebar: oklch(0.985 0 0);
32:  --sidebar-foreground: oklch(0.145 0 0);
33:  --sidebar-primary: oklch(0.205 0 0);
34:  --sidebar-primary-foreground: oklch(0.985 0 0);
35:  --sidebar-accent: oklch(1 0 0);
36:  --sidebar-accent-foreground: oklch(0.205 0 0);
37:  --sidebar-border: oklch(0.922 0 0);
38:  --sidebar-ring: oklch(0.708 0 0);
42:  --background: oklch(0.145 0 0);
43:  --foreground: oklch(0.985 0 0);
44:  --card: oklch(0.205 0 0);
45:  --card-foreground: oklch(0.985 0 0);
46:  --popover: oklch(0.205 0 0);
47:  --popover-foreground: oklch(0.985 0 0);
48:  --primary: oklch(0.922 0 0);
49:  --primary-foreground: oklch(0.205 0 0);
50:  --secondary: oklch(0.269 0 0);
51:  --secondary-foreground: oklch(0.985 0 0);
52:  --muted: oklch(0.269 0 0);
53:  --muted-foreground: oklch(0.708 0 0);
54:  --accent: oklch(0.269 0 0);
6:export interface ContentLimits { first_name_runes: number, sms_runes: number, email_subject_runes: number, email_body_runes: number }
44:  if (!KNOWN_EVENTS.has(name) || dataLines.length === 0)
```

</details>

**🔧 Skill** ```json
{
  "skill": "ste-writing"
}
```

<details><summary>Result</summary>

```
Launching skill: ste-writing
```

</details>

> _[skill instructions loaded: ste-writing]_

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/goal.md`
```
# Goal: Review and approve

## Why

The case flow has five steps. Step 3 says that the marketer previews the content, edits it if necessary, and approves it. Claude suggests, and the human makes the last decision.

This slice builds step 3. The marketer opens a review window, checks each message with a sample user, edits the text, and approves. A second confirm popup comes before the approval. The backend then stores the campaign. Slice 07 sends the stored campaign.

## Input

- The case brief, section "Kullanıcı akışı" step 3, and the sections "Kitle" and "İçerik".
- Slice 04: the brief sanitization, the rule validator, the audience query. See `docs/specs/04-segment-rules/goal.md`.
- Slice 05: the content types, the variable parser, the renderer, the validator, and the content cards on `/campaigns/new`. See `docs/specs/05-content-generation/goal.md`.
- The mockup that the user accepted on 2026-11-10. A copy is in `docs/specs/06-review-approve/review-mockup.html`.
- User decisions of 2026-11-10:
  - The review is one window. The marketer reviews, edits and approves in it.
  - SMS and Email are tabs. The Email tab shows the subject and the body.
  - One button "Approve & send" approves the campaign. It opens a second confirm popup first.
  - The design is clean, uses design tokens, works on a phone and is easy to use.
  - If the content of a language failed in slice 05, the marketer can still approve. The campaign skips that language.
  - A later slice can add a retry for one failed language. This slice does not.

## Output

1. A migration that adds the campaign tables.
2. A Go package `internal/campaign`. It validates an approval request and stores the campaign.
3. Three endpoints in `internal/httpapi`:
   - `POST /api/segments/sample` gives sample first names for the preview.
   - `POST /api/campaigns` stores an approved campaign.
   - `GET /api/campaigns/{id}` returns a stored campaign.
4. On `/campaigns/new`: a "Review & approve" button, the review window and the confirm popup.
5. On `/campaigns/{id}`: a read-only view of the stored campaign and its status.

This slice does not send messages. Slice 07 reads the approved campaigns and sends them. Slice 08 builds the live dashboard on `/campaigns/{id}`.

## Flow

The marketer does these steps:

1. The marketer writes the brief and clicks "Find audience". Slices 04 and 05 run as before.
2. When the content stream gives `done`, and one or more languages have a `final` event, the "Review & approve" button comes on.
3. The marketer clicks "Review & approve". The review window opens.
4. The marketer checks the SMS and the email of each language. The marketer edits the text if necessary.
5. The marketer clicks "Approve & send". The confirm popup opens.
6. The marketer clicks "Send now". The page sends the approval request.
7. On success, the page goes to `/campaigns/{id}`. The page shows the campaign with the status "Approved".

For each request to `POST /api/campaigns`, the backend does these steps i
… [11715 more chars truncated]
```

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/goal.md (file state is current in your context — no need to Read it back)
```

</details>

**🔧 Bash** _Save mockup with spec, read builder check prompt, run STE check_
```bash
cd ~/Desktop/projects/case-studies/insider_one && cp .superpowers/brainstorm/12471-1794309965/content/review-v2.html docs/specs/06-review-approve/review-mockup.html && cat ~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/brainstorming/builder-check-prompt.md; python3 .claude/skills/ste-writing/scripts/ste_check.py docs/specs/06-review-approve/goal.md
```

<details><summary>Result</summary>

````
# Builder Check Prompt

Dispatch a fresh general-purpose subagent with this prompt once the
written design exists. Fill in the domain and the document path.

```
You are a talented [DOMAIN, e.g. "software engineer", "event planner",
"bakery consultant"]. You've been handed a design document and asked to
put together a plan to build what it describes. You will not be able to
talk to the person who wrote it.

Read [DOCUMENT_PATH].

Report only the questions you'd still need answered before you could plan
this well: places where you'd have to guess and a wrong guess would
matter. For each, give the question and why the answer changes the plan.
Rank them by importance.

Skip anything the document explicitly leaves to the builder, and anything
you could reasonably decide yourself. Skip nitpicks about wording.

If nothing important is left, say "nothing important left".
```
````

</details>

**🔧 Agent** **Builder check on slice 06 spec** (general-purpose)
```
You are a talented full-stack software engineer (Go backend, Vue 3 + shadcn-vue frontend). You've been handed a design document and asked to put together a plan to build what it describes. You will not be able to talk to the person who wrote it.

Read ~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/goal.md. You may read the repo it lives in (CLAUDE.md, docs/specs/04-*/ and 05-*/, docs/decisions-log.md, backend/, frontend/src/) for context. Do not edit anything.

Report only the questions you'd still need answered before you could plan this well: places where you'd have to guess and a wrong guess would matter. For each, give the question and why the answer changes the plan. Rank them by importance.

Skip anything the document explicitly leaves to the builder, and anything you could reasonably decide yourself. Skip nitpicks about wording.

If nothing important is left, say "nothing important left".
```

<details><summary>Result</summary>

```
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a70ceeeb4eb731f59 (internal ID - do not mention to user. Use SendMessage with to: 'a70ceeeb4eb731f59', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/tasks/a70ceeeb4eb731f59.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:36:13</sub>

I've written the spec draft in [docs/specs/06-review-approve/goal.md](docs/specs/06-review-approve/goal.md) and saved the mockup next to it. A reviewer is reading the spec now. When it's done, I may come back once with a few questions. After that I'll give you the spec to approve.

**🔧 Bash** _Apply builder-check answers to the spec_
```bash
python3 - <<'EOF'
p='docs/specs/06-review-approve/goal.md'
s=open(p).read()
def rep(a,b):
    global s
    assert a in s, a
    s=s.replace(a,b,1)
rep("""- If the sample request fails, the preview uses the name "Alex" and shows a short note. The approval still works.""",
"""- If the sample request fails, the preview uses the name "Alex" and shows a short note. The approval still works.
- The select shows first names only. The mockup shows surnames and a "To:" address. The build leaves them out, because only first names reach the browser.
- The email preview shows the sender as "Your brand". Slice 07 owns the real sender. This slice adds no sender setting.""")
rep("""- On an error, the popup closes and the review window shows the message. For `content_invalid`, each error shows under its field.""",
"""- On an error, the popup closes and the review window shows the message. For `content_invalid`, each error shows under its field.
- An error with no field, for example a language with no reachable users left, shows in a banner at the top of the window.""")
rep("""- A language of the audience with no entry in `contents` is a skipped language. This is valid.""",
"""- A language of the audience with no entry in `contents` is a skipped language. This is valid.
- If a language in `contents` has no reachable users at approval, the backend returns `content_invalid` with the path `contents[i].language`. It does not skip the language silently. The marketer must run "Find audience" again.""")
rep("""- The edits stay while the page stays open.""",
"""- The edits stay while the page stays open.
- The marketer cannot leave out a language that has final content. The only skipped languages are the failed languages.
- The preview shows first names only and the sender "Your brand".""")
open(p,'w').write(s)
EOF
sed -n 1,40p docs/specs/05-content-generation/requirements.md >/dev/null; grep -n "06-review" -A6 tasks/backlog.json
```

<details><summary>Result</summary>

```
193:    "id": "06-review-approve",
194-    "title": "Review and approve",
195-    "spec": null,
196-    "acceptance": [],
197-    "passes": false
198-  },
199-  {
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:37:54</sub>

Now the requirements file.

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/requirements.md`
```
# Requirements: Review and approve

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

### Backend

- **FR-1.** `POST /api/campaigns` with the example brief, the example rule, valid `tr` content and a new UUID key returns `201`. The database has one `campaigns` row with `status = approved` and one `campaign_contents` row for `tr`.
- **FR-2.** The response of FR-1 has the campaign id, the status, the description and the content with `sms_count`, `email_count` and `first_name_runes`. The counts match `CountAudience` for the same rule.
- **FR-3.** A second request with the same key returns `200` and the same campaign. The database still has one campaign. The second body can be different, and the backend does not check it.
- **FR-4.** Ten requests with the same key at the same time give one campaign. Each response has the same id. A Postgres test covers this.
- **FR-5.** A request with no `Idempotency-Key` header, or with a key that is not a UUID, returns `400` with `code: invalid_request`. The database has no new row.
- **FR-6.** These bodies return `400` with `code: invalid_request`: a body over 64 KiB, a body that is not valid JSON, and a body with no `brief`, no `rule` or no `contents`.
- **FR-7.** A brief that fails the slice 04 checks returns `400 invalid_brief`. A rule that fails the slice 04 validator returns `400 invalid_rule` with `details`. The cases include `"Istanbul' OR 1=1 --"` for `city`.
- **FR-8.** These `contents` return `422 content_invalid` with the errors and their paths in `details`. A table test has one case for each:
  - an empty array
  - a language that is not in the audience
  - the same language two times
  - an unknown key in an entry
  - an SMS of 161 runes after rendering with the `first_name_runes` of that language
  - a subject of 81 runes after rendering, and a subject with `\n`
  - `{{last_name}}` in a field, and an HTML tag in a field
- **FR-9.** An error path has the index, for example `contents[1].sms`. With two bad entries, `details` has the errors of both.
- **FR-10.** The backend trims each field before it stores the field.
- **FR-11.** A rule with an audience that has `tr` and `en` and a body with only `tr` returns `201`. The campaign has no `en` row.
- **FR-12.** A rule with no reachable user returns `422 empty_audience`.
- **FR-13.** If a query fails or takes more than 2 seconds, the backend returns `503 db_unavailable` and stores nothing.
- **FR-14.** The insert of the campaign and its content rows is one transaction. A test that makes the content insert fail finds no campaign row.
- **FR-15.** An update of a `campaign_contents` row fails. A Postgres test covers this.
- **FR-16.** `GET /api/campaigns/{id}` returns the stored campaign. An unknown id returns `404 not_found`. An id of `abc`, `0` or `-1` returns `400 invalid_request`.
- **FR-17.** `POST /api/segments/sample` returns up to 5 distinct first names for each language with reachable users. The f
… [4531 more chars truncated]
```

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/requirements.md (file state is current in your context — no need to Read it back)
```

</details>

**🔧 Bash** _Run STE checker on spec files_
```bash
python3 .claude/skills/ste-writing/scripts/ste_check.py docs/specs/06-review-approve/goal.md docs/specs/06-review-approve/requirements.md
```

<details><summary>Result</summary>

```
docs/specs/06-review-approve/requirements.md:14: sentence-length: 30 words: **FR-6.** These bodies return CODE with CODE: a body over 64 KiB, a body that is not valid JSON, and a body with no CODE, no CODE or no CODE.
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:38:30</sub>

Written for: you, to approve before I write the plan.

The slice 06 spec is ready. It's on a new branch, `06-review-approve`, and nothing is committed yet.

- [goal.md](docs/specs/06-review-approve/goal.md) covers the flow, the review window, the data, the validation, the API and the campaign page.
- [requirements.md](docs/specs/06-review-approve/requirements.md) lists 35 functional and 12 non-functional checks.
- [review-mockup.html](docs/specs/06-review-approve/review-mockup.html) is a copy of the mockup you accepted.

**What the slice builds:**
- A "Review & approve" button opens the review window from the mockup. It has SMS and Email tabs, language tabs, live counters, a preview with a sample user, and a confirm popup.
- On "Send now", the backend checks the rule, the audience and every edit again. It then saves the campaign as approved and the page goes to `/campaigns/{id}`. Slice 07 does the real sending.
- If a language failed, the campaign skips it, as you said. The footer and the confirm popup name the skipped language. Retrying one language is listed as future work.

A reviewer read the spec and came back with no questions only you can answer.

**Calls I made that you should check while reading:**
1. The campaign is saved only at approval. There is no draft record before that.
2. A double click or a repeated request can't create two campaigns. The page sends one fixed key with each approval, and the database rejects a second campaign with the same key.
3. "Preview as" shows real first names from the audience, longest name first. No surnames, emails or phone numbers reach the browser. So I left out the surname and the "To:" line that were in the mockup. The sender shows as "Your brand" until slice 07.
4. The campaign stores the recipient counts at approval, but not the recipient list. Slice 07 gets the recipients at send time.
5. Nobody can change the stored content after approval.
6. Edits stay if the marketer closes and reopens the window. "Regenerate content" and "Find audience" throw them away.
7. The marketer can't leave out a language that has final content. Only failed languages are skipped.
8. If a language has no reachable users left at approval, the backend rejects the approval with an error. It doesn't drop that language without telling the marketer.

Each of these gets an entry in `docs/decisions-log.md` during the build.

If you approve the spec, I'll write the plan in `docs/plans/06-review-approve.md` next.

## 👤 User  <sub>2026-11-10 11:41:09</sub>

we can have drafts actually clicking them opens their preview and we can approve them then it sent

## 🤖 Claude  <sub>2026-11-10 11:41:14</sub>

Got it, that changes call 1. Campaigns get a draft state before approval. Here's my guess at how it works, so you can correct it:

- A draft campaign holds the brief, the rule and the content.
- `/campaigns` lists the campaigns, and each draft has a "Draft" badge.
- Clicking a draft opens the same review window. The marketer can preview it, edit it and approve it, and approving sends it.
- Edits in the window are saved to the draft, so the marketer can leave and come back later.
- An approved campaign can't change any more.

One question decides how drafts get created: should a draft be saved automatically when Claude finishes the content, or only when the marketer clicks a "Save draft" button in the review window?

## 👤 User  <sub>2026-11-10 11:41:33</sub>

claude can create a draft yes after claude finishes

**🔧 Read** `~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/goal.md`

<details><summary>Result</summary>

```
1	# Goal: Review and approve
2	
3	## Why
4	
5	The case flow has five steps. Step 3 says that the marketer previews the content, edits it if necessary, and approves it. Claude suggests, and the human makes the last decision.
6	
7	This slice builds step 3. The marketer opens a review window, checks each message with a sample user, edits the text, and approves. A second confirm popup comes before the approval. The backend then stores the campaign. Slice 07 sends the stored campaign.
8	
9	## Input
10	
11	- The case brief, section "Kullanıcı akışı" step 3, and the sections "Kitle" and "İçerik".
12	- Slice 04: the brief sanitization, the rule validator, the audience query. See `docs/specs/04-segment-rules/goal.md`.
13	- Slice 05: the content types, the variable parser, the renderer, the validator, and the content cards on `/campaigns/new`. See `docs/specs/05-content-generation/goal.md`.
14	- The mockup that the user accepted on 2026-11-10. A copy is in `docs/specs/06-review-approve/review-mockup.html`.
15	- User decisions of 2026-11-10:
16	  - The review is one window. The marketer reviews, edits and approves in it.
17	  - SMS and Email are tabs. The Email tab shows the subject and the body.
18	  - One button "Approve & send" approves the campaign. It opens a second confirm popup first.
19	  - The design is clean, uses design tokens, works on a phone and is easy to use.
20	  - If the content of a language failed in slice 05, the marketer can still approve. The campaign skips that language.
21	  - A later slice can add a retry for one failed language. This slice does not.
22	
23	## Output
24	
25	1. A migration that adds the campaign tables.
26	2. A Go package `internal/campaign`. It validates an approval request and stores the campaign.
27	3. Three endpoints in `internal/httpapi`:
28	   - `POST /api/segments/sample` gives sample first names for the preview.
29	   - `POST /api/campaigns` stores an approved campaign.
30	   - `GET /api/campaigns/{id}` returns a stored campaig
… [489 more chars truncated]
```

</details>

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/goal.md`
```
# Goal: Review and approve

## Why

The case flow has five steps. Step 3 says that the marketer previews the content, edits it if necessary, and approves it. Claude suggests, and the human makes the last decision.

This slice builds step 3. When Claude finishes the content, the backend saves the campaign as a draft. The marketer opens the draft in a review window, checks each message with a sample user, edits the text, and approves. A second confirm popup comes before the approval. Slice 07 sends the approved campaign.

## Input

- The case brief, section "Kullanıcı akışı" step 3, and the sections "Kitle" and "İçerik".
- Slice 04: the brief sanitization, the rule validator, the audience query. See `docs/specs/04-segment-rules/goal.md`.
- Slice 05: the content endpoint, the content types, the variable parser, the renderer, the validator, and the content cards on `/campaigns/new`. See `docs/specs/05-content-generation/goal.md`.
- The mockup that the user accepted on 2026-11-10. A copy is in `docs/specs/06-review-approve/review-mockup.html`.
- User decisions of 2026-11-10:
  - The review is one window. The marketer reviews, edits and approves in it.
  - SMS and Email are tabs. The Email tab shows the subject and the body.
  - One button "Approve & send" approves the campaign. It opens a second confirm popup first.
  - The design is clean, uses design tokens, works on a phone and is easy to use.
  - If the content of a language failed in slice 05, the marketer can still approve. The campaign skips that language.
  - A later slice can add a retry for one failed language. This slice does not.
  - Campaigns have drafts. When Claude finishes the content, the backend creates a draft.
  - A click on a draft opens its review window. The marketer approves the draft there, and then the campaign goes out.

## Output

1. A migration that adds the campaign tables.
2. A Go package `internal/campaign`. It creates, reads, updates and approves campaigns.
3. A change to `POST /api/content/generate`: at the end of the stream, the backend saves a draft and sends a `campaign` event.
4. New endpoints in `internal/httpapi`:
   - `GET /api/campaigns` lists the campaigns.
   - `GET /api/campaigns/{id}` returns one campaign.
   - `PUT /api/campaigns/{id}/contents` saves the edits of a draft.
   - `POST /api/campaigns/{id}/approve` saves the last edits and approves the draft.
   - `GET /api/campaigns/{id}/sample` gives sample first names for the preview.
5. A review window component. `/campaigns/new` and `/campaigns` both open it.
6. On `/campaigns/new`: a "Review & approve" button.
7. On `/campaigns`: the campaign list. A click on a draft opens the review window.
8. On `/campaigns/{id}`: a read-only view of an approved campaign.

This slice does not send messages. Slice 07 reads the approved campaigns and sends them. Slice 08 builds the live dashboard on `/campaigns/{id}`.

## Flow

### New campaign

1. The marketer writes the brief and clicks "Find audience". Slices 04 and 
… [16049 more chars truncated]
```

<details><summary>Result</summary>

```
The file ~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/goal.md has been updated successfully. (file state is current in your context — no need to Read it back)
```

</details>

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/requirements.md`
```
# Requirements: Review and approve

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

### Draft creation

- **FR-1.** `POST /api/content/generate` with the example brief, the example rule and the default mock ends with a `campaign` event, then `done`. The database has one campaign with `status = draft` and one `campaign_contents` row for `tr`.
- **FR-2.** The draft has the sanitized brief, the rule, the description, the final content, the counts and `first_name_runes`. The counts match `CountAudience` for the same rule.
- **FR-3.** With two languages, one final and one failed, the draft has one content row. `skipped_languages` has the failed language.
- **FR-4.** If all languages fail, the stream has no `campaign` event and the database has no new campaign.
- **FR-5.** If the draft save fails, the stream has an `error` event with no `language` and `code: draft_not_saved`, then `done`.
- **FR-6.** After a client disconnect, the backend saves no draft. The test checks that the database has no new campaign.
- **FR-7.** A request with the `campaign_id` of a draft and the same brief and rule replaces the content of that draft. The database still has one campaign.
- **FR-8.** A request with the `campaign_id` of an approved campaign, or with a different brief or rule, gives an `error` event with no `language` and `code: campaign_conflict`. The campaign does not change.

### Save and approve

- **FR-9.** `PUT /api/campaigns/{id}/contents` with valid edits returns `200`. `GET` then returns the edits, trimmed.
- **FR-10.** `POST /api/campaigns/{id}/approve` with valid content returns `200`. The campaign has `status = approved`, `approved_at` set, and the content of the request.
- **FR-11.** A second approval returns `200` and the same campaign. It changes nothing, also with a different body.
- **FR-12.** Ten approvals of one draft at the same time give one approval. `approved_at` has one value. A Postgres test covers this.
- **FR-13.** A save on an approved campaign returns `409 campaign_not_draft`. A direct SQL update of the content of an approved campaign fails, or the store blocks it. A Postgres test covers the chosen way.
- **FR-14.** These bodies return `400 invalid_request` for save and approve: a body over 64 KiB, a body that is not valid JSON, and a body with no `contents`. An id of `abc`, `0` or `-1` also returns `400`. An unknown id returns `404 not_found`.
- **FR-15.** These `contents` return `422 content_invalid` with the errors and their paths in `details`. A table test has one case for each:
  - a missing language of the draft
  - a language that is not in the draft
  - the same language two times
  - an unknown key in an entry
  - an SMS of 161 runes after rendering with the `first_name_runes` of that language
  - a subject of 81 runes after rendering, and a subject with `\n`
  - `{{last_name}}` in a field, and an HTML tag in a field
- **FR-16.** An error path has the index, for example 
… [5973 more chars truncated]
```

**🔧 Bash** _Run STE checker on updated spec_
```bash
python3 .claude/skills/ste-writing/scripts/ste_check.py docs/specs/06-review-approve/goal.md docs/specs/06-review-approve/requirements.md
```

<details><summary>Result</summary>

```
The file ~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/requirements.md has been updated successfully. (file state is current in your context — no need to Read it back)
```

</details>

<details><summary>Result</summary>

```
docs/specs/06-review-approve/goal.md:137: present-perfect: has unsaved
docs/specs/06-review-approve/requirements.md:16: sentence-length: 26 words: **FR-8.** A request with the CODE of an approved campaign, or with a different brief or rule, gives an CODE event with no CODE and CODE.
docs/specs/06-review-approve/requirements.md:25: sentence-length: 27 words: **FR-14.** These bodies return CODE for save and approve: a body over 64 KiB, a body that is not valid JSON, and a body with no CODE.
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:43:26</sub>

Written for: you, to approve before I write the plan.

I've updated the spec to add drafts: [goal.md](docs/specs/06-review-approve/goal.md) and [requirements.md](docs/specs/06-review-approve/requirements.md). Nothing is committed yet.

**How drafts work now:**
- When Claude finishes the content, the backend saves the campaign as a draft. It stores only content that passed the checks. Languages that failed are recorded as skipped.
- `/campaigns` lists all campaigns, each with a "Draft" or "Approved" badge. Clicking a draft opens the review window from the mockup. Clicking an approved campaign opens its page.
- On `/campaigns/new`, the "Review & approve" button opens the same window as soon as the draft exists.
- In the window, the marketer previews and edits the content. Then they either click "Save draft" or click "Approve & send", which opens the confirm popup before anything happens.
- An approved campaign is locked, and slice 07 sends it.

**What changed from the first version:**
- The idempotency key is gone. Approving locks the campaign row and checks that it is still a draft, in one step. So a double click or a repeated request approves only once, and the second request changes nothing.
- The browser sends the rule once, when Claude generates the content. After that, the backend uses the rule stored with the draft. It checks the rule, the audience and the character limits again on every save and on approval.

**New calls I made that you should check:**
1. "Regenerate content" replaces the content of the same draft. A new "Find audience" makes a new draft.
2. Edits are saved only with "Save draft" or with the approval. There is no autosave. Closing the window with unsaved edits asks "Discard your changes?" first.
3. No draft is saved if every language fails, or if the browser disconnects before the end.
4. Out of scope for this slice: deleting a draft, and editing a draft's brief or rule.

The earlier calls still stand: first names only in the preview, the sender shown as "Your brand", counts stored at save time but no recipient list, and no adding or removing languages in the window.

If you approve the spec, I'll write the plan next.

## 👤 User  <sub>2026-11-10 11:43:33</sub>

go on

**🔧 Skill** ```json
{
  "skill": "superpowers:writing-plans"
}
```

<details><summary>Result</summary>

```
Launching skill: superpowers:writing-plans
```

</details>

> _[skill instructions loaded: writing-plans]_

**🔧 Bash** _Look at slice 05 plan format_
```bash
head -60 docs/plans/05-content-generation.md; grep -n "^### Task\|^## " docs/plans/05-content-generation.md; wc -l docs/plans/05-content-generation.md
```

**🔧 Bash** _Read content handler, errors, body helpers and api main_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend && cat internal/httpapi/content.go internal/httpapi/errors.go internal/httpapi/body.go; cat cmd/api/main.go
```

<details><summary>Result</summary>

```
package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/segment"
)

// ContentGenerator serves POST /api/content/generate. It streams the content
// of each audience language as Server-Sent Events.
type ContentGenerator struct {
	LLM        llm.Client
	Allow      segment.Allowlist
	Count      AudienceCounter
	Now        func() time.Time // rule clock
	Clock      func() time.Time // draft throttle clock; tests inject it
	LLMTimeout time.Duration    // timeout of each LLM call
	DraftEvery time.Duration    // minimum time between two drafts of a language
}

// RegisterContent adds POST /api/content/generate.
func RegisterContent(mux *http.ServeMux, g *ContentGenerator) {
	mux.HandleFunc("POST /api/content/generate", g.generate)
}

type audienceEvent struct {
	Description string                   `json:"description"`
	Languages   []segment.LanguageCounts `json:"languages"`
}

type draftEvent struct {
	Language string          `json:"language"`
	Content  content.Content `json:"content"`
}

type retryEvent struct {
	Language string               `json:"language"`
	Errors   []content.FieldError `json:"errors,omitempty"`
}

type finalEvent struct {
	Language string          `json:"language"`
	Content  content.Content `json:"content"`
	Limits   content.Limits  `json:"limits"`
}

type languageErrorEvent struct {
	Language string               `json:"language"`
	Code     string               `json:"code"`
	Message  string               `json:"message"`
	Details  []content.FieldError `json:"details,omitempty"`
}

const noContentToolText = "You did not call " + content.ToolName + ". Do not reply with text. Call " + content.ToolName + " with your answer.\n"

func (g *ContentGenerator) generate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		Brief string          `json:"brief"`
		Rule  json.RawM
… [12084 more chars truncated]
```

</details>

<details><summary>Result</summary>

````
# 05 — Content Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After the segment step, Claude writes the SMS and the email for each language of the audience. The text streams onto the page. The backend validates it against the limits after it renders `{{first_name}}`.

**Architecture:** This slice adds one package, one LLM method and one endpoint:

```
internal/content   ParseVars, Render, Validate, Limits, ToolSchema, SystemPrompt,
                   UserMessage, RepairMessage, LanguageOf, MockContent, PartialContent
internal/llm       Client.Stream, Response.CutOff, Mock streaming and default function
internal/segment   LanguageCounts.MaxFirstName (json:"-"), filled by CountAudience
internal/httpapi   POST /api/content/generate (ContentGenerator), SSE writer
cmd/api            wires the content generator and the mock default function
frontend           lib/content.ts (stream reader), content-cards.vue, auto-start in brief-audience.vue
```

`internal/content` imports nothing from `internal/`. `internal/httpapi` joins `content`, `segment` and `llm`. `cmd/api` gives the mock its default answers (D-034).

**Tech Stack:** Go 1.27, `github.com/anthropics/anthropic-sdk-go/v2` v2.8.0 (beta streaming), `pgx/v5`, Vue 3, native `fetch` with `ReadableStream`, Vitest.

**Spec:** `docs/specs/05-content-generation/goal.md` and `requirements.md`. Decisions D-030 to D-034.

## Global Constraints

- Limits after rendering with the longest first name: SMS 160 runes, subject 80 runes, body 2,000 runes. Count runes, not bytes.
- The only variable is `{{first_name}}`, exactly, with no spaces. Each `{` or `}` outside it fails validation.
- No control characters except `\n` and `\t` in `email_body`. No `\n` or `\r` in `email_subject`. No HTML tag: `<` followed by a letter, 
… [3084 more chars truncated]
````

</details>

**🔧 Bash** _Read CORS, storetest, migrate and api wiring details_
```bash
cat internal/httpapi/cors.go | grep -n "Methods\|Headers"; grep -rn "maxBodyBytes\|countTimeout\|type AudienceCounter" internal/httpapi/*.go | grep -v _test; cat internal/store/storetest/*.go | head -60; cat internal/store/migrate.go | sed -n 1,40p; grep -n "func.*MarshalJSON\|^type Rule" internal/segment/*.go; sed -n 1,40p cmd/api/content.go 2>/dev/null; ls cmd/api
```

<details><summary>Result</summary>

```
25:				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
26:				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
internal/httpapi/body.go:14:// decodeBody reads the whole request body, up to maxBodyBytes, and decodes
internal/httpapi/body.go:19:	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
internal/httpapi/content.go:86:	cctx, cancel := context.WithTimeout(ctx, countTimeout)
internal/httpapi/segments.go:16:	maxBodyBytes = 16 * 1024
internal/httpapi/segments.go:17:	countTimeout = 2 * time.Second
internal/httpapi/segments.go:21:type AudienceCounter func(ctx context.Context, w segment.Where) (segment.Audience, error)
internal/httpapi/segments.go:114:	cctx, cancel := context.WithTimeout(ctx, countTimeout)
// Package storetest creates throwaway databases for tests.
package storetest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// NewDB creates an empty database on the server in TEST_DATABASE_URL and
// returns its URL. It skips the test when the variable is empty.
func NewDB(t testing.TB) string {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("random name: %v", err)
	}
	name := "t_" + hex.EncodeToString(buf)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, base)
	if err != nil {
		t.Fatalf("connect to test server: %v", err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		c, err := pgx.Connect(ctx, base)
		if err != nil {
			t.Errorf("connect for cleanup: %v", err)
			return
		}
		defer c.Close(ctx)
		if _, err := c.Exec(ctx, "DROP
… [1329 more chars truncated]
```

</details>

**🔧 Bash** _Survey frontend lib, router and sidebar_
```bash
cd ~/Desktop/projects/case-studies/insider_one/frontend && ls src/lib src/router src/constants src/components/app-sidebar; sed -n 1,80p src/lib/content.ts; grep -rn "campaigns" src/components/app-sidebar src/constants | head; cat package.json | grep -n "\"vue\|router\|vitest\|sonner\|uuid\|@vueuse"
```

<details><summary>Result</summary>

```
src/components/app-sidebar:
__tests__
data
index.vue
menu-button.vue
nav-team-collapsible.vue
nav-team-vercel.vue
nav-team.vue
types.ts

src/constants:
__tests__
app-config.ts
app.ts
sidebar-data.ts
themes.ts

src/lib:
__tests__
api-client.ts
content.ts
segments.ts
utils.ts

src/router:
__tests__
guard
index.ts
import type { Counts } from '@/lib/segments'

import { API_BASE_URL } from '@/constants/app-config'

export interface Content { sms: string, email_subject: string, email_body: string }
export interface ContentLimits { first_name_runes: number, sms_runes: number, email_subject_runes: number, email_body_runes: number }
export type ContentEvent
  = | { event: 'audience', data: { description: string, languages: (Counts & { code: string })[] } }
    | { event: 'draft', data: { language: string, content: Content } }
    | { event: 'retry', data: { language: string, errors?: { path: string, message: string }[] } }
    | { event: 'final', data: { language: string, content: Content, limits: ContentLimits } }
    | { event: 'error', data: { language: string, code: string, message: string } }
    | { event: 'done', data: Record<string, never> }

const FALLBACK_MESSAGE = 'Something went wrong. Try again.'
const KNOWN_EVENTS = new Set(['audience', 'draft', 'retry', 'final', 'error', 'done'])

export class ContentRequestError extends Error {
  code: string

  constructor(code: string, message: string) {
    super(message)
    this.name = 'ContentRequestError'
    this.code = code
  }
}

function parseBlock(block: string): ContentEvent | null {
  let name = ''
  const dataLines: string[] = []
  for (const line of block.split('\n')) {
    if (line === '' || line.startsWith(':'))
      continue
    const idx = line.indexOf(':')
    const field = idx === -1 ? line : line.slice(0, idx)
    let value = idx === -1 ? '' : line.slice(idx + 1)
    if (value.startsWith(' '))
      value = value.slice(1)
    if (field === 'event')
      name = value
    else if (field === 'data')
    
… [2759 more chars truncated]
```

</details>

**🔧 Bash** _Read api client, segments lib and campaign component wiring_
```bash
cat src/lib/api-client.ts | head -50; sed -n 1,60p src/lib/segments.ts; grep -n "ContentCards\|runId\|rule\|<template" src/components/campaign/brief-audience.vue | head -30; sed -n 80,200p src/components/campaign/content-cards.vue | grep -n "generateContent\|Regenerate\|function\|emit"
```

<details><summary>Result</summary>

```
/**
 * ofetch: https://github.com/unjs/ofetch
 */
import { ofetch } from 'ofetch'

import { API_BASE_URL, API_TIMEOUT } from '@/constants/app-config'

export const apiFetch = ofetch.create({
  baseURL: API_BASE_URL,
  timeout: API_TIMEOUT ?? 5000,
})
import { apiFetch } from '@/lib/api-client'

export interface Counts { email: number, sms: number, reachable: number, matched: number, no_consent: number }
export interface SegmentPreview {
  rule: unknown
  description: string
  unmapped: string[]
  audience: Counts & { languages: (Counts & { code: string })[] }
}

const FALLBACK_MESSAGE = 'Something went wrong. Try again.'

// Two LLM calls can take about 40 s, so the default API timeout is too short.
const PREVIEW_TIMEOUT_MS = 60000

export async function previewSegment(brief: string): Promise<SegmentPreview> {
  return apiFetch<SegmentPreview>('/segments/preview', {
    method: 'POST',
    body: { brief },
    timeout: PREVIEW_TIMEOUT_MS,
  })
}

export function apiErrorMessage(err: unknown): string {
  const data = (err as { data?: { error?: { message?: unknown } } } | null)?.data
  const message = data?.error?.message
  return typeof message === 'string' && message !== '' ? message : FALLBACK_MESSAGE
}
12:import ContentCards from './content-cards.vue'
25:const runId = ref(0)
35:  runId.value = 0
42:      contentRule.value = data.rule
43:      runId.value++
49:<template>
166:    <ContentCards
167:      v-if="runId > 0"
169:      :rule="contentRule"
170:      :run-id="runId"
4:async function start() {
13:    await generateContent(props.brief, props.rule, (e) => {
56:        Regenerate content
```

</details>

**🔧 Bash** _Check token mapping, toast setup and frontend test style_
```bash
grep -n "@theme\|--color-destructive\|^\.dark\|^:root" src/assets/index.css; grep -rn "toast\|Toaster" src --include=*.vue -l | head -3; ls src/components/campaign/__tests__; grep -n "fetch\|vi.stub\|mock" src/components/campaign/__tests__/content-cards.test.ts | head -8
```

<details><summary>Result</summary>

```
6::root {
41:.dark {
75:@theme inline {
90:  --color-destructive: var(--destructive);
91:  --color-destructive-foreground: var(--destructive-foreground);
159:@theme inline {
(eval):1: no matches found: --include=*.vue
brief-audience.test.ts
content-cards.test.ts
12:vi.mock('@/lib/content', async () => {
24:const gen = vi.mocked(generateContent)
46:  gen.mockImplementationOnce((_b, _r, onEvent) => {
56:  beforeEach(() => gen.mockReset())
63:    expect(gen.mock.calls[0]![0]).toBe('my brief')
64:    expect(gen.mock.calls[0]![1]).toEqual({ a: 1 })
144:    expect(gen.mock.calls[1]![0]).toBe('my brief')
145:    expect(gen.mock.calls[1]![1]).toEqual({ a: 1 })
```

</details>

**🔧 Bash** _Find toaster usage, generateContent signature, test seeding helpers_
```bash
grep -rln "Toaster\|toast(" src | head; grep -n "generateContent(" -A3 src/lib/content.ts | head; grep -rn "seed\.\(Load\|Generate\)" ../backend/internal --include=*_test.go | head -3
```

<details><summary>Result</summary>

```
src/App.vue
src/components/ui/sonner/Sonner.vue
src/components/ui/sonner/index.ts
85:export async function generateContent(
86-  brief: string,
87-  rule: unknown,
88-  onEvent: (e: ContentEvent) => void,
(eval):1: no matches found: --include=*_test.go
```

</details>

**🔧 Bash** _Check generateContent params and seed helpers_
```bash
cd ~/Desktop/projects/case-studies/insider_one && sed -n 85,95p frontend/src/lib/content.ts; grep -rln "seed.Load\|seed.Generate" backend/internal | head; grep -n "func Load\|func Generate" backend/internal/seed/*.go
```

<details><summary>Result</summary>

```
export async function generateContent(
  brief: string,
  rule: unknown,
  onEvent: (e: ContentEvent) => void,
  signal: AbortSignal,
): Promise<void> {
  try {
    const res = await fetch(`${API_BASE_URL}/content/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Accept': 'text/event-stream' },
      body: JSON.stringify({ brief, rule }),
backend/internal/seed/load_test.go
backend/internal/segment/audience_test.go
backend/internal/seed/load.go:15:func Load(ctx context.Context, pool *pgxpool.Pool, o Options, reset bool) (inserted bool, err error) {
backend/internal/seed/generate.go:55:func Generate(o Options) ([]User, []Event) {
```

</details>

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/docs/plans/06-review-approve.md`
````
# 06 — Review and Approve Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When Claude finishes the content, the backend saves a draft campaign. The marketer opens the draft in a review window, previews it with a sample user, edits it and approves it after a second confirm.

**Architecture:** This slice adds one package, one migration, five endpoints and three frontend views:

```
store/migrations  00002_campaigns.sql: campaigns, campaign_contents, immutability triggers
internal/segment  SampleFirstNames (first names only)
internal/campaign types, Store (pgx), CheckContents, Service (Save, Approve, Sample, Get, List)
internal/httpapi  campaigns.go (GET list, GET one, PUT contents, POST approve, GET sample)
                  content.go: optional campaign_id, draft save, `campaign` event
cmd/api           wires campaign.Service and the draft saver
frontend          lib/campaigns.ts, lib/limits.ts, review-dialog.vue, message-preview.vue,
                  Review button in content-cards.vue, pages campaigns/index.vue and [id].vue
```

`internal/campaign` imports only `content`, `segment` and `store` from `internal/`. `internal/httpapi` joins `campaign` with the existing packages.

**Tech Stack:** Go 1.27, `pgx/v5`, goose, Vue 3, shadcn-vue (`Dialog`, `Tabs`, `AlertDialog`, `Select`, `Textarea`, `Input`, `Badge`, `Table`), `vue-sonner`, Vitest.

**Spec:** `docs/specs/06-review-approve/goal.md` and `requirements.md`. The mockup is `docs/specs/06-review-approve/review-mockup.html`.

## Global Constraints

- Limits after rendering with `first_name_runes`: SMS 160 runes, subject 80 runes, body 2,000 runes. The backend uses `content.Validate`. The browser mirrors it in `lib/limits.ts`.
- Status values in this slice: `draft` and `approved`. Only a draft can change. After approval, the content and the brief, rule, description and skipped languages cannot change.
- Body limit 64 KiB for `PUT /api/campaigns/{id}/contents` and `POST /api/campaigns/{id}/approve`. The content endpoint keeps 16 KiB.
- Each query has the `2s` statement timeout of slice 04.
- Error codes: `invalid_request`, `not_found`, `campaign_not_draft`, `content_invalid`, `db_unavailable`. Stream error codes with no language: `draft_not_saved`, `campaign_conflict`.
- No user row in the campaign tables. Only `GET /api/campaigns/{id}/sample` sends first names. No endpoint sends a surname, an email address or a phone number.
- Logs: campaign id, languages, counts, repeat flag. Never the brief, the content text or the API key.
- UI: no hard-coded colour. New token `--warning` in `:root` and `.dark`, mapped as `--color-warning` in `@theme inline`. No horizontal scroll at 360 px. The review window fills the screen below 720 px.
- Preview sender text: `Your brand`. Fallback preview name: `Alex`.
- T
… [30720 more chars truncated]
````

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/docs/plans/06-review-approve.md (file state is current in your context — no need to Read it back)
```

</details>

**🔧 Bash** _Fix two plan lines and run STE check_
```bash
sed -i '' 's/one `LanguageContent` for each final language with `Reachable`-filtered counts and `MaxFirstName`./one `LanguageContent` for each final language. It takes `SMS`, `Email` and `MaxFirstName` from the `segment.LanguageCounts` of that language./; s/- Layout classes follow the mockup: `max-sm` full screen (`h-dvh w-screen max-w-none rounded-none` below the `md` breakpoint), `grid md:grid-cols-2`, full-width tab lists and buttons on a phone./- Layout classes follow the mockup. Below the `md` breakpoint, the dialog uses `h-dvh w-screen max-w-none rounded-none`. The body uses `grid md:grid-cols-2`. Tab lists and footer buttons are full width on a phone./' docs/plans/06-review-approve.md && python3 .claude/skills/ste-writing/scripts/ste_check.py docs/plans/06-review-approve.md; wc -l docs/plans/06-review-approve.md docs/specs/06-review-approve/*.md
```

<details><summary>Result</summary>

```
docs/plans/06-review-approve.md:42: sentence-length: 29 words: **The marketer clicks "Regenerate content" after a slow first run, and the old stream saves a draft late.** The old stream must not save after the page aborted it.
docs/plans/06-review-approve.md:44: sentence-length: 31 words: **A long Turkish name, for example CODE, joins the audience between the draft and the approval.** The approval must use the new CODE, so an SMS that fit before can fail.
docs/plans/06-review-approve.md:76: sentence-length: 27 words: Then each of these fails: an update of CODE, a delete of the content row, an insert of an CODE row, an update of CODE, and CODE.
docs/plans/06-review-approve.md:306: sentence-length: 28 words: [ ] **Step 5: Postgres check** CODE (Postgres, seed users, real CODE as the saver, default mock): after the stream, CODE returns a draft with a CODE row.
docs/plans/06-review-approve.md:351: sentence-length: 26 words: one test for each row of the spec error table and for each FR-14 body case (CODE, CODE, CODE, over 64 KiB, not JSON, no CODE)
docs/plans/06-review-approve.md:405: sentence-length: 29 words: CODE: the FR-9 cases of slice 05 (CODE 20, SMS 154 runes + one variable = 160 passes, 155 fails, subject 80 passes and 81 fails, CODE counts 1).
docs/plans/06-review-approve.md:492: sentence-length: 34 words: [ ] **Step 1: CODE.** Replace "TBD per slice" in Flow step 3 with the draft save, the review window, the save and approve endpoints, the row lock, the triggers and the sample endpoint.
docs/plans/06-review-approve.md:493: sentence-length: 27 words: [ ] **Step 2: CODE.** Add D-036 to D-045, one for each item of the spec section "Calls by Claude", and one for the triggers (Task 1).
docs/plans/06-review-approve.md:498: sentence-length: 29 words: Headless Chrome at 1280 px and at 360 px: CODE with the example brief, "Review & approve", the review window, "Approve & send", "Send now", then CODE with "Approved".
     508 docs/plans/06-review-approve.md
     348 doc
… [216 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:46:52</sub>

The plan for slice 06 is ready: [docs/plans/06-review-approve.md](docs/plans/06-review-approve.md). It has 9 tasks, and nothing is committed yet. Like the earlier plans, it leaves the commit to you after your review.

| # | Task | What it delivers |
|---|---|---|
| 1 | Migration and sample names | The campaign tables, plus database rules that lock a campaign after approval. A query for the "Preview as" first names. |
| 2 | Campaign store | Create, replace, read and list drafts. One update step that locks the row and handles save and approve. |
| 3 | Content check and service | Checks edits against the current audience and the limits. A repeated approval returns the campaign and changes nothing. |
| 4 | Draft save | The content stream saves the draft at the end. "Regenerate content" replaces the content of the same draft. |
| 5 | Endpoints and wiring | The five campaign endpoints, the 64 KiB body limit, and CORS that allows `PUT`. |
| 6 | Frontend API and limit check | The API calls, and a browser copy of the character-limit check. |
| 7 | Review window | The window from the mockup, the `--warning` token, the "Discard your changes?" prompt and the confirm popup. |
| 8 | Pages | The "Review & approve" button, the campaign list on `/campaigns`, and `/campaigns/{id}`. |
| 9 | Docs and end-to-end check | The architecture flow, decisions D-036 to D-045, the backlog checks, and a compose run with headless Chrome at 1280 px and 360 px. |

The plan also names five risks that tests pin down:
- **A late save:** an aborted stream must not save a draft after the browser left.
- **Rule key order:** Postgres stores the rule with a different key order, which must not break "Regenerate content".
- **A longer name:** a user with a longer first name can join the audience between the draft and the approval.
- **Two tabs:** the same draft can be approved in two tabs at once.
- **Escape or outside click:** closing the window this way with unsaved edits must also ask first.

How do you want to run it?
- **Subagent-driven (my recommendation):** a fresh agent builds each task and a fresh reviewer checks it before the next one starts. The backend tasks 1 to 5 depend on each other's names and types, and the approval and lock logic is where a mistake would cost most. This is also how slice 05 ran.
- **Native:** I build all tasks in this session, then one reviewer checks the whole branch at the end. This is faster and cheaper, but nothing is reviewed until the end.

## 👤 User  <sub>2026-11-10 11:47:20</sub>

go with native

**🔧 Skill** ```json
{
  "skill": "superpowers:executing-plans"
}
```

<details><summary>Result</summary>

```
Launching skill: superpowers:executing-plans
```

</details>

> _[skill instructions loaded: executing-plans]_

**🔧 Bash** _Set up the plan workspace and check tools_
```bash
S=~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills; W=$($S/subagent-driven-development/scripts/sdd-workspace docs/plans/06-review-approve.md); echo $W; ls $W; ls $S/executing-plans/scripts; which go; docker ps --format '{{.Names}}' | head
```

<details><summary>Result</summary>

```
~/Desktop/projects/case-studies/insider_one/.superpowers/sdd/06-review-approve
plan-path
task-done
task-start
/opt/homebrew/bin/go
insider_one-frontend-1
insider_one-api-1
insider_one-mock-provider-1
insider_one-redis-1
insider_one-postgres-1
```

</details>

**🔧 Skill** ```json
{
  "skill": "superpowers:test-driven-development"
}
```

<details><summary>Result</summary>

```
Launching skill: superpowers:test-driven-development
```

</details>

> _[skill instructions loaded: test-driven-development]_

**🔧 Bash** _Read Makefile test target and create the ledger_
```bash
grep -n "test" Makefile | head -30; W=.superpowers/sdd/06-review-approve; cat > $W/progress.md <<'EOF'
# SDD ledger — plan: docs/plans/06-review-approve.md
Spec: docs/specs/06-review-approve/goal.md + requirements.md
Pre-flight:
- T2 consumes nothing from T1 except the migration tables — ok.
- T3 consumes T2 Store.Update/Get and T1 SampleFirstNames via Sampler func type (signature without pool) — ok, cmd/api adapts.
- T4 consumes campaign.Draft/LanguageContent/ErrConflict/ErrNotFound from T2 — ok.
- T5 consumes T3 Service methods; CampaignService interface matches Service signatures — ok.
- T6/T7/T8 frontend: T7 consumes T6 lib names; T8 consumes T7 props/emits — ok.
EOF
echo done
```

<details><summary>Result</summary>

```
7:# Tests join the cc-test network to reach the test Postgres by name.
8:GO_TEST = docker run --rm --network cc-test -e TEST_DATABASE_URL -v "$(CURDIR)/backend:/src" -w /src golang:1.27-alpine go
9:TEST_DATABASE_URL = postgres://test:test@cc-test-pg:5432/test?sslmode=disable
14:TEST_DATABASE_URL = postgres://test:test@127.0.0.1:55432/test?sslmode=disable
30:.PHONY: up down test lint verify lint-docs
43:# Starts a temporary Postgres, runs the Go tests and removes the container.
44:# The exit code is the exit code of the Go tests.
45:test:
46:	@docker network inspect cc-test >/dev/null 2>&1 || docker network create cc-test >/dev/null
47:	@docker rm -f cc-test-pg >/dev/null 2>&1 || true
48:	docker run -d --name cc-test-pg --network cc-test \
49:		-e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test \
51:	@i=0; until docker exec cc-test-pg pg_isready -U test -d test -h 127.0.0.1 >/dev/null 2>&1; do \
52:		i=$$((i+1)); if [ $$i -gt 60 ]; then echo "test Postgres not ready"; docker rm -f cc-test-pg >/dev/null; exit 1; fi; \
54:	@export TEST_DATABASE_URL='$(TEST_DATABASE_URL)'; $(GO_TEST) test ./...; rc=$$?; \
55:		docker rm -f cc-test-pg >/dev/null 2>&1; exit $$rc
56:	$(PNPM) test:run
64:verify: lint test
done
```

</details>

**🔧 Bash** _Read local Go test setup in Makefile_
```bash
sed -n 10,30p Makefile; sed -n 49,50p Makefile
```

<details><summary>Result</summary>

```
else
GO = cd backend && go
GOFMT = cd backend && gofmt
GO_TEST = cd backend && go
TEST_DATABASE_URL = postgres://test:test@127.0.0.1:55432/test?sslmode=disable
endif

# Same for pnpm. The anonymous /app/node_modules volume keeps Linux packages
# out of the host's frontend/node_modules. The pnpm store stays in the
# container (/tmp). The container runs as root, so at the end it gives the
# files that it wrote (dist/, src/types/) back to the host user.
ifeq ($(shell command -v pnpm 2>/dev/null),)
PNPM = docker run --rm -v "$(CURDIR)/frontend:/app" -v /app/node_modules -e CI=true \
	-e HOST_ID=$(shell id -u):$(shell id -g) -w /app node:24-alpine \
	sh -c 'corepack enable && pnpm install --frozen-lockfile --store-dir /tmp/pnpm-store && pnpm "$$@"; rc=$$?; \
	chown -R "$$HOST_ID" dist src/types 2>/dev/null; exit $$rc' pnpm
else
PNPM = pnpm -C frontend
endif

.PHONY: up down test lint verify lint-docs
		-e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test \
		-p 127.0.0.1:55432:5432 postgres:17-alpine >/dev/null
```

</details>

**🔧 Bash** _Start dev test Postgres and read existing DB tests_
```bash
docker rm -f cc-dev-pg >/dev/null 2>&1; docker run -d --name cc-dev-pg -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test -p 127.0.0.1:55433:5432 postgres:17-alpine >/dev/null && until docker exec cc-dev-pg pg_isready -U test -d test -h 127.0.0.1 >/dev/null 2>&1; do sleep 1; done; echo ready; cd ~/Desktop/projects/case-studies/insider_one/backend && sed -n 1,80p internal/segment/audience_test.go; cat internal/store/migrate_test.go | head -50
```

<details><summary>Result</summary>

```
ready
package segment

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

func seedDB(t *testing.T) (context.Context, *pgxpool.Pool, []seed.User) {
	t.Helper()
	url := storetest.NewDB(t)
	ctx := context.Background()
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	o := seed.Options{Random: 42, Now: testNow, Users: 500}
	users, _ := seed.Generate(o)
	if _, err := seed.Load(ctx, pool, o, false); err != nil {
		t.Fatalf("load: %v", err)
	}
	return ctx, pool, users
}

func compile(t *testing.T, js string) Where {
	t.Helper()
	return Default().Compile(mustRule(t, js), testNow)
}

// goCounts counts the users that match, by the Go predicate.
func goCounts(users []seed.User, match func(seed.User) bool) (Audience, map[string]Counts) {
	byLang := map[string]Counts{}
	var all Counts
	add := func(c *Counts, u seed.User) {
		c.Matched++
		if u.EmailConsent {
			c.Email++
		}
		if u.SMSConsent {
			c.SMS++
		}
		if u.EmailConsent || u.SMSConsent {
			c.Reachable++
		}
	}
	for _, u := range users {
		if !match(u) {
			continue
		}
		add(&all, u)
		c := byLang[u.Language]
		add(&c, u)
		byLang[u.Language] = c
	}
	all.NoConsent = all.Matched - all.Reachable
	for k, c := range byLang {
		c.NoConsent = c.Matched - c.Reachable
		byLang[k] = c
	}
	return Audience{Counts: all}, byLang
}

func within(t *time.Time, days int) bool {
	return t != nil && !t.Before(testNow.AddDate(0, 0, -days))
}

func olderThan(t *time.Time, days int) bool {
package store_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

func migrated(t *testing.T) (context.Context, *pgx.Conn) {
	t.Helper()
	ctx 
… [1098 more chars truncated]
```

</details>

**🔧 Bash** _Find segment test helpers and seed user struct_
```bash
grep -n "testNow =\|func mustRule\|exampleRule\|const example" internal/segment/*_test.go | head; grep -n "type User struct" -A20 internal/seed/generate.go | head -25
```

<details><summary>Result</summary>

```
internal/segment/audience_test.go:95:		{"example", exampleRule, func(u seed.User) bool {
internal/segment/audience_test.go:154:	got, err = CountAudience(ctx, pool, compile(t, exampleRule))
internal/segment/audience_test.go:286:	got, err := CountAudience(ctx, pool, compile(t, exampleRule))
internal/segment/compile_test.go:11:var testNow = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
internal/segment/compile_test.go:13:func mustRule(t *testing.T, js string) Rule {
internal/segment/compile_test.go:169:	w := Default().Compile(mustRule(t, exampleRule), testNow)
internal/segment/describe_test.go:6:	got := Default().Describe(mustRule(t, exampleRule))
internal/segment/describe_test.go:11:	if again := Default().Describe(mustRule(t, exampleRule)); again != got {
internal/segment/validate_test.go:12:const exampleRule = `{
internal/segment/validate_test.go:44:	r, errs := Default().ParseRule([]byte(exampleRule))
28:type User struct {
29-	ID                                int64
30-	FirstName, LastName, Email, Phone string
31-	Language, Country, City           string
32-	BirthDate, CreatedAt              time.Time
33-	EmailConsent, SMSConsent          bool
34-	LastCartAddAt, LastPurchaseAt     *time.Time
35-	OrderCount                        int
36-	TotalSpentKurus                   int64
37-}
38-
39-// Event is one row of the events table.
40-type Event struct {
41-	ID, UserID  int64
42-	Type        string // "cart_add" or "purchase"
43-	AmountKurus *int64 // nil for cart_add
44-	OccurredAt  time.Time
45-}
46-
47-type gen struct {
48-	f      *gofakeit.Faker
```

</details>

**🔧 Bash** _Read example rule and its Go predicate_
```bash
sed -n 12,40p internal/segment/validate_test.go; sed -n 90,110p internal/segment/audience_test.go
```

<details><summary>Result</summary>

```
const exampleRule = `{
  "op": "and",
  "conditions": [
    { "field": "city", "op": "eq", "value": "Istanbul" },
    { "field": "last_cart_add_at", "op": "within_days", "value": 30 },
    { "op": "or", "conditions": [
      { "field": "last_purchase_at", "op": "is_null" },
      { "field": "last_purchase_at", "op": "older_than_days", "value": 30 }
    ]}
  ]
}`

func cond(field, op, value string) string {
	if value == "" {
		return fmt.Sprintf(`{"field":%q,"op":%q}`, field, op)
	}
	return fmt.Sprintf(`{"field":%q,"op":%q,"value":%s}`, field, op, value)
}

func root(conds ...string) string {
	return `{"op":"and","conditions":[` + strings.Join(conds, ",") + `]}`
}

func nest(depth int) string {
	s := cond("order_count", "gt", "1")
	for i := 0; i < depth; i++ {
		s = `{"op":"or","conditions":[` + s + `]}`
	}
	return s
	tests := []struct {
		name  string
		rule  string
		match func(seed.User) bool
	}{
		{"example", exampleRule, func(u seed.User) bool {
			return u.City == "Istanbul" && within(u.LastCartAddAt, 30) &&
				(u.LastPurchaseAt == nil || olderThan(u.LastPurchaseAt, 30))
		}},
		{"or group", `{"op":"or","conditions":[
			{"field":"city","op":"eq","value":"Ankara"},
			{"field":"order_count","op":"gte","value":3}]}`, func(u seed.User) bool {
			return u.City == "Ankara" || u.OrderCount >= 3
		}},
		{"age between", `{"op":"and","conditions":[{"field":"age","op":"between","value":[25,40]}]}`,
			func(u seed.User) bool {
				return !u.BirthDate.After(yearsAgo(25)) && u.BirthDate.After(yearsAgo(41))
			}},
		{"empty root", `{"op":"and","conditions":[]}`, func(seed.User) bool { return true }},
	}
	for _, tc := range tests {
```

</details>

**🔧 Bash** _Write Task 1 tests and run them to see them fail_
```bash
cat >> internal/store/migrate_test.go <<'EOF'

func TestMigrate_CampaignTriggers(t *testing.T) {
	ctx, conn := migrated(t)
	var id int64
	if err := conn.QueryRow(ctx, `INSERT INTO campaigns (brief, rule, description, status, created_at, updated_at)
VALUES ('b', '{"op":"and","conditions":[]}', 'd', 'draft', now(), now()) RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("insert campaign: %v", err)
	}
	insContent := `INSERT INTO campaign_contents (campaign_id, language, sms, email_subject, email_body, sms_count, email_count, first_name_runes)
VALUES ($1, $2, 's', 'subj', 'body', 1, 1, 5)`
	if _, err := conn.Exec(ctx, insContent, id, "tr"); err != nil {
		t.Fatalf("insert content: %v", err)
	}
	if _, err := conn.Exec(ctx, `UPDATE campaign_contents SET sms = 's2' WHERE campaign_id = $1`, id); err != nil {
		t.Fatalf("update draft content: %v", err)
	}
	if _, err := conn.Exec(ctx, `UPDATE campaigns SET status = 'approved', approved_at = now() WHERE id = $1`, id); err != nil {
		t.Fatalf("approve: %v", err)
	}
	fails := []struct {
		name string
		sql  string
		args []any
	}{
		{"update sms", `UPDATE campaign_contents SET sms = 'x' WHERE campaign_id = $1`, []any{id}},
		{"delete content", `DELETE FROM campaign_contents WHERE campaign_id = $1`, []any{id}},
		{"insert content", insContent, []any{id, "en"}},
		{"update brief", `UPDATE campaigns SET brief = 'x' WHERE id = $1`, []any{id}},
		{"back to draft", `UPDATE campaigns SET status = 'draft' WHERE id = $1`, []any{id}},
	}
	for _, f := range fails {
		if _, err := conn.Exec(ctx, f.sql, f.args...); err == nil {
			t.Errorf("%s: want error after approval, got nil", f.name)
		}
	}
}
EOF
cat > internal/segment/sample_test.go <<'EOF'
package segment

import (
	"reflect"
	"regexp"
	"sort"
	"testing"
	"unicode/utf8"

	"campaigncomposer/internal/seed"
)

// goSample computes the expected sample from the seed users.
func goSample(users []seed.User, match func(seed.User) bool, n int) map[string][]string {
	type entry struct {
		name string
		id   int64
	}
	byLang := map[string]map[string]int64{}
	for _, u := range users {
		if !match(u) || !(u.EmailConsent || u.SMSConsent) {
			continue
		}
		m := byLang[u.Language]
		if m == nil {
			m = map[string]int64{}
			byLang[u.Language] = m
		}
		if id, ok := m[u.FirstName]; !ok || u.ID < id {
			m[u.FirstName] = u.ID
		}
	}
	out := map[string][]string{}
	for lang, m := range byLang {
		var es []entry
		for name, id := range m {
			es = append(es, entry{name, id})
		}
		longest := es[0].name
		for _, e := range es {
			lr, cr := utf8.RuneCountInString(e.name), utf8.RuneCountInString(longest)
			if lr > cr || (lr == cr && e.name < longest) {
				longest = e.name
			}
		}
		sort.Slice(es, func(i, j int) bool { return es[i].id < es[j].id })
		names := []string{longest}
		for _, e := range es {
			if len(names) == n {
				break
			}
			if e.name != longest {
				names = append(names, e.name)
			}
		}
		out[lang] = names
	}
	return out
}

func exampleMatch(u seed.User) bool {
… [3483 more chars truncated]
```

<details><summary>Result</summary>

```
# campaigncomposer/internal/segment [campaigncomposer/internal/segment.test]
internal/segment/sample_test.go:69:14: undefined: SampleFirstNames
internal/segment/sample_test.go:89:14: undefined: SampleFirstNames
internal/segment/sample_test.go:115:14: undefined: SampleFirstNames
internal/segment/sample_test.go:131:12: undefined: SampleFirstNames
internal/segment/sample_test.go:135:12: undefined: SampleFirstNames
internal/segment/sample_test.go:145:7: undefined: sampleQuery
internal/segment/sample_test.go:154:15: undefined: SampleFirstNames
--- FAIL: TestMigrate_CampaignTriggers (0.05s)
    migrate_test.go:77: insert campaign: ERROR: relation "campaigns" does not exist (SQLSTATE 42P01)
FAIL
FAIL	campaigncomposer/internal/store	0.621s
FAIL	campaigncomposer/internal/segment [build failed]
FAIL
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:48:25</sub>

Red as expected. Now the migration and the sample function.

**🔧 Bash** _Write migration and SampleFirstNames, run tests_
```bash
cat > internal/store/migrations/00002_campaigns.sql <<'EOF'
-- +goose Up
CREATE TABLE campaigns (
    id                bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    brief             text NOT NULL,
    rule              jsonb NOT NULL,
    description       text NOT NULL,
    status            text NOT NULL CHECK (status IN ('draft', 'approved')),
    skipped_languages text[] NOT NULL DEFAULT '{}',
    created_at        timestamptz NOT NULL,
    updated_at        timestamptz NOT NULL,
    approved_at       timestamptz,
    CHECK ((status = 'draft') = (approved_at IS NULL))
);

CREATE INDEX campaigns_created_at_idx ON campaigns (created_at DESC, id DESC);

CREATE TABLE campaign_contents (
    campaign_id      bigint NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    language         text NOT NULL CHECK (language IN ('tr', 'en')),
    sms              text NOT NULL,
    email_subject    text NOT NULL,
    email_body       text NOT NULL,
    sms_count        bigint NOT NULL CHECK (sms_count >= 0),
    email_count      bigint NOT NULL CHECK (email_count >= 0),
    first_name_runes integer NOT NULL CHECK (first_name_runes >= 0),
    PRIMARY KEY (campaign_id, language)
);

-- Only the content of a draft can change.
-- +goose StatementBegin
CREATE FUNCTION campaign_contents_draft_only() RETURNS trigger AS $$
DECLARE
    st text;
BEGIN
    SELECT status INTO st FROM campaigns
    WHERE id = CASE WHEN TG_OP = 'DELETE' THEN OLD.campaign_id ELSE NEW.campaign_id END;
    IF st IS NOT NULL AND st <> 'draft' THEN
        RAISE EXCEPTION 'campaign content cannot change after approval';
    END IF;
    IF TG_OP = 'UPDATE' AND NEW.campaign_id <> OLD.campaign_id THEN
        RAISE EXCEPTION 'campaign content cannot move to another campaign';
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER campaign_contents_draft_only
BEFORE INSERT OR UPDATE OR DELETE ON campaign_contents
FOR EACH ROW EXECUTE FUNCTION campaign_contents_draft_only();

-- After approval, the campaign keeps its brief, rule, description, skipped
-- languages and approval time. The status can only move forward.
-- +goose StatementBegin
CREATE FUNCTION campaigns_frozen_after_approval() RETURNS trigger AS $$
BEGIN
    IF OLD.status <> 'draft' AND (
        NEW.brief IS DISTINCT FROM OLD.brief OR
        NEW.rule IS DISTINCT FROM OLD.rule OR
        NEW.description IS DISTINCT FROM OLD.description OR
        NEW.skipped_languages IS DISTINCT FROM OLD.skipped_languages OR
        NEW.approved_at IS DISTINCT FROM OLD.approved_at OR
        NEW.status = 'draft'
    ) THEN
        RAISE EXCEPTION 'campaign cannot change after approval';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER campaigns_frozen_after_approval
BEFORE UPDATE ON campaigns
FOR EACH ROW EXECUTE FUNCTION campaigns_frozen_after_approval();

-- +goose Down
DROP TABLE campaign_co
… [2379 more chars truncated]
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/store	0.246s
ok  	campaigncomposer/internal/segment	0.268s
```

</details>

**🔧 Bash** _Run Task 1 tests against dev Postgres_
```bash
export TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55433/test?sslmode=disable'; go test -count=1 -v -run 'Sample|CampaignTriggers' ./internal/store/ ./internal/segment/ 2>&1 | grep -E "^(---|ok|FAIL|\s+\w+_test)" | head -20
```

<details><summary>Result</summary>

```
--- PASS: TestMigrate_CampaignTriggers (0.12s)
ok  	campaigncomposer/internal/store	0.280s
--- PASS: TestSampleFirstNames (0.09s)
--- PASS: TestSampleFirstNames_LongestFirst (0.08s)
--- PASS: TestSampleFirstNames_NoConsentExcluded (0.08s)
--- PASS: TestSampleFirstNames_Stable (0.07s)
--- PASS: TestSampleQuery_Columns (0.00s)
--- PASS: TestSampleFirstNames_EmptyWhere (0.00s)
ok  	campaigncomposer/internal/segment	0.489s
```

</details>

**🔧 Bash** _Record Task 1 in ledger and check Rule API_
```bash
cd ~/Desktop/projects/case-studies/insider_one && echo "Task 1: complete (no commits per plan; tests: go test -run 'Sample|CampaignTriggers' ./internal/store ./internal/segment → 7/7 pass with Postgres)
Task 1: Ruling: added CHECK ((status = 'draft') = (approved_at IS NULL)) and a campaigns(created_at DESC, id DESC) index — keeps approved_at consistent and serves the list — cost if wrong: slice 07 must keep approved_at set for later statuses (it does)." >> .superpowers/sdd/06-review-approve/progress.md; sed -n 50,80p backend/internal/segment/rule.go
```

<details><summary>Result</summary>

```
func (*Group) isNode()     {}
func (*Condition) isNode() {}

// Rule is a validated segment rule. Only ParseRule makes a non-zero Rule.
// The zero Rule is an empty "and" group, which means all users.
type Rule struct{ root Group }

// Root returns the root group.
func (r Rule) Root() Group {
	g := r.root
	if g.Op == "" {
		g.Op = GroupAnd
	}
	return g
}

// FieldError is one validation error.
type FieldError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (e FieldError) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return e.Path + ": " + e.Message
}

// MarshalJSON writes the rule in the same shape as the input.
func (r Rule) MarshalJSON() ([]byte, error) { return json.Marshal(marshalGroup(r.Root())) }
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:48:57</sub>

Task 2: store tests first.

**🔧 Bash** _Write Task 2 store tests and confirm they fail_
```bash
mkdir -p ~/Desktop/projects/case-studies/insider_one/backend/internal/campaign && cat > ~/Desktop/projects/case-studies/insider_one/backend/internal/campaign/store_test.go <<'EOF'
package campaign

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

func emptyDB(t *testing.T) (context.Context, *pgxpool.Pool) {
	t.Helper()
	url := storetest.NewDB(t)
	ctx := context.Background()
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return ctx, pool
}

const emptyRule = `{"op":"and","conditions":[]}`

func lc(lang, sms string) LanguageContent {
	return LanguageContent{
		Language:       lang,
		Content:        content.Content{SMS: sms, EmailSubject: "Subject " + lang, EmailBody: "Body " + lang},
		SMSCount:       10,
		EmailCount:     20,
		FirstNameRunes: 7,
	}
}

func draft(brief string, contents ...LanguageContent) Draft {
	return Draft{Brief: brief, Description: "desc", Rule: json.RawMessage(emptyRule), Skipped: []string{}, Contents: contents}
}

func TestStore_CreateAndGet(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	d := draft("brief", lc("tr", "merhaba"), lc("en", "hello"))
	d.Skipped = []string{}
	id, err := s.CreateDraft(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	c, err := s.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != id || c.Status != StatusDraft || c.Brief != "brief" || c.Description != "desc" || c.ApprovedAt != nil {
		t.Fatalf("campaign = %+v", c)
	}
	if len(c.Contents) != 2 || c.Contents[0].Language != "en" || c.Contents[1].Language != "tr" {
		t.Fatalf("contents not sorted by language: %+v", c.Contents)
	}
	if c.Contents[1] != lc("tr", "merhaba") {
		t.Fatalf("tr content = %+v", c.Contents[1])
	}
	if c.SkippedLanguages == nil || len(c.SkippedLanguages) != 0 {
		t.Fatalf("skipped = %#v, want empty non-nil", c.SkippedLanguages)
	}
	var rule map[string]any
	if err := json.Unmarshal(c.Rule, &rule); err != nil || rule["op"] != "and" {
		t.Fatalf("rule = %s", c.Rule)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	ctx, pool := emptyDB(t)
	if _, err := NewStore(pool).Get(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestStore_ReplaceDraft(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	id, err := s.CreateDraft(ctx, draft("brief", lc("tr", "one"), lc("en", "two")))
	if err != nil {
		t.Fatal(err)
	}
	d := draft("brief", lc("tr", "new"))
	d.Skipped = []string{"en"}
	if err := s.ReplaceDraft(ctx, id, d); err != nil {
		t.Fatal(err)
	}
	c, err := s.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Contents) != 1 || c.Contents[0].SMS != "new" {
		t.Fatalf("contents = %+v", c.Contents)
	}
	if len(c.SkippedL
… [5868 more chars truncated]
```

<details><summary>Result</summary>

```
# campaigncomposer/internal/campaign [campaigncomposer/internal/campaign.test]
internal/campaign/store_test.go:34:27: undefined: LanguageContent
internal/campaign/store_test.go:35:9: undefined: LanguageContent
internal/campaign/store_test.go:44:38: undefined: LanguageContent
internal/campaign/store_test.go:44:55: undefined: Draft
```

</details>

**🔧 Bash** _Implement campaign types and store, run tests_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend/internal/campaign && cat > campaign.go <<'EOF'
// Package campaign stores draft and approved campaigns and checks the edits
// of the marketer. It imports only content, segment and store from internal.
package campaign

import (
	"encoding/json"
	"errors"
	"time"

	"campaigncomposer/internal/content"
)

// Status is the state of a campaign.
type Status string

const (
	StatusDraft    Status = "draft"
	StatusApproved Status = "approved"
)

// LanguageContent is the copy of one language with the audience snapshot of
// the last save.
type LanguageContent struct {
	Language string `json:"language"`
	content.Content
	SMSCount       int64 `json:"sms_count"`
	EmailCount     int64 `json:"email_count"`
	FirstNameRunes int   `json:"first_name_runes"`
}

// Campaign is one stored campaign. Contents are sorted by language.
type Campaign struct {
	ID               int64             `json:"id"`
	Status           Status            `json:"status"`
	Brief            string            `json:"brief"`
	Rule             json.RawMessage   `json:"-"`
	Description      string            `json:"description"`
	SkippedLanguages []string          `json:"skipped_languages"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	ApprovedAt       *time.Time        `json:"approved_at"`
	Contents         []LanguageContent `json:"contents"`
}

// Summary is one row of the campaign list. It has no content text.
type Summary struct {
	ID          int64      `json:"id"`
	Status      Status     `json:"status"`
	Description string     `json:"description"`
	Languages   []string   `json:"languages"`
	CreatedAt   time.Time  `json:"created_at"`
	ApprovedAt  *time.Time `json:"approved_at"`
}

// Draft is the input of CreateDraft and ReplaceDraft.
type Draft struct {
	Brief, Description string
	Rule               json.RawMessage // canonical JSON from segment.Rule.MarshalJSON
	Skipped            []string
	Contents           []LanguageContent
}

var (
	ErrNotFound = errors.New("campaign: not found")
	ErrNotDraft = errors.New("campaign: not a draft")
	ErrConflict = errors.New("campaign: brief or rule differs, or not a draft")
)
EOF
cat > store.go <<'EOF'
package campaign

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// statementTimeout limits each query, as in the segment package.
const statementTimeout = "2s"

// Store reads and writes campaigns in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore returns a Store on pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) inTx(ctx context.Context, mode pgx.TxAccessMode, fn func(pgx.Tx) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: mode})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = '"+stat
… [7235 more chars truncated]
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/campaign	1.069s
```

</details>

**🔧 Bash** _Confirm store tests ran without skips_
```bash
TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55433/test?sslmode=disable' go test -count=1 -v ./internal/campaign/ 2>&1 | grep -cE "^--- PASS"; TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55433/test?sslmode=disable' go test -count=1 -v ./internal/campaign/ 2>&1 | grep -E "SKIP|FAIL" | head; echo "Task 2: complete (tests: go test ./internal/campaign → 11/11 pass with Postgres)" >> ../.superpowers/sdd/06-review-approve/progress.md
```

<details><summary>Result</summary>

```
11
```

</details>

**🔧 Bash** _Write Task 3 tests and confirm they fail_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend/internal/campaign && cat > check_test.go <<'EOF'
package campaign

import (
	"strings"
	"testing"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/segment"
)

func audOf(langs ...segment.LanguageCounts) map[string]segment.LanguageCounts {
	m := map[string]segment.LanguageCounts{}
	for _, l := range langs {
		m[l.Code] = l
	}
	return m
}

func lcount(code string, reachable, sms, email, maxName int64) segment.LanguageCounts {
	return segment.LanguageCounts{Code: code, Counts: segment.Counts{SMS: sms, Email: email, Reachable: reachable, Matched: reachable}, MaxFirstName: maxName}
}

const okTR = `{"language":"tr","sms":"  Merhaba {{first_name}}  ","email_subject":"Konu","email_body":"Gövde"}`
const okEN = `{"language":"en","sms":"Hi {{first_name}}","email_subject":"Subject","email_body":"Body"}`

func paths(errs []content.FieldError) []string {
	var p []string
	for _, e := range errs {
		p = append(p, e.Path)
	}
	return p
}

func hasPath(errs []content.FieldError, path string) bool {
	for _, e := range errs {
		if e.Path == path {
			return true
		}
	}
	return false
}

func TestCheckContents_Valid(t *testing.T) {
	aud := audOf(lcount("tr", 100, 80, 90, 12), lcount("en", 10, 8, 9, 6))
	got, errs := CheckContents([]byte(`{"contents":[`+okTR+`,`+okEN+`]}`), []string{"en", "tr"}, aud)
	if len(errs) > 0 {
		t.Fatalf("errors: %v", errs)
	}
	if len(got) != 2 || got[0].Language != "en" || got[1].Language != "tr" {
		t.Fatalf("got %+v", got)
	}
	tr := got[1]
	if tr.SMS != "Merhaba {{first_name}}" {
		t.Errorf("sms not trimmed: %q", tr.SMS)
	}
	if tr.SMSCount != 80 || tr.EmailCount != 90 || tr.FirstNameRunes != 12 {
		t.Errorf("counts = %+v", tr)
	}
}

func TestCheckContents(t *testing.T) {
	aud := audOf(lcount("tr", 100, 80, 90, 20), lcount("en", 10, 8, 9, 6))
	long := func(n int) string { return strings.Repeat("a", n) }
	tests := []struct {
		name, body string
		langs      []string
		want       string // a path that must be in the errors
	}{
		{"not an object", `[]`, []string{"tr"}, ""},
		{"no contents", `{}`, []string{"tr"}, "contents"},
		{"empty array", `{"contents":[]}`, []string{"tr"}, "contents"},
		{"missing draft language", `{"contents":[` + okTR + `]}`, []string{"tr", "en"}, "contents"},
		{"language not in draft", `{"contents":[` + okTR + `,` + okEN + `]}`, []string{"tr"}, "contents[1].language"},
		{"same language twice", `{"contents":[` + okTR + `,` + okTR + `]}`, []string{"tr"}, "contents[1].language"},
		{"unknown key", `{"contents":[{"language":"tr","sms":"a","email_subject":"b","email_body":"c","x":1}]}`, []string{"tr"}, "contents[0].x"},
		{"entry not an object", `{"contents":["x"]}`, []string{"tr"}, "contents[0]"},
		// 140 runes + {{first_name}} rendered with 20 runes = 160 passes; 141 fails.
		{"sms 161 runes", `{"contents":[{"language":"tr","sms":"{{first_name}}` + long(141) + `","email_subject":"b","email_body":"c"}]}`, []string{"tr"}, "contents[0].s
… [9358 more chars truncated]
```

<details><summary>Result</summary>

```
# campaigncomposer/internal/campaign [campaigncomposer/internal/campaign.test]
internal/campaign/service_test.go:21:68: undefined: Service
internal/campaign/service_test.go:28:10: undefined: Service
internal/campaign/service_test.go:43:57: undefined: Service
internal/campaign/service_test.go:62:56: undefined: Service
```

</details>

**🔧 Bash** _Implement CheckContents and Service, run tests_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend/internal/campaign && cat > check.go <<'EOF'
package campaign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/segment"
)

// InvalidError carries the validation errors of a save or an approval.
type InvalidError struct{ Errors []content.FieldError }

func (e *InvalidError) Error() string {
	parts := make([]string, 0, len(e.Errors))
	for _, fe := range e.Errors {
		parts = append(parts, fe.Path+": "+fe.Message)
	}
	return "campaign: invalid content: " + strings.Join(parts, "; ")
}

var entryKeys = map[string]bool{"language": true, "sms": true, "email_subject": true, "email_body": true}

// CheckContents decodes body ({"contents": [...]}) and checks it against the
// languages of the draft and the current audience. It returns the trimmed
// contents with the new counts and first_name_runes, sorted by language.
func CheckContents(body json.RawMessage, draftLangs []string, aud map[string]segment.LanguageCounts) ([]LanguageContent, []content.FieldError) {
	var errs []content.FieldError
	fail := func(path, format string, args ...any) {
		errs = append(errs, content.FieldError{Path: path, Message: fmt.Sprintf(format, args...)})
	}

	var top map[string]json.RawMessage
	if t := bytes.TrimSpace(body); len(t) == 0 || t[0] != '{' || json.Unmarshal(body, &top) != nil {
		fail("", "must be a JSON object")
		return nil, errs
	}
	var entries []json.RawMessage
	if raw, ok := top["contents"]; !ok || json.Unmarshal(raw, &entries) != nil || len(entries) == 0 {
		fail("contents", "must be an array with one entry for each language")
		return nil, errs
	}

	inDraft := map[string]bool{}
	for _, l := range draftLangs {
		inDraft[l] = true
	}
	seen := map[string]bool{}
	var out []LanguageContent
	for i, raw := range entries {
		base := fmt.Sprintf("contents[%d]", i)
		var m map[string]json.RawMessage
		if t := bytes.TrimSpace(raw); len(t) == 0 || t[0] != '{' || json.Unmarshal(raw, &m) != nil {
			fail(base, "must be a JSON object")
			continue
		}
		var unknown []string
		for k := range m {
			if !entryKeys[k] {
				unknown = append(unknown, k)
			}
		}
		sort.Strings(unknown)
		for _, k := range unknown {
			fail(base+"."+k, "unknown key")
		}

		var lang string
		if json.Unmarshal(m["language"], &lang) != nil || lang == "" {
			fail(base+".language", "must be a language code")
			continue
		}
		switch l, ok := aud[lang]; {
		case !inDraft[lang]:
			fail(base+".language", "%q is not a language of this draft", lang)
			continue
		case seen[lang]:
			fail(base+".language", "%q appears more than once", lang)
			continue
		case !ok || l.Reachable == 0:
			seen[lang] = true
			fail(base+".language", "no user of language %q can get a message now", lang)
			continue
		}
		seen[lang] = true
		l := aud[lang]

		text := map[string]json.RawMessage{}
		for _, k := range []string{"sms", "email_subject", "email_body"} {
			if v, ok := m[k]
… [4922 more chars truncated]
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/campaign	1.654s
```

</details>

**🔧 Bash** _Run campaign tests with race detector and vet_
```bash
TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55433/test?sslmode=disable' go test -count=1 -race -v ./internal/campaign/ 2>&1 | grep -E "^(--- |ok|FAIL)" | grep -v "PASS" ; TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55433/test?sslmode=disable' go test -count=1 -v ./internal/campaign/ 2>&1 | grep -cE "^\s*--- PASS"; go vet ./internal/campaign/
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/campaign	3.351s
36
```

</details>

**🔧 Bash** _Survey content handler tests and generator constructor_
```bash
echo "Task 3: complete (tests: go test -race ./internal/campaign → 36 pass incl. subtests, Postgres)" >> ../.superpowers/sdd/06-review-approve/progress.md; grep -n "^func \|ContentGenerator{" internal/httpapi/content_test.go | head -50; cat cmd/api/mockdefault.go | grep -n "func newContentGenerator" -A20
```

<details><summary>Result</summary>

```
31:func (f *langCounter) count(_ context.Context, _ segment.Where) (segment.Audience, error) {
49:func (f *langCounter) callCount() int {
55:func trLang() segment.LanguageCounts {
59:func enLang() segment.LanguageCounts {
63:func frozenClock() time.Time { return time.Unix(1_700_000_000, 0) }
65:func newGenerator(c llm.Client, f *langCounter) *ContentGenerator {
66:	return &ContentGenerator{
73:func serveContent(g *ContentGenerator, body string) *httptest.ResponseRecorder {
77:func serveContentCtx(ctx context.Context, g *ContentGenerator, body string) *httptest.ResponseRecorder {
87:func exampleRule() string {
95:func genBody(brief, rule string) string {
102:func contentStep(raw string) llm.MockStep {
111:func contentMock(f func(lang string, attempt int) llm.MockStep) *llm.Mock {
120:func defaultContent(lang string, _ int) llm.MockStep {
133:func (c *fakeStream) Generate(context.Context, llm.Request) (llm.Response, error) {
137:func (c *fakeStream) Stream(ctx context.Context, req llm.Request, on func(llm.StreamEvent)) (llm.Response, error) {
144:func (c *fakeStream) requests() []llm.Request {
160:func dataOf(t *testing.T, e sseEvent) langData {
170:func forLang(t *testing.T, evs []sseEvent, lang string) []sseEvent {
185:func names(evs []sseEvent) []string {
195:func streamEvents(t *testing.T, rec *httptest.ResponseRecorder) []sseEvent {
203:func TestGenerate_Example(t *testing.T) {
271:func TestGenerate_BadBody(t *testing.T) {
297:func TestGenerate_BadBrief(t *testing.T) {
311:func TestGenerate_BadRule(t *testing.T) {
332:func TestGenerate_EmptyAudience(t *testing.T) {
348:func TestGenerate_DBError(t *testing.T) {
359:func TestGenerate_RuleRoundTrip(t *testing.T) {
376:func TestGenerate_TwoLanguages(t *testing.T) {
423:func TestGenerate_Repair(t *testing.T) {
539:func TestGenerate_LLMErrors(t *testing.T) {
586:func TestGenerate_Throttle(t *testing.T) {
635:func TestGenerate_Reset(t *testing.T) {
679:func TestGenerate_Disconnect(t *testing.T) {
714:func TestGenerate_L
… [408 more chars truncated]
```

</details>

**🔧 Bash** _Read content test helpers and key tests_
```bash
sed -n 20,130p internal/httpapi/content_test.go; sed -n 150,270p internal/httpapi/content_test.go; sed -n 376,425p internal/httpapi/content_test.go; sed -n 679,714p internal/httpapi/content_test.go
```

<details><summary>Result</summary>

```
"campaigncomposer/internal/segment"
)

// langCounter is a fake AudienceCounter that returns fixed languages.
type langCounter struct {
	mu    sync.Mutex
	calls int
	langs []segment.LanguageCounts
	err   error
}

func (f *langCounter) count(_ context.Context, _ segment.Where) (segment.Audience, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil {
		return segment.Audience{}, f.err
	}
	a := segment.Audience{Languages: f.langs}
	for _, l := range f.langs {
		a.Email += l.Email
		a.SMS += l.SMS
		a.Reachable += l.Reachable
		a.Matched += l.Matched
	}
	a.NoConsent = a.Matched - a.Reachable
	return a, nil
}

func (f *langCounter) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func trLang() segment.LanguageCounts {
	return segment.LanguageCounts{Code: "tr", Counts: segment.Counts{Email: 290, SMS: 205, Reachable: 330, Matched: 412, NoConsent: 82}, MaxFirstName: 14}
}

func enLang() segment.LanguageCounts {
	return segment.LanguageCounts{Code: "en", Counts: segment.Counts{Email: 40, SMS: 30, Reachable: 45, Matched: 50, NoConsent: 5}, MaxFirstName: 9}
}

func frozenClock() time.Time { return time.Unix(1_700_000_000, 0) }

func newGenerator(c llm.Client, f *langCounter) *ContentGenerator {
	return &ContentGenerator{
		LLM: c, Allow: segment.Default(), Count: f.count,
		Now: time.Now, Clock: frozenClock,
		LLMTimeout: time.Second, DraftEvery: 100 * time.Millisecond,
	}
}

func serveContent(g *ContentGenerator, body string) *httptest.ResponseRecorder {
	return serveContentCtx(context.Background(), g, body)
}

func serveContentCtx(ctx context.Context, g *ContentGenerator, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	RegisterContent(mux, g)
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/content/generate", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// exampleRule is the rule object of llm.ExampleRuleJSON.
func exampleRule() string {
	var ex 
… [7918 more chars truncated]
```

</details>

**🔧 Bash** _Write Task 4 tests and confirm they fail_
```bash
cat > internal/httpapi/content_draft_test.go <<'EOF'
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/content"
	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/segment"
)

// fakeDrafts records the draft saves.
type fakeDrafts struct {
	mu       sync.Mutex
	created  []campaign.Draft
	replaced map[int64]campaign.Draft
	err      error
}

func (f *fakeDrafts) CreateDraft(_ context.Context, d campaign.Draft) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return 0, f.err
	}
	f.created = append(f.created, d)
	return 41, nil
}

func (f *fakeDrafts) ReplaceDraft(_ context.Context, id int64, d campaign.Draft) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	if f.replaced == nil {
		f.replaced = map[int64]campaign.Draft{}
	}
	f.replaced[id] = d
	return nil
}

func (f *fakeDrafts) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.created) + len(f.replaced)
}

func draftGenerator(c llm.Client, f *langCounter, d *fakeDrafts) *ContentGenerator {
	g := newGenerator(c, f)
	g.Drafts = d
	return g
}

type streamErr struct {
	Language *string `json:"language"`
	Code     string  `json:"code"`
	Message  string  `json:"message"`
}

func lastEvents(evs []sseEvent, n int) []string {
	var out []string
	for _, e := range evs[len(evs)-n:] {
		out = append(out, e.Event)
	}
	return out
}

func TestGenerate_SavesDraft(t *testing.T) {
	m := contentMock(defaultContent)
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	d := &fakeDrafts{}
	evs := streamEvents(t, serveContent(draftGenerator(m, f, d), genBody("  "+exampleBrief+"  ", exampleRule())))
	if got := strings.Join(lastEvents(evs, 3), ","); got != "final,campaign,done" {
		t.Fatalf("last events %s", got)
	}
	var ce struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(evs[len(evs)-2].Data, &ce); err != nil || ce.ID != 41 || ce.Status != "draft" {
		t.Fatalf("campaign event %s", evs[len(evs)-2].Data)
	}
	if len(d.created) != 1 {
		t.Fatalf("created %d drafts", len(d.created))
	}
	got := d.created[0]
	brief, _ := segment.SanitizeBrief(exampleBrief)
	rule, _ := segment.Default().ParseRule([]byte(exampleRule()))
	wantRule, _ := rule.MarshalJSON()
	if got.Brief != brief || string(got.Rule) != string(wantRule) || got.Description != segment.Default().Describe(rule) {
		t.Errorf("draft head %+v", got)
	}
	if len(got.Skipped) != 0 {
		t.Errorf("skipped %v", got.Skipped)
	}
	want, _, _ := content.Validate(content.MockContent("tr"), 14)
	if len(got.Contents) != 1 {
		t.Fatalf("contents %+v", got.Contents)
	}
	c := got.Contents[0]
	if c.Language != "tr" || c.Content != want || c.SMSCount != 205 || c.EmailCount != 290 || c.FirstNameRunes != 14 {
		t.Errorf("content %+v", c)
	}
}

func TestGenerate_DraftSkipsFailedLanguage(t *testing.T) {
	m := contentMock(func(lang string, atte
… [4733 more chars truncated]
```

<details><summary>Result</summary>

```
# campaigncomposer/internal/httpapi [campaigncomposer/internal/httpapi.test]
internal/httpapi/content_draft_test.go:58:4: g.Drafts undefined (type *ContentGenerator has no field or method Drafts)
FAIL	campaigncomposer/internal/httpapi [build failed]
FAIL
```

</details>

**🔧 Bash** _Implement draft save in content endpoint and run tests_
```bash
python3 - <<'EOF'
p='internal/httpapi/content.go'
s=open(p).read()
def rep(a,b):
    global s
    assert a in s, a
    s=s.replace(a,b,1)
rep('''	"campaigncomposer/internal/content"''','''	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/content"''')
rep('''	DraftEvery time.Duration    // minimum time between two drafts of a language
}''','''	DraftEvery time.Duration    // minimum time between two drafts of a language
	Drafts     DraftSaver       // stores the draft at the end; nil saves nothing
}

// DraftSaver stores the draft at the end of a content stream.
type DraftSaver interface {
	CreateDraft(ctx context.Context, d campaign.Draft) (int64, error)
	ReplaceDraft(ctx context.Context, id int64, d campaign.Draft) error
}

// draftSaveTimeout limits the draft save at the end of the stream.
const draftSaveTimeout = 2 * time.Second''')
rep('''type languageErrorEvent struct {''','''type campaignEvent struct {
	ID     int64           `json:"id"`
	Status campaign.Status `json:"status"`
}

// streamErrorEvent is an error event with no language.
type streamErrorEvent struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// langResult is the outcome of one language. final is nil when it failed.
type langResult struct {
	lang  segment.LanguageCounts
	final *finalEvent
}

type languageErrorEvent struct {''')
rep('''		Brief string          `json:"brief"`
		Rule  json.RawMessage `json:"rule"`
	}
	if err := decodeBody(w, r, &in); err != nil || in.Brief == "" || isAbsentJSON(in.Rule) {''','''		Brief      string          `json:"brief"`
		Rule       json.RawMessage `json:"rule"`
		CampaignID *int64          `json:"campaign_id"`
	}
	if err := decodeBody(w, r, &in); err != nil || in.Brief == "" || isAbsentJSON(in.Rule) ||
		(in.CampaignID != nil && *in.CampaignID <= 0) {''')
rep('''	var wg sync.WaitGroup
	for _, l := range langs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.runLanguage(ctx, sse, l, description, brief)
		}()
	}
	wg.Wait()
	if ctx.Err() != nil {
		slog.InfoContext(ctx, "content generate: client disconnected", "languages", len(langs))
		return
	}
	_ = sse.send("done", struct{}{})
}''','''	results := make([]langResult, len(langs))
	var wg sync.WaitGroup
	for i, l := range langs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = langResult{lang: l, final: g.runLanguage(ctx, sse, l, description, brief)}
		}()
	}
	wg.Wait()
	if ctx.Err() != nil {
		// The marketer did not see the result, so no draft is saved.
		slog.InfoContext(ctx, "content generate: client disconnected", "languages", len(langs))
		return
	}
	ruleJSON, err := rule.MarshalJSON()
	if err == nil {
		g.saveDraft(ctx, sse, in.CampaignID, campaign.Draft{Brief: brief, Description: description, Rule: ruleJSON}, results)
	} else {
		slog.ErrorContext(ctx, "content generate: marshal rule", "err", err)
		_ = sse.send("error", streamErrorEvent{Code: "draft_not_saved", Message: "The draft could not be saved. Try again."})
	}
	_ = sse.send("done
… [2993 more chars truncated]
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/httpapi	3.640s
```

</details>

**🔧 Bash** _Add Postgres draft test and run it_
```bash
grep -ln "storetest" internal/httpapi/*_test.go; cat >> internal/httpapi/content_draft_test.go <<'EOF'

func TestGenerate_DraftInDatabase(t *testing.T) {
	url := storetest.NewDB(t)
	ctx := context.Background()
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := seed.Load(ctx, pool, seed.Options{Random: 42, Now: time.Now(), Users: 500}, false); err != nil {
		t.Fatal(err)
	}
	count := func(ctx context.Context, w segment.Where) (segment.Audience, error) { return segment.CountAudience(ctx, pool, w) }
	st := campaign.NewStore(pool)
	g := &ContentGenerator{
		LLM: contentMock(defaultContent), Allow: segment.Default(), Count: count,
		Now: time.Now, Clock: frozenClock, LLMTimeout: time.Second, DraftEvery: 100 * time.Millisecond,
		Drafts: st,
	}
	rule := `{"op":"and","conditions":[{"field":"city","op":"eq","value":"Istanbul"}]}`
	evs := streamEvents(t, serveContent(g, genBody(exampleBrief, rule)))
	var ce struct {
		ID int64 `json:"id"`
	}
	for _, e := range evs {
		if e.Event == "campaign" {
			_ = json.Unmarshal(e.Data, &ce)
		}
	}
	if ce.ID == 0 {
		t.Fatalf("no campaign event: %v", names(evs))
	}
	c, err := st.Get(ctx, ce.ID)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := segment.Default().ParseRule([]byte(rule))
	aud, err := count(ctx, segment.Default().Compile(r, time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != campaign.StatusDraft || len(c.Contents) != len(aud.Languages) {
		t.Fatalf("campaign %+v, audience languages %d", c, len(aud.Languages))
	}
	for _, l := range aud.Languages {
		found := false
		for _, lc := range c.Contents {
			if lc.Language == l.Code {
				found = true
				if lc.SMSCount != l.SMS || lc.EmailCount != l.Email || lc.FirstNameRunes != int(l.MaxFirstName) {
					t.Errorf("%s: stored %+v, audience %+v", l.Code, lc, l)
				}
			}
		}
		if !found {
			t.Errorf("no content row for %s", l.Code)
		}
	}
	// A regenerate with the same brief and rule replaces the same draft.
	b, _ := json.Marshal(exampleBrief)
	evs = streamEvents(t, serveContent(g, fmt.Sprintf(`{"brief":%s,"rule":%s,"campaign_id":%d}`, b, rule, ce.ID)))
	if !strings.Contains(string(evs[len(evs)-2].Data), fmt.Sprintf(`"id":%d`, ce.ID)) {
		t.Fatalf("regenerate: %s %s", evs[len(evs)-2].Event, evs[len(evs)-2].Data)
	}
	list, _ := st.List(ctx, 10)
	if len(list) != 1 {
		t.Fatalf("campaigns after regenerate: %d", len(list))
	}
}
EOF
sed -i '' 's#"campaigncomposer/internal/segment"\n)#X#' internal/httpapi/content_draft_test.go; python3 - <<'EOF'
p='internal/httpapi/content_draft_test.go'
s=open(p).read()
s=s.replace('''	"campaigncomposer/internal/segment"
)''','''	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/segment"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)''',1).replace('''	"time"
''','''	"time"

	"github.com/jackc/pgx/v5/pgxpool"
''',1)
open(p,'w').write(s)
EOF
TEST_DATABASE_URL='po
… [134 more chars truncated]
```

<details><summary>Result</summary>

```
2026/11/10 14:53:09 INFO content generate language=tr attempts=1 outcome=final first_name_runes=7 sms_runes=76 email_subject_runes=44 email_body_runes=148
2026/11/10 14:53:09 INFO content generate: draft saved campaign_id=1 replaced=true languages=1 skipped=[]
--- PASS: TestGenerate_DraftInDatabase (0.13s)
PASS
ok  	campaigncomposer/internal/httpapi	0.497s
```

</details>

**🔧 Bash** _Ledger Task 4 and check httpapi test helpers_
```bash
cat >> ../.superpowers/sdd/06-review-approve/progress.md <<'EOF'
Task 4: Ruling: a nil ContentGenerator.Drafts saves no draft and sends no campaign event, instead of updating every older test with a fake saver — keeps the slice 05 tests unchanged — cost if wrong: cmd/api must wire Drafts (Task 5 does; the e2e check proves it).
Task 4: Ruling: campaign_id that is not a positive integer (0, -1, "7", 1.5) returns 400 invalid_request — spec says "0 or less" and the decoder rejects the others — cost if wrong: none.
Task 4: complete (tests: go test -race ./internal/httpapi + DraftInDatabase with Postgres → pass)
EOF
sed -n 1,60p internal/httpapi/segments_test.go | grep -n "func serve\|httptest" | head; grep -n "func Test" internal/httpapi/cors_test.go
```

<details><summary>Result</summary>

```
10:	"net/http/httptest"
45:func post(h http.Handler, body string) *httptest.ResponseRecorder {
46:	req := httptest.NewRequest(http.MethodPost, "/api/segments/preview", strings.NewReader(body))
47:	rec := httptest.NewRecorder()
16:func TestCORS_KnownOrigin(t *testing.T) {
39:func TestCORS_UnknownOrigin(t *testing.T) {
```

</details>

**🔧 Bash** _Read CORS test_
```bash
sed -n 1,40p internal/httpapi/cors_test.go
```

<details><summary>Result</summary>

```
package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func corsHandler() http.Handler {
	return CORS([]string{"http://localhost:3000"}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func TestCORS_KnownOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/segments/preview", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	corsHandler().ServeHTTP(rec, req)
	h := rec.Header()
	if rec.Code != 204 || h.Get("Access-Control-Allow-Origin") != "http://localhost:3000" ||
		!strings.Contains(h.Get("Access-Control-Allow-Methods"), "POST") ||
		!strings.Contains(h.Get("Access-Control-Allow-Headers"), "Content-Type") ||
		h.Get("Access-Control-Max-Age") != "600" || h.Get("Vary") != "Origin" {
		t.Errorf("%d %v", rec.Code, h)
	}

	req = httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec = httptest.NewRecorder()
	corsHandler().ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("simple request: %d %v", rec.Code, rec.Header())
	}
}

func TestCORS_UnknownOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/segments/preview", nil)
```

</details>

## 🤖 Claude  <sub>2026-11-10 11:53:23</sub>

Task 5: endpoint tests.

**🔧 Bash** _Write Task 5 handler tests and confirm they fail_
```bash
cat >> internal/httpapi/cors_test.go <<'EOF'

func TestCORS_AllowsPut(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/campaigns/1/contents", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	rec := httptest.NewRecorder()
	corsHandler().ServeHTTP(rec, req)
	if !strings.Contains(rec.Header().Get("Access-Control-Allow-Methods"), "PUT") {
		t.Errorf("methods %q", rec.Header().Get("Access-Control-Allow-Methods"))
	}
}
EOF
cat > internal/httpapi/campaigns_test.go <<'EOF'
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/content"
)

// fakeCampaigns is a CampaignService with fixed answers.
type fakeCampaigns struct {
	c       campaign.Campaign
	list    []campaign.Summary
	sample  []campaign.SampleNames
	err     error
	repeat  bool
	gotID   int64
	gotBody json.RawMessage
	calls   int
}

func (f *fakeCampaigns) Get(_ context.Context, id int64) (campaign.Campaign, error) {
	f.gotID, f.calls = id, f.calls+1
	return f.c, f.err
}
func (f *fakeCampaigns) List(context.Context) ([]campaign.Summary, error) {
	f.calls++
	return f.list, f.err
}
func (f *fakeCampaigns) Save(_ context.Context, id int64, b json.RawMessage) (campaign.Campaign, error) {
	f.gotID, f.gotBody, f.calls = id, b, f.calls+1
	return f.c, f.err
}
func (f *fakeCampaigns) Approve(_ context.Context, id int64, b json.RawMessage) (campaign.Campaign, bool, error) {
	f.gotID, f.gotBody, f.calls = id, b, f.calls+1
	return f.c, f.repeat, f.err
}
func (f *fakeCampaigns) SampleNames(_ context.Context, id int64) ([]campaign.SampleNames, error) {
	f.gotID, f.calls = id, f.calls+1
	return f.sample, f.err
}

func sampleCampaign() campaign.Campaign {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	return campaign.Campaign{
		ID: 12, Status: campaign.StatusDraft, Brief: "b", Rule: json.RawMessage(`{"op":"and","conditions":[]}`),
		Description: "d", SkippedLanguages: []string{"en"}, CreatedAt: now, UpdatedAt: now,
		Contents: []campaign.LanguageContent{{Language: "tr", Content: content.Content{SMS: "s", EmailSubject: "e", EmailBody: "b"},
			SMSCount: 980, EmailCount: 1212, FirstNameRunes: 14}},
	}
}

func serveCampaign(f *fakeCampaigns, method, path, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	RegisterCampaigns(mux, f)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func errCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var b errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}
	return b.Error.Code
}

const okBody = `{"contents":[{"language":"tr","sms":"s","email_subject":"e","email_body":"b"}]}`

func TestGetCampaign_Shape(t *testing.T) {
	f := &fakeCampaigns{c: sampleCampaig
… [5703 more chars truncated]
```

<details><summary>Result</summary>

```
# campaigncomposer/internal/httpapi [campaigncomposer/internal/httpapi.test]
internal/httpapi/campaigns_test.go:62:2: undefined: RegisterCampaigns
FAIL	campaigncomposer/internal/httpapi [build failed]
FAIL
```

</details>

**🔧 Bash** _Implement campaign handlers, body limit and CORS change_
```bash
python3 - <<'EOF'
p='internal/httpapi/body.go'
s=open(p).read()
s=s.replace('''func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))''','''func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	return decodeBodyLimit(w, r, v, maxBodyBytes)
}

// decodeBodyLimit is decodeBody with a given byte limit.
func decodeBodyLimit(w http.ResponseWriter, r *http.Request, v any, limit int64) error {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))''')
open(p,'w').write(s)
p='internal/httpapi/cors.go'
s=open(p).read()
s=s.replace('"GET, POST, OPTIONS"','"GET, POST, PUT, OPTIONS"')
open(p,'w').write(s)
EOF
cat > internal/httpapi/campaigns.go <<'EOF'
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"campaigncomposer/internal/campaign"
)

// maxCampaignBodyBytes limits the body of a save or an approval.
const maxCampaignBodyBytes = 64 * 1024

// CampaignService is the part of campaign.Service that the handlers use.
type CampaignService interface {
	Get(ctx context.Context, id int64) (campaign.Campaign, error)
	List(ctx context.Context) ([]campaign.Summary, error)
	Save(ctx context.Context, id int64, body json.RawMessage) (campaign.Campaign, error)
	Approve(ctx context.Context, id int64, body json.RawMessage) (campaign.Campaign, bool, error)
	SampleNames(ctx context.Context, id int64) ([]campaign.SampleNames, error)
}

type campaignHandlers struct{ svc CampaignService }

// RegisterCampaigns adds the campaign routes.
func RegisterCampaigns(mux *http.ServeMux, svc CampaignService) {
	h := campaignHandlers{svc: svc}
	mux.HandleFunc("GET /api/campaigns", h.list)
	mux.HandleFunc("GET /api/campaigns/{id}", h.get)
	mux.HandleFunc("PUT /api/campaigns/{id}/contents", h.save)
	mux.HandleFunc("POST /api/campaigns/{id}/approve", h.approve)
	mux.HandleFunc("GET /api/campaigns/{id}/sample", h.sample)
}

// campaignID reads the id path value. It writes 400 and returns false when the
// id is not a positive integer.
func campaignID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "The campaign id must be a positive integer.", nil)
		return 0, false
	}
	return id, true
}

// contentsBody reads a save or approval body. It writes 400 and returns false
// when the body is not a JSON object with a contents key.
func contentsBody(w http.ResponseWriter, r *http.Request) (json.RawMessage, bool) {
	var raw json.RawMessage
	var m map[string]json.RawMessage
	if err := decodeBodyLimit(w, r, &raw, maxCampaignBodyBytes); err != nil ||
		json.Unmarshal(raw, &m) != nil || m == nil || isAbsentJSON(m["contents"]) {
		writeError(w, http.StatusBadRequest, "invalid_request", "The request body must be JSON with contents, of at most 64 KiB.", nil)
		return nil, false
	}
	return raw, true

… [3038 more chars truncated]
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/httpapi	2.423s
```

</details>

**🔧 Bash** _Wire campaign service in cmd/api and run make verify_
```bash
python3 - <<'EOF'
p='cmd/api/main.go'
s=open(p).read()
s=s.replace('''	"campaigncomposer/internal/config"''','''	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/config"''')
s=s.replace('''	httpapi.RegisterContent(mux, newContentGenerator(llmClient, segment.Default(), count))''','''	campaigns := campaign.NewStore(pool)
	gen := newContentGenerator(llmClient, segment.Default(), count)
	gen.Drafts = campaigns
	httpapi.RegisterContent(mux, gen)
	httpapi.RegisterCampaigns(mux, &campaign.Service{
		Store: campaigns,
		Allow: segment.Default(),
		Count: campaign.AudienceCounter(count),
		Sample: func(ctx context.Context, w segment.Where, langs []string, n int) (map[string][]string, error) {
			return segment.SampleFirstNames(ctx, pool, w, langs, n)
		},
		Now: time.Now,
	})''')
open(p,'w').write(s)
EOF
go build ./... && go vet ./... && gofmt -l . ; cd .. && make verify > .superpowers/sdd/06-review-approve/verify5.log 2>&1; echo rc=$?; tail -15 .superpowers/sdd/06-review-approve/verify5.log
```

<details><summary>Result</summary>

```
internal/httpapi/content_draft_test.go
rc=2
cd backend && go vet ./...
gofmt needed:
internal/httpapi/content_draft_test.go
make: *** [lint] Error 1
```

</details>

**🔧 Bash** _Format Go files and rerun make verify_
```bash
gofmt -w backend/internal && make verify > .superpowers/sdd/06-review-approve/verify5.log 2>&1; echo rc=$?; grep -E "^(ok|FAIL|---)|Test Files|Tests " .superpowers/sdd/06-review-approve/verify5.log | head -30
```

<details><summary>Result</summary>

```
rc=0
ok  	campaigncomposer/cmd/api	0.208s
ok  	campaigncomposer/internal/campaign	2.005s
ok  	campaigncomposer/internal/config	(cached)
ok  	campaigncomposer/internal/content	(cached)
ok  	campaigncomposer/internal/httpapi	2.488s
ok  	campaigncomposer/internal/llm	(cached)
ok  	campaigncomposer/internal/mockprovider	(cached)
ok  	campaigncomposer/internal/seed	1.338s
ok  	campaigncomposer/internal/segment	1.504s
ok  	campaigncomposer/internal/store	0.797s
 Test Files  13 passed (13)
      Tests  90 passed (90)
```

</details>

**🔧 Bash** _Ledger Task 5 and read frontend lib test style_
```bash
echo "Task 5: complete (tests: make verify → exit 0; Go all ok, frontend 90/90)" >> .superpowers/sdd/06-review-approve/progress.md; ls frontend/src/lib/__tests__; sed -n 1,40p frontend/src/lib/__tests__/segments.test.ts; sed -n 1,30p frontend/src/lib/__tests__/content.test.ts; sed -n 95,140p frontend/src/lib/content.ts
```

<details><summary>Result</summary>

```
api-client.test.ts
content.test.ts
segments.test.ts
import { FetchError } from 'ofetch'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { apiFetch } from '@/lib/api-client'
import { apiErrorMessage, previewSegment } from '@/lib/segments'

vi.mock('@/lib/api-client', () => ({ apiFetch: vi.fn() }))

describe('segments client', () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset()
  })

  it('previewSegment uses a 60s timeout', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ description: 'x' })
    await previewSegment('hello')
    expect(apiFetch).toHaveBeenCalledWith('/segments/preview', { method: 'POST', body: { brief: 'hello' }, timeout: 60000 })
  })

  it('apiErrorMessage reads error.message from a FetchError body', () => {
    const err = new FetchError('boom')
    err.data = { error: { code: 'llm_unavailable', message: 'The model is down.' } }
    expect(apiErrorMessage(err)).toBe('The model is down.')
  })

  it('apiErrorMessage falls back for an Error without data', () => {
    expect(apiErrorMessage(new Error('x'))).toBe('Something went wrong. Try again.')
    expect(apiErrorMessage(null)).toBe('Something went wrong. Try again.')
  })
})
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ContentEvent } from '@/lib/content'

import { API_BASE_URL } from '@/constants/app-config'
import { ContentRequestError, generateContent, parseSSE } from '@/lib/content'

const enc = new TextEncoder()

function sseResponse(chunks: string[], status = 200): Response {
  const stream = new ReadableStream<Uint8Array>({
    start(c) {
      for (const ch of chunks) c.enqueue(enc.encode(ch))
      c.close()
    },
  })
  return new Response(stream, { status, headers: { 'Content-Type': 'text/event-stream' } })
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('parseSSE', () => {
  it('splits events across chunks', () => {
    const first = parseSSE('event: draft\ndata: {"language":"tr","con', '')
    expect
… [1437 more chars truncated]
```

</details>

**🔧 Bash** _Write Task 6 frontend lib tests and confirm they fail_
```bash
cd ~/Desktop/projects/case-studies/insider_one/frontend && cat > src/lib/__tests__/limits.test.ts <<'EOF'
import { describe, expect, it } from 'vitest'

import { checkContent, LIMITS, renderedLength, runeLength } from '@/lib/limits'

const ok = { sms: 'Hi {{first_name}}', email_subject: 'Subject', email_body: 'Body' }

describe('limits', () => {
  it('has the case limits', () => {
    expect(LIMITS).toEqual({ sms: 160, email_subject: 80, email_body: 2000 })
  })

  it('counts code points, not UTF-16 units', () => {
    expect(runeLength('ş')).toBe(1)
    expect(runeLength('😀')).toBe(1)
  })

  it('renders with the longest name (slice 05 FR-9)', () => {
    const v = '{{first_name}}'
    expect(renderedLength(`${v}${'a'.repeat(140)}`, 20)).toBe(160)
    expect(checkContent({ ...ok, sms: `${v}${'a'.repeat(140)}` }, 20).sms).toBeUndefined()
    expect(checkContent({ ...ok, sms: `${v}${'a'.repeat(141)}` }, 20).sms).toBe('Too long after the name is filled in.')
    expect(checkContent({ ...ok, email_subject: `${v}${'b'.repeat(60)}` }, 20).email_subject).toBeUndefined()
    expect(checkContent({ ...ok, email_subject: `${v}${'b'.repeat(61)}` }, 20).email_subject).toBe('Too long after the name is filled in.')
    expect(checkContent({ ...ok, sms: `${v}${'ş'.repeat(140)}` }, 20).sms).toBeUndefined()
  })

  it('gives 160 for 146 runes plus one variable at 14 (FR-30)', () => {
    expect(renderedLength(`{{first_name}}${'a'.repeat(146)}`, 14)).toBe(160)
    expect(renderedLength(`{{first_name}}${'a'.repeat(147)}`, 14)).toBe(161)
  })

  it('accepts valid content', () => {
    expect(checkContent(ok, 14)).toEqual({})
  })

  it.each([
    ['empty', { ...ok, sms: '   ' }, 'sms', 'Write some text.'],
    ['wrong variable', { ...ok, sms: 'Hi {{last_name}}' }, 'sms', 'Only {{first_name}} is allowed.'],
    ['single brace', { ...ok, email_body: 'Hi {first_name}' }, 'email_body', 'Only {{first_name}} is allowed.'],
    ['spaced variable', { ...ok, sms: 'Hi {{ first_name }}' }, 'sms', 'Only {{first_name}} is allowed.'],
    ['subject newline', { ...ok, email_subject: 'a\nb' }, 'email_subject', 'The subject must be one line.'],
    ['html', { ...ok, email_body: '<b>x</b>' }, 'email_body', 'HTML is not allowed.'],
    ['control char', { ...ok, sms: 'a\u0007b' }, 'sms', 'Remove the special characters.'],
    ['newline in sms', { ...ok, sms: 'a\nb' }, 'sms', 'Remove the special characters.'],
  ] as const)('rejects %s', (_name, c, field, message) => {
    expect(checkContent(c, 14)[field]).toBe(message)
  })

  it('allows newline and tab in the body and a lone < in text', () => {
    expect(checkContent({ ...ok, email_body: 'a\n\tb < 3' }, 14)).toEqual({})
  })
})
EOF
cat > src/lib/__tests__/campaigns.test.ts <<'EOF'
import { FetchError } from 'ofetch'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { apiFetch } from '@/lib/api-client'
import { approveCampaign, CampaignRequestError, getCampaign, getSampleNames, listCampaigns, saveCon
… [3672 more chars truncated]
```

<details><summary>Result</summary>

```
121|     await generateContent('b', { a: 1 }, () => {}, ctl.signal, 7)
    122|     expect(JSON.parse(fetchMock.mock.calls[0]![1].body)).toEqual({ bri…
    123|     expect(JSON.parse(fetchMock.mock.calls[1]![1].body)).toEqual({ bri…
       |                                                          ^
    124|   })
    125| })

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[4/4]⎯


 Test Files  3 failed | 2 passed (5)
      Tests  2 failed | 17 passed (19)
   Start at  14:55:21
   Duration  234ms (import 49%, tests 27%, transform 21%, worker 4%)
```

</details>

**🔧 Bash** _Implement limits, campaigns client and content.ts changes_
```bash
cat > src/lib/limits.ts <<'EOF'
import type { Content } from '@/lib/content'

// The same limits and checks as the backend validator (internal/content).
// The backend decides; this check only helps the marketer while typing.
export const LIMITS = { sms: 160, email_subject: 80, email_body: 2000 } as const

const VAR = '{{first_name}}'
const TAG = /<[a-z/!]/i

export function runeLength(s: string): number {
  return [...s].length
}

// The length after each {{first_name}} is replaced by a name of firstNameRunes runes.
export function renderedLength(text: string, firstNameRunes: number): number {
  const vars = text.split(VAR).length - 1
  return runeLength(text) - vars * runeLength(VAR) + vars * firstNameRunes
}

function hasControl(s: string, allowBody: boolean): boolean {
  for (const ch of s) {
    const c = ch.codePointAt(0)!
    if (allowBody && (ch === '\n' || ch === '\t'))
      continue
    if (c < 0x20 || (c >= 0x7F && c <= 0x9F))
      return true
  }
  return false
}

export function checkContent(c: Content, firstNameRunes: number): Partial<Record<keyof Content, string>> {
  const out: Partial<Record<keyof Content, string>> = {}
  for (const key of ['sms', 'email_subject', 'email_body'] as const) {
    const text = c[key].trim()
    if (text === '') {
      out[key] = 'Write some text.'
      continue
    }
    const rest = text.split(VAR).join('')
    if (rest.includes('{') || rest.includes('}'))
      out[key] = 'Only {{first_name}} is allowed.'
    else if (key === 'email_subject' && /[\r\n]/.test(text))
      out[key] = 'The subject must be one line.'
    else if (hasControl(text, key === 'email_body'))
      out[key] = 'Remove the special characters.'
    else if (TAG.test(text))
      out[key] = 'HTML is not allowed.'
    else if (renderedLength(text, firstNameRunes) > LIMITS[key])
      out[key] = 'Too long after the name is filled in.'
  }
  return out
}
EOF
cat > src/lib/campaigns.ts <<'EOF'
import type { Content } from '@/lib/content'

import { apiFetch } from '@/lib/api-client'

export type CampaignStatus = 'draft' | 'approved'
export interface LanguageContent extends Content { language: string, sms_count: number, email_count: number, first_name_runes: number }
export interface Campaign {
  id: number
  status: CampaignStatus
  brief: string
  description: string
  skipped_languages: string[]
  created_at: string
  updated_at: string
  approved_at: string | null
  contents: LanguageContent[]
}
export interface CampaignSummary { id: number, status: CampaignStatus, description: string, languages: string[], created_at: string, approved_at: string | null }
export interface FieldError { path: string, message: string }
export type ContentInput = Content & { language: string }

export const LANGUAGE_NAMES: Record<string, string> = { tr: 'Turkish', en: 'English' }

export function languageName(code: string): string {
  return LANGUAGE_NAMES[code] ?? code.toUpperCase()
}

const FALLBACK_MESSAGE = 'Something went wrong. Try again.'

ex
… [3243 more chars truncated]
```

<details><summary>Result</summary>

```
Test Files  1 failed | 4 passed (5)
      Tests  2 failed | 38 passed (40)
   Start at  14:55:44
   Duration  213ms (import 47%, transform 26%, tests 23%, worker 3%)

src/components/campaign/content-cards.vue(61,31): error TS2339: Property 'language' does not exist on type '{ language: string; content: Content; } | { language: string; errors?: { path: string; message: string; }[] | undefined; } | { language: string; content: Content; limits: ContentLimits; } | { ...; } | { ...; }'.
  Property 'language' does not exist on type '{ id: number; status: CampaignStatus; }'.
```

</details>

**🔧 Bash** _Show failing frontend lib tests_
```bash
pnpm vitest run src/lib 2>&1 | grep -E "FAIL|✗|×|AssertionError|Expected|Received" | head -20
```

<details><summary>Result</summary>

```
× rejects with CampaignRequestError and details on 422 2ms
     × falls back on a network error 0ms
 FAIL  src/lib/__tests__/campaigns.test.ts > campaigns client > rejects with CampaignRequestError and details on 422
 FAIL  src/lib/__tests__/campaigns.test.ts > campaigns client > falls back on a network error
```

</details>

**🔧 Bash** _Show details of failing campaigns tests_
```bash
pnpm vitest run src/lib/__tests__/campaigns.test.ts 2>&1 | grep -A12 "rejects with" | head -40
```

<details><summary>Result</summary>

```
× rejects with CampaignRequestError and details on 422 1ms
     × falls back on a network error 0ms

⎯⎯⎯⎯⎯⎯⎯ Failed Tests 2 ⎯⎯⎯⎯⎯⎯⎯

 FAIL  src/lib/__tests__/campaigns.test.ts > campaigns client > rejects with CampaignRequestError and details on 422
FetchError: 422
 ❯ src/lib/__tests__/campaigns.test.ts:45:17
     43|
     44|   it('rejects with CampaignRequestError and details on 422', async () …
     45|     const err = new FetchError('422')
       |                 ^
     46|     err.data = { error: { code: 'content_invalid', message: 'Bad.', de…
     47|     err.statusCode = 422

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[1/2]⎯

 FAIL  src/lib/__tests__/campaigns.test.ts > campaigns client > falls back on a network error
Error: net
 ❯ src/lib/__tests__/campaigns.test.ts:58:43
     56|
     57|   it('falls back on a network error', async () => {
```

</details>

**🔧 Bash** _Debug campaigns test failure output_
```bash
pnpm vitest run src/lib/__tests__/campaigns.test.ts 2>&1 | sed -n '/falls back on a network error$/,/⎯⎯⎯/p' | head -30; pnpm vitest run src/lib/__tests__/campaigns.test.ts 2>&1 | grep -i "unhandled\|serialized" | head
```

<details><summary>Result</summary>

```
FAIL  src/lib/__tests__/campaigns.test.ts > campaigns client > falls back on a network error
Error: net
 ❯ src/lib/__tests__/campaigns.test.ts:58:43
     56|
     57|   it('falls back on a network error', async () => {
     58|     vi.mocked(apiFetch).mockRejectedValue(new Error('net'))
       |                                           ^
     59|     const got = await getCampaign(1).catch(e => e)
     60|     expect(got).toBeInstanceOf(CampaignRequestError)

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[2/2]⎯
```

</details>

**🔧 Bash** _Debug error wrapping with a temporary test_
```bash
cat > src/lib/__tests__/tmp.test.ts <<'EOF'
import { it, vi } from 'vitest'
import { apiFetch } from '@/lib/api-client'
import { getCampaign } from '@/lib/campaigns'
vi.mock('@/lib/api-client', () => ({ apiFetch: vi.fn() }))
it('x', async () => {
  vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
  try { await getCampaign(1); console.log('no throw') } catch (e) { console.log('caught', (e as Error).name, (e as Error).message) }
})
EOF
pnpm vitest run src/lib/__tests__/tmp.test.ts 2>&1 | grep -E "caught|no throw|✓|×"; rm src/lib/__tests__/tmp.test.ts
```

<details><summary>Result</summary>

```
(Bash completed with no output)
```

</details>

**🔧 Bash** _Show full output of temporary debug test_
```bash
cat > src/lib/__tests__/tmp.test.ts <<'EOF'
import { it, vi } from 'vitest'
import { apiFetch } from '@/lib/api-client'
import { getCampaign } from '@/lib/campaigns'
vi.mock('@/lib/api-client', () => ({ apiFetch: vi.fn() }))
it('x', async () => {
  vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
  try { await getCampaign(1); console.log('no throw') } catch (e) { console.log('caught', (e as Error).name, (e as Error).message) }
})
EOF
pnpm vitest run src/lib/__tests__/tmp.test.ts 2>&1 | tail -20; rm src/lib/__tests__/tmp.test.ts
```

<details><summary>Result</summary>

```
RUN  v5.0.5 ~/Desktop/projects/case-studies/insider_one/frontend


 Test Files  1 passed (1)
      Tests  1 passed (1)
   Start at  14:56:16
   Duration  60ms (import 51%, transform 21%, tests 14%, worker 14%)
```

</details>

**🔧 Bash** _Use lazy rejections in tests and rerun_
```bash
sed -i '' 's/vi.mocked(apiFetch).mockRejectedValue(err)/vi.mocked(apiFetch).mockImplementation(() => Promise.reject(err))/; s/vi.mocked(apiFetch).mockRejectedValue(new Error(.net.))/vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error("net")))/' src/lib/__tests__/campaigns.test.ts && pnpm vitest run src/lib 2>&1 | tail -4
```

<details><summary>Result</summary>

```
Tests  2 failed | 38 passed (40)
   Start at  14:56:22
   Duration  197ms (import 48%, tests 28%, transform 21%, worker 3%)
```

</details>

**🔧 Bash** _Inspect remaining failure details_
```bash
grep -n "Promise.reject\|mockRejected" src/lib/__tests__/campaigns.test.ts; pnpm vitest run src/lib/__tests__/campaigns.test.ts 2>&1 | grep -B2 -A25 "FAIL .*422" | head -50
```

<details><summary>Result</summary>

```
48:    vi.mocked(apiFetch).mockImplementation(() => Promise.reject(err))
58:    vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error("net")))
⎯⎯⎯⎯⎯⎯⎯ Failed Tests 2 ⎯⎯⎯⎯⎯⎯⎯

 FAIL  src/lib/__tests__/campaigns.test.ts > campaigns client > rejects with CampaignRequestError and details on 422
FetchError: 422
 ❯ src/lib/__tests__/campaigns.test.ts:45:17
     43|
     44|   it('rejects with CampaignRequestError and details on 422', async () …
     45|     const err = new FetchError('422')
       |                 ^
     46|     err.data = { error: { code: 'content_invalid', message: 'Bad.', de…
     47|     err.statusCode = 422

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[1/2]⎯

 FAIL  src/lib/__tests__/campaigns.test.ts > campaigns client > falls back on a network error
Error: net
 ❯ src/lib/__tests__/campaigns.test.ts:58:65
     56|
     57|   it('falls back on a network error', async () => {
     58|     vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Er…
       |                                                                 ^
     59|     const got = await getCampaign(1).catch(e => e)
     60|     expect(got).toBeInstanceOf(CampaignRequestError)

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[2/2]⎯


 Test Files  1 failed (1)
```

</details>

**🔧 Bash** _Print the caught error name and message_
```bash
cat > src/lib/__tests__/tmp.test.ts <<'EOF'
import { expect, it, vi } from 'vitest'
import { apiFetch } from '@/lib/api-client'
import { getCampaign } from '@/lib/campaigns'
vi.mock('@/lib/api-client', () => ({ apiFetch: vi.fn() }))
it('x', async () => {
  vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
  const got = await getCampaign(1).catch(e => e)
  expect(String(got?.name) + '|' + String(got?.message)).toBe('?')
})
EOF
pnpm vitest run src/lib/__tests__/tmp.test.ts 2>&1 | grep -E "Expected|Received|expected" | head; rm src/lib/__tests__/tmp.test.ts
```

<details><summary>Result</summary>

```
AssertionError: expected 'CampaignRequestError|Something went w…' to be '?' // Object.is equality
Expected: "?"
Received: "CampaignRequestError|Something went wrong. Try again."
```

</details>

**🔧 Bash** _Run only the network error test with full output_
```bash
pnpm vitest run src/lib/__tests__/campaigns.test.ts -t "network" 2>&1 | sed -n 1,60p
```

<details><summary>Result</summary>

```
RUN  v5.0.5 ~/Desktop/projects/case-studies/insider_one/frontend

 ❯ src/lib/__tests__/campaigns.test.ts (7 tests | 1 failed | 6 skipped) 3ms
   ❯ campaigns client (7)
     × falls back on a network error 2ms

⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯

 FAIL  src/lib/__tests__/campaigns.test.ts > campaigns client > falls back on a network error
Error: net
 ❯ src/lib/__tests__/campaigns.test.ts:58:65
     56|
     57|   it('falls back on a network error', async () => {
     58|     vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Er…
       |                                                                 ^
     59|     const got = await getCampaign(1).catch(e => e)
     60|     expect(got).toBeInstanceOf(CampaignRequestError)

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[1/1]⎯


 Test Files  1 failed (1)
      Tests  1 failed | 6 skipped (7)
   Start at  14:56:42
   Duration  81ms (import 70%, transform 11%, tests 10%, worker 8%)
```

</details>

**🔧 Bash** _Isolate which assertion fails_
```bash
cat > src/lib/__tests__/tmp.test.ts <<'EOF'
import { expect, it, vi } from 'vitest'
import { apiFetch } from '@/lib/api-client'
import { CampaignRequestError, getCampaign } from '@/lib/campaigns'
vi.mock('@/lib/api-client', () => ({ apiFetch: vi.fn() }))
it('a', async () => {
  vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
  const got = await getCampaign(1).catch(e => e)
  expect(got instanceof CampaignRequestError).toBe(true)
})
it('b', async () => {
  vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
  const got = await getCampaign(1).catch(e => e)
  expect(got).toBeInstanceOf(CampaignRequestError)
})
it('c', async () => {
  vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
  const got = await getCampaign(1).catch(e => e)
  expect(got.details).toEqual([])
})
EOF
pnpm vitest run src/lib/__tests__/tmp.test.ts 2>&1 | grep -E "✓|×"; rm src/lib/__tests__/tmp.test.ts
```

<details><summary>Result</summary>

```
(Bash completed with no output)
```

</details>

**🔧 Bash** _Run isolation tests with verbose reporter_
```bash
cat > src/lib/__tests__/tmp.test.ts <<'EOF'
import { expect, it, vi } from 'vitest'
import { apiFetch } from '@/lib/api-client'
import { CampaignRequestError, getCampaign } from '@/lib/campaigns'
vi.mock('@/lib/api-client', () => ({ apiFetch: vi.fn() }))
it('a', async () => {
  vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
  const got = await getCampaign(1).catch(e => e)
  expect(got instanceof CampaignRequestError).toBe(true)
})
it('b', async () => {
  vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
  const got = await getCampaign(1).catch(e => e)
  expect(got).toBeInstanceOf(CampaignRequestError)
})
it('c', async () => {
  vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
  const got = await getCampaign(1).catch(e => e)
  expect(got.details).toEqual([])
})
EOF
pnpm vitest run --reporter=verbose src/lib/__tests__/tmp.test.ts 2>&1 | head -30; rm src/lib/__tests__/tmp.test.ts
```

<details><summary>Result</summary>

```
RUN  v5.0.5 ~/Desktop/projects/case-studies/insider_one/frontend

 ✓ src/lib/__tests__/tmp.test.ts > a 1ms
 ✓ src/lib/__tests__/tmp.test.ts > b 0ms
 ✓ src/lib/__tests__/tmp.test.ts > c 0ms

 Test Files  1 passed (1)
      Tests  3 passed (3)
   Start at  14:56:58
   Duration  66ms (import 51%, transform 22%, tests 14%, worker 13%)
```

</details>

**🔧 Bash** _View network error test code_
```bash
sed -n 55,66p src/lib/__tests__/campaigns.test.ts
```

<details><summary>Result</summary>

```
})

  it('falls back on a network error', async () => {
    vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error("net")))
    const got = await getCampaign(1).catch(e => e)
    expect(got).toBeInstanceOf(CampaignRequestError)
    expect(got.message).toBe('Something went wrong. Try again.')
    expect(got.details).toEqual([])
  })
})
```

</details>

**🔧 Bash** _Fix beforeEach returning the mock as a teardown_
```bash
sed -i '' 's/beforeEach(() => vi.mocked(apiFetch).mockReset())/beforeEach(() => {\n    vi.mocked(apiFetch).mockReset()\n  })/; s/new Error("net")/new Error('"'"'net'"'"')/' src/lib/__tests__/campaigns.test.ts && pnpm vitest run src/lib 2>&1 | tail -4
```

<details><summary>Result</summary>

```
Tests  40 passed (40)
   Start at  14:57:10
   Duration  317ms (import 56%, transform 24%, tests 15%, worker 5%)
[This command modified 1 file you've previously read: src/lib/__tests__/campaigns.test.ts. Call Read before editing.]
```

</details>

**🔧 Bash** _Ledger Task 6 and inspect dialog/tabs components_
```bash
cd ~/Desktop/projects/case-studies/insider_one && cat >> .superpowers/sdd/06-review-approve/progress.md <<'EOF'
Task 6: Ruling: lib/campaigns.ts uses apiFetch (ofetch) like lib/segments.ts, and tests mock apiFetch, not fetch — same pattern as the existing client — cost if wrong: none.
Task 6: Ruling: added languageName() helper next to LANGUAGE_NAMES — the cards, dialog and list all need it — cost if wrong: none.
Task 6: complete (tests: pnpm vitest run src/lib → 40/40). Note: content-cards.vue type error from the new event union is fixed in Task 8.
EOF
cd frontend && ls src/components/ui/dialog src/components/ui/tabs src/components/ui/alert-dialog src/components/ui/select; sed -n 1,60p src/components/ui/dialog/DialogContent.vue
```

<details><summary>Result</summary>

```
src/components/ui/alert-dialog:
AlertDialog.vue
AlertDialogAction.vue
AlertDialogCancel.vue
AlertDialogContent.vue
AlertDialogDescription.vue
AlertDialogFooter.vue
AlertDialogHeader.vue
AlertDialogTitle.vue
AlertDialogTrigger.vue
index.ts

src/components/ui/dialog:
Dialog.vue
DialogClose.vue
DialogContent.vue
DialogDescription.vue
DialogFooter.vue
DialogHeader.vue
DialogOverlay.vue
DialogScrollContent.vue
DialogTitle.vue
DialogTrigger.vue
index.ts

src/components/ui/select:
Select.vue
SelectContent.vue
SelectGroup.vue
SelectItem.vue
SelectItemText.vue
SelectLabel.vue
SelectScrollDownButton.vue
SelectScrollUpButton.vue
SelectSeparator.vue
SelectTrigger.vue
SelectValue.vue
index.ts

src/components/ui/tabs:
Tabs.vue
TabsContent.vue
TabsList.vue
TabsTrigger.vue
index.ts
<script setup lang="ts">
import type { DialogContentEmits, DialogContentProps } from "reka-ui"
import type { HTMLAttributes } from "vue"
import { reactiveOmit } from "@vueuse/core"
import { X } from "@lucide/vue"
import {
  DialogClose,
  DialogContent,
  DialogPortal,
  useForwardPropsEmits,
} from "reka-ui"
import { cn } from "@/lib/utils"
import DialogOverlay from "./DialogOverlay.vue"

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(defineProps<DialogContentProps & { class?: HTMLAttributes["class"], showCloseButton?: boolean }>(), {
  showCloseButton: true,
})
const emits = defineEmits<DialogContentEmits>()

const delegatedProps = reactiveOmit(props, "class")

const forwarded = useForwardPropsEmits(delegatedProps, emits)
</script>

<template>
  <DialogPortal>
    <DialogOverlay />
    <DialogContent
      data-slot="dialog-content"
      v-bind="{ ...$attrs, ...forwarded }"
      :class="
        cn(
          'bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate
… [739 more chars truncated]
```

</details>

**🔧 Bash** _Read content-cards test setup and vitest config_
```bash
sed -n 1,60p src/components/campaign/__tests__/content-cards.test.ts; cat vitest.config.* 2>/dev/null | head -30; grep -rn "setupFiles" vite.config.ts vitest.config.ts 2>/dev/null
```

<details><summary>Result (error)</summary>

```
Exit code 2
// @vitest-environment happy-dom

import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { ContentEvent } from '@/lib/content'

import { ContentRequestError, generateContent } from '@/lib/content'

import ContentCards from '../content-cards.vue'

vi.mock('@/lib/content', async () => {
  class ContentRequestError extends Error {
    code: string
    constructor(code: string, message: string) {
      super(message)
      this.code = code
    }
  }
  return { generateContent: vi.fn(), ContentRequestError }
})

type OnEvent = (e: ContentEvent) => void
const gen = vi.mocked(generateContent)
const content = { sms: 'Hi Ayse', email_subject: 'Sale', email_body: 'Body' }
const limits = { first_name_runes: 5, sms_runes: 121, email_subject_runes: 44, email_body_runes: 10 }
const audience: ContentEvent = {
  event: 'audience',
  data: {
    description: 'd',
    languages: [
      { code: 'tr', email: 100, sms: 50, reachable: 110, matched: 120, no_consent: 10 },
      { code: 'en', email: 30, sms: 20, reachable: 40, matched: 45, no_consent: 5 },
    ],
  },
}

function make(runId = 1) {
  return mount(ContentCards, { props: { brief: 'my brief', rule: { a: 1 }, runId } })
}

// Starts a request that stays open and returns its onEvent.
function open(): { get: () => OnEvent, finish: () => void } {
  let on: OnEvent = () => {}
  let finish = () => {}
  gen.mockImplementationOnce((_b, _r, onEvent) => {
    on = onEvent
    return new Promise<void>((res) => {
      finish = res
    })
  })
  return { get: () => on, finish: () => finish() }
}

describe('content cards', () => {
  beforeEach(() => gen.mockReset())

  it('starts on the run id and shows progress', async () => {
    open()
    const w = make()
(eval):1: no matches found: vitest.config.*
```

</details>

**🔧 Bash** _Write review dialog tests and confirm they fail_
```bash
cat > src/components/campaign/__tests__/review-dialog.test.ts <<'EOF'
// @vitest-environment happy-dom

import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { toast } from 'vue-sonner'

import type { Campaign } from '@/lib/campaigns'

import { approveCampaign, CampaignRequestError, getCampaign, getSampleNames, saveContents } from '@/lib/campaigns'

import ReviewDialog from '../review-dialog.vue'

vi.mock('@/lib/campaigns', async (orig) => {
  const real = await orig<typeof import('@/lib/campaigns')>()
  return {
    ...real,
    getCampaign: vi.fn(),
    getSampleNames: vi.fn(),
    saveContents: vi.fn(),
    approveCampaign: vi.fn(),
  }
})
vi.mock('vue-sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

function campaign(over: Partial<Campaign> = {}): Campaign {
  return {
    id: 12,
    status: 'draft',
    brief: 'b',
    description: 'Istanbul users',
    skipped_languages: [],
    created_at: '2026-11-10T12:00:00Z',
    updated_at: '2026-11-10T12:00:00Z',
    approved_at: null,
    contents: [
      { language: 'en', sms: 'Hi {{first_name}}', email_subject: 'Sale {{first_name}}', email_body: 'Hello {{first_name}}', sms_count: 20, email_count: 30, first_name_runes: 9 },
      { language: 'tr', sms: 'Merhaba {{first_name}}', email_subject: 'İndirim', email_body: 'Merhaba {{first_name}},\nSepet', sms_count: 980, email_count: 1212, first_name_runes: 14 },
    ],
    ...over,
  }
}

const getC = vi.mocked(getCampaign)
const getS = vi.mocked(getSampleNames)
const save = vi.mocked(saveContents)
const approve = vi.mocked(approveCampaign)

async function make(c: Campaign = campaign()) {
  getC.mockResolvedValue(c)
  const w = mount(ReviewDialog, { props: { campaignId: c.id, open: true, 'onUpdate:open': (v: boolean) => w.setProps({ open: v }) }, attachTo: document.body })
  await flushPromises()
  return w
}

const $ = (sel: string) => document.body.querySelector<HTMLElement>(sel)
const $$ = (sel: string) => [...document.body.querySelectorAll<HTMLElement>(sel)]
const text = (sel: string) => $(sel)?.textContent?.trim() ?? ''

async function pressTab(sel: string) {
  const el = $(sel)!
  el.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, button: 0 }))
  el.dispatchEvent(new MouseEvent('click', { bubbles: true, button: 0 }))
  await flushPromises()
}

async function type(sel: string, value: string) {
  const el = $(sel) as HTMLInputElement | HTMLTextAreaElement
  el.value = value
  el.dispatchEvent(new Event('input', { bubbles: true }))
  await flushPromises()
}

async function click(sel: string) {
  $(sel)!.click()
  await flushPromises()
}

beforeEach(() => {
  getC.mockReset()
  getS.mockReset()
  save.mockReset()
  approve.mockReset()
  vi.mocked(toast.success).mockReset()
  getS.mockResolvedValue({ tr: ['Abdurrahmangazi', 'Ayşe', 'Mehmet'], en: ['Christopher', 'Ann'] })
})

afterEach(() => {
  document.body.innerHTML = ''
})

describe('revie
… [8211 more chars truncated]
```

<details><summary>Result</summary>

```
Tests  no tests
   Start at  14:58:10
   Duration  276ms (environment 99%, worker 1%)
```

</details>

**🔧 Bash** _Implement review dialog, preview and warning token; run tests_
```bash
cat > src/components/campaign/message-preview.vue <<'EOF'
<script setup lang="ts">
import { computed } from 'vue'

import type { Content } from '@/lib/content'

const props = defineProps<{ channel: 'sms' | 'email', content: Content, name: string }>()

const VAR = '{{first_name}}'
const render = (s: string) => s.split(VAR).join(props.name)

const sms = computed(() => render(props.content.sms))
const subject = computed(() => render(props.content.email_subject))
const body = computed(() => render(props.content.email_body))
</script>

<template>
  <div v-if="channel === 'sms'" class="bg-muted min-h-56 rounded-2xl border p-4">
    <p class="text-muted-foreground mb-3 text-center text-xs">
      Your brand · now
    </p>
    <p data-test="sms-preview" class="bg-secondary text-secondary-foreground max-w-[85%] rounded-2xl rounded-bl-sm px-3 py-2.5 text-sm leading-snug break-words whitespace-pre-wrap">
      {{ sms }}
    </p>
  </div>
  <div v-else data-test="email-preview" class="overflow-hidden rounded-lg border">
    <div class="bg-muted border-b px-4 py-3">
      <p class="text-muted-foreground text-xs">
        From: Your brand
      </p>
      <p class="mt-1 font-medium break-words">
        {{ subject }}
      </p>
    </div>
    <p class="px-4 py-4 text-sm leading-relaxed break-words whitespace-pre-wrap">
      {{ body }}
    </p>
  </div>
</template>
EOF
cat > src/components/campaign/review-dialog.vue <<'EOF'
<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'

import type { Campaign, ContentInput } from '@/lib/campaigns'
import type { Content } from '@/lib/content'

import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { approveCampaign, CampaignRequestError, getCampaign, getSampleNames, languageName, saveContents } from '@/lib/campaigns'
import { checkContent, LIMITS, renderedLength } from '@/lib/limits'

import MessagePreview from './message-preview.vue'

type Field = keyof Content
type Channel = 'sms' | 'email'

const props = defineProps<{ campaignId: number }>()
const emit = defineEmits<{ approved: [id: number] }>()
const open = defineModel<boolean>('open', { default: false })

const FALLBACK_NAME = 'Alex'
const FIELDS: Field[] = ['sms', 'email_subject', 'email_body']

const campaign 
… [17164 more chars truncated]
```

<details><summary>Result</summary>

```
22:  --warning: oklch(0.666 0.179 58.318);
58:  --warning: oklch(0.769 0.188 70.08);
93:  --color-warning: var(--warning);
Spinner.vue
index.ts
10:  defineProps<SelectTriggerProps & { class?: HTMLAttributes["class"], size?: "sm" | "default" }>(),
11:  { size: "default" },
14:const delegatedProps = reactiveOmit(props, "class", "size")
      Tests  no tests
```

</details>

**🔧 Bash** _Show test run errors_
```bash
pnpm vitest run src/components/campaign/__tests__/review-dialog.test.ts 2>&1 | head -40
```

<details><summary>Result</summary>

```
RUN  v5.0.5 ~/Desktop/projects/case-studies/insider_one/frontend

 ❯ src/components/campaign/__tests__/review-dialog.test.ts (0 test)

⎯⎯⎯⎯⎯⎯ Failed Suites 1 ⎯⎯⎯⎯⎯⎯⎯

 FAIL  src/components/campaign/__tests__/review-dialog.test.ts [ src/components/campaign/__tests__/review-dialog.test.ts ]
SyntaxError: Error parsing JavaScript expression: Unterminated string constant. (1:1)
317 |                  </p>
318 |                  <p class="text-muted-foreground text-xs">
319 |  ...}} characters). Use <code class="bg-muted rounded px-1">{{ '{{first_name}}' }}</code> for the name.
    |                                                                 ^
320 |                  </p>
321 |                </div>
⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[1/1]⎯


 Test Files  1 failed (1)
      Tests  no tests
   Start at  14:59:18
   Duration  218ms (environment 98%, worker 2%)
```

</details>

**🔧 Bash** _Fix template literal issue and rerun tests_
```bash
python3 - <<'EOF'
p='src/components/campaign/review-dialog.vue'
s=open(p).read()
s=s.replace("""<code class="bg-muted rounded px-1">{{ '{{first_name}}' }}</code>""","""<code class="bg-muted rounded px-1" v-text="VAR_LABEL" />""")
s=s.replace("""const FALLBACK_NAME = 'Alex'""","""const FALLBACK_NAME = 'Alex'
const VAR_LABEL = '{{first_name}}'""")
open(p,'w').write(s)
EOF
pnpm vitest run src/components/campaign/__tests__/review-dialog.test.ts 2>&1 | grep -E "✓|×|Tests |Error" | head -30
```

<details><summary>Result</summary>

```
× shows field errors and a tab mark 42ms
⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯
AssertionError: expected '' to be 'The subject must be one line.' // Object.is equality
      Tests  1 failed | 14 passed (15)
```

</details>

**🔧 Bash** _Adjust subject newline test and run full frontend suite_
```bash
python3 - <<'EOF'
p='src/components/campaign/__tests__/review-dialog.test.ts'
s=open(p).read()
s=s.replace("""    await type('[data-test="subject-input"]', 'a\\nb')
    expect(text('[data-test="subject-error"]')).toBe('The subject must be one line.')""","""    // A single-line input drops line breaks, so the browser cannot send one.
    await type('[data-test="subject-input"]', 'a\\nb')
    expect(($('[data-test="subject-input"]') as HTMLInputElement).value).not.toContain('\\n')
    await type('[data-test="subject-input"]', `{{first_name}}${'b'.repeat(67)}`)
    expect(text('[data-test="subject-error"]')).toBe('Too long after the name is filled in.')""")
open(p,'w').write(s)
EOF
pnpm vitest run src/components/campaign/__tests__/review-dialog.test.ts 2>&1 | grep -E "×|Tests |Error" | head; pnpm vitest run 2>&1 | tail -4
```

<details><summary>Result</summary>

```
Tests  15 passed (15)
      Tests  128 passed (128)
   Start at  14:59:37
   Duration  1.47s (import 42%, tests 33%, environment 17%, transform 8%, worker 1%)
```

</details>

**🔧 Bash** _Run frontend lint and type check_
```bash
CI=true pnpm lint 2>&1 | tail -20; pnpm vue-tsc -b 2>&1 | tail -5
```

<details><summary>Result</summary>

```
$ eslint .

~/Desktop/projects/case-studies/insider_one/frontend/src/components/campaign/__tests__/review-dialog.test.ts
  50:44  error  Inconsistently quoted property 'campaignId' found  style/quote-props
  50:62  error  Inconsistently quoted property 'open' found        style/quote-props

~/Desktop/projects/case-studies/insider_one/frontend/src/components/campaign/review-dialog.vue
  414:52  warning  Expected 1 line break after opening tag (`<dd>`), but no line breaks found    vue/singleline-html-element-content-newline
  414:72  warning  Expected 1 line break before closing tag (`</dd>`), but no line breaks found  vue/singleline-html-element-content-newline

✖ 4 problems (2 errors, 2 warnings)
  2 errors and 2 warnings potentially fixable with the `--fix` option.

[ELIFECYCLE] Command failed with exit code 1.
src/components/campaign/content-cards.vue(61,31): error TS2339: Property 'language' does not exist on type '{ language: string; content: Content; } | { language: string; errors?: { path: string; message: string; }[] | undefined; } | { language: string; content: Content; limits: ContentLimits; } | { ...; } | { ...; }'.
  Property 'language' does not exist on type '{ id: number; status: CampaignStatus; }'.
```

</details>

**🔧 Bash** _Autofix lint, ledger Task 7, read content-cards rest_
```bash
pnpm eslint --fix src/components/campaign src/lib >/dev/null 2>&1; CI=true pnpm lint 2>&1 | tail -4; cat >> ../.superpowers/sdd/06-review-approve/progress.md <<'EOF'
Task 7: Ruling: FR-32 "subject with a line break" cannot happen in the browser — the single-line Input drops line breaks — the test checks that instead, and the backend still rejects \n (FR-15) — cost if wrong: none.
Task 7: Ruling: all close paths go through Dialog @update:open (requestClose), not escape-key-down/pointer-down-outside handlers — one path covers close button, Escape and outside click — cost if wrong: none; the Escape test passes.
Task 7: Ruling: an approved campaign in the window shows a "Close" button and the preview at full width — spec says read-only, no save or approve — cost if wrong: none.
Task 7: complete (tests: pnpm vitest run → 128/128; review-dialog 15/15)
EOF
sed -n 80,170p src/components/campaign/content-cards.vue
```

<details><summary>Result</summary>

```
$ eslint .
  }
}

async function start() {
  controller?.abort()
  const ctl = new AbortController()
  controller = ctl
  const mine = ++token
  cards.value = []
  requestError.value = ''
  setRunning(true)
  try {
    await generateContent(props.brief, props.rule, (e) => {
      if (mine === token)
        handle(e)
    }, ctl.signal)
  }
  catch (err) {
    if (mine !== token)
      return
    if (err instanceof ContentRequestError && err.code !== 'stream_closed') {
      requestError.value = err.message
    }
    else {
      for (const c of cards.value) {
        if (!c.final && !c.error)
          c.closed = true
      }
    }
  }
  finally {
    if (mine === token)
      setRunning(false)
  }
}

watch(() => props.runId, (id) => {
  if (id > 0)
    start()
}, { immediate: true })

onBeforeUnmount(() => {
  token++
  controller?.abort()
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <p v-if="running" class="text-sm text-muted-foreground" role="status">
        Generating content…
      </p>
      <span v-else />
      <Button variant="outline" :disabled="running" @click="start">
        Regenerate content
      </Button>
    </div>

    <Alert v-if="requestError" variant="destructive">
      <AlertDescription>{{ requestError }}</AlertDescription>
    </Alert>

    <Card v-for="card in cards" :key="card.code" :data-language="card.code">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          {{ langName(card.code) }}
          <span class="text-sm font-normal text-muted-foreground">
            {{ card.email }} email, {{ card.sms }} SMS
          </span>
          <Badge v-if="!card.final && !card.error && card.draft" variant="secondary">
            Draft
          </Badge>
        </CardTitle>
        <p v-if="card.retrying" class="text-sm text-muted-foreground" role="status">
          Fixing content…
        </p>
      </CardHeader>
      <CardContent class="space-y-4">
   
… [545 more chars truncated]
```

</details>

**🔧 Bash** _Read the rest of content-cards tests and router usage_
```bash
sed -n 60,80p src/components/campaign/__tests__/content-cards.test.ts; tail -30 src/components/campaign/__tests__/content-cards.test.ts; grep -rn "useRouter\|from 'vue-router'" src --include=*.vue | head -3
```

<details><summary>Result</summary>

```
const w = make()
    await flushPromises()
    expect(gen).toHaveBeenCalledTimes(1)
    expect(gen.mock.calls[0]![0]).toBe('my brief')
    expect(gen.mock.calls[0]![1]).toEqual({ a: 1 })
    expect(w.text()).toContain('Generating content…')
  })

  it('does not start without a run', async () => {
    make(0)
    await flushPromises()
    expect(gen).not.toHaveBeenCalled()
  })

  it('streams drafts and shows final with counters', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    await flushPromises()
    expect(w.text()).toContain('Turkish')
    expect(w.text()).toContain('English')
    expect(w.find('[data-language="en"]').text()).toContain('The connection closed early. Try again.')
  })

  it('shows text, not HTML', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'draft', data: { language: 'tr', content: { ...content, sms: '<b>x</b>' } } })
    await flushPromises()
    expect(w.text()).toContain('<b>x</b>')
    expect(w.find('b').exists()).toBe(false)
  })

  it('shows an error before the stream', async () => {
    gen.mockRejectedValueOnce(new ContentRequestError('empty_audience', 'No customers match.'))
    const w = make()
    await flushPromises()
    expect(w.text()).toContain('No customers match.')
    expect(w.find('button').attributes('disabled')).toBeUndefined()
  })

  it('emits running', async () => {
    const r = open()
    const w = make()
    await flushPromises()
    r.finish()
    await flushPromises()
    expect(w.emitted('update:running')!.map(a => a[0])).toEqual([true, false])
  })
})
(eval):1: no matches found: --include=*.vue
```

</details>

**🔧 Bash** _Add Task 8 content-cards tests and confirm they fail_
```bash
python3 - <<'EOF'
p='src/components/campaign/__tests__/content-cards.test.ts'
s=open(p).read()
s=s.replace("""import ContentCards from '../content-cards.vue'
""","""import ContentCards from '../content-cards.vue'

const push = vi.fn()
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))
vi.mock('../review-dialog.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      props: { campaignId: Number, open: Boolean },
      emits: ['approved', 'update:open'],
      setup(props, { emit }) {
        return () => props.open
          ? h('div', { 'data-test': 'review-stub', 'data-id': props.campaignId }, [
              h('button', { 'data-test': 'stub-approve', 'onClick': () => emit('approved', props.campaignId) }, 'approve'),
            ])
          : null
      },
    }),
  }
})
""",1)
s=s.rstrip()
assert s.endswith('})')
s=s[:-2]+"""
  it('review button needs a campaign event', async () => {
    const r = open()
    const w = make()
    await flushPromises()
    const btn = () => w.find('[data-test="review-button"]')
    expect(btn().exists()).toBe(false)
    r.get()(audience)
    r.get()({ event: 'final', data: { language: 'tr', content, limits } })
    r.get()({ event: 'campaign', data: { id: 7, status: 'draft' } })
    await flushPromises()
    expect(btn().attributes('disabled')).toBeDefined()
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(btn().attributes('disabled')).toBeUndefined()
    await btn().trigger('click')
    expect(w.find('[data-test="review-stub"]').attributes('data-id')).toBe('7')
    await w.find('[data-test="stub-approve"]').trigger('click')
    expect(push).toHaveBeenCalledWith('/campaigns/7')
  })

  it('no review button without a campaign event', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'error', data: { language: 'tr', code: 'llm_error', message: 'x' } })
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(w.find('[data-test="review-button"]').exists()).toBe(false)
  })

  it('regenerate sends the campaign id and a new run does not', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'campaign', data: { id: 7, status: 'draft' } })
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(gen.mock.calls[0]![4]).toBeUndefined()
    open()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(gen.mock.calls[1]![4]).toBe(7)
    open()
    await w.setProps({ runId: 2 })
    await flushPromises()
    expect(gen.mock.calls[2]![4]).toBeUndefined()
  })

  it('shows an error with no language in the content area', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'final', data: { language: 'tr', content, limits } })
    r.get()({ event: 'error', data: { code: 'draf
… [482 more chars truncated]
```

<details><summary>Result</summary>

```
15:    data-slot="alert"
     × review button needs a campaign event 2ms
     × regenerate sends the campaign id and a new run does not 2ms
     × shows an error with no language in the content area 1ms
⎯⎯⎯⎯⎯⎯⎯ Failed Tests 3 ⎯⎯⎯⎯⎯⎯⎯
      Tests  3 failed | 12 passed (15)
```

</details>

## 🤖 Claude  <sub>2026-11-10 12:00:23</sub>

Now content-cards changes.

**🔧 Bash** _View content-cards handle function_
```bash
sed -n 40,80p src/components/campaign/content-cards.vue
```

<details><summary>Result</summary>

```
running.value = v
  emit('update:running', v)
}

function cardFor(code: string) {
  return cards.value.find(c => c.code === code)
}

function handle(e: ContentEvent) {
  if (e.event === 'audience') {
    cards.value = e.data.languages.map(l => reactive({
      code: l.code,
      email: l.email,
      sms: l.sms,
      retrying: false,
      closed: false,
    }))
    return
  }
  if (e.event === 'done')
    return
  const card = cardFor(e.data.language)
  if (!card)
    return
  if (e.event === 'draft') {
    card.draft = e.data.content
    card.retrying = false
  }
  else if (e.event === 'retry') {
    card.draft = undefined
    card.retrying = true
  }
  else if (e.event === 'final') {
    card.final = { content: e.data.content, limits: e.data.limits }
    card.retrying = false
    card.error = undefined
  }
  else if (e.event === 'error') {
    card.error = e.data.message
    card.retrying = false
  }
```

</details>

**🔧 Bash** _Update content-cards script for campaign id and review_
```bash
python3 - <<'EOF'
p='src/components/campaign/content-cards.vue'
s=open(p).read()
def rep(a,b):
    global s
    assert a in s, a
    s=s.replace(a,b,1)
rep("""import { onBeforeUnmount, reactive, ref, watch } from 'vue'
""","""import { onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
""")
rep("""import { ContentRequestError, generateContent } from '@/lib/content'
""","""import { ContentRequestError, generateContent } from '@/lib/content'

import ReviewDialog from './review-dialog.vue'
""")
rep("""const requestError = ref('')
""","""const requestError = ref('')
const campaignId = ref<number | null>(null)
const reviewOpen = ref(false)
const router = useRouter()
""")
rep("""  if (e.event === 'done')
    return
  const card = cardFor(e.data.language)""","""  if (e.event === 'done')
    return
  if (e.event === 'campaign') {
    campaignId.value = e.data.id
    return
  }
  if (e.event === 'error' && e.data.language === undefined) {
    requestError.value = e.data.message
    return
  }
  const card = cardFor(e.data.language!)""")
rep("""async function start() {
  controller?.abort()""","""// A regenerate keeps the draft id, so the backend replaces that draft.
// A new run (a new "Find audience") starts a new campaign.
async function start(regenerate = false) {
  if (!regenerate) {
    campaignId.value = null
    reviewOpen.value = false
  }
  const draftId = campaignId.value ?? undefined
  controller?.abort()""")
rep("""    }, ctl.signal)""","""    }, ctl.signal, draftId)""")
rep("""      <Button variant="outline" :disabled="running" @click="start">""","""      <Button variant="outline" :disabled="running" @click="start(true)">""")
open(p,'w').write(s)
EOF
grep -n "^  </div>\n</template>" src/components/campaign/content-cards.vue; tail -12 src/components/campaign/content-cards.vue
```

<details><summary>Result</summary>

```
222:  </div>
223|</template>
                Email body
              </h4>
              <p class="whitespace-pre-wrap" data-field="body">
                {{ (card.final?.content ?? card.draft)!.email_body }}
              </p>
            </div>
          </div>
        </template>
      </CardContent>
    </Card>
  </div>
</template>
```

</details>

**🔧 Bash** _Add review button and dialog to content cards, run tests_
```bash
python3 - <<'EOF'
p='src/components/campaign/content-cards.vue'
s=open(p).read()
i=s.rindex("""    </Card>
  </div>
</template>""")
s=s[:i]+"""    </Card>

    <div v-if="campaignId !== null" class="flex justify-end">
      <Button data-test="review-button" class="w-full sm:w-auto" :disabled="running" @click="reviewOpen = true">
        Review &amp; approve
      </Button>
    </div>
    <ReviewDialog
      v-if="campaignId !== null"
      v-model:open="reviewOpen"
      :campaign-id="campaignId"
      @approved="id => router.push(`/campaigns/${id}`)"
    />
  </div>
</template>
"""
open(p,'w').write(s)
EOF
pnpm vitest run src/components/campaign 2>&1 | grep -E "×|Tests " | head; pnpm vue-tsc -b 2>&1 | tail -5
```

<details><summary>Result</summary>

```
Tests  42 passed (42)
```

</details>

**🔧 Bash** _Check badge variants, table parts and auto-import config_
```bash
grep -n "variant" src/components/ui/badge/index.ts | head; ls src/components/ui/table; grep -rn "AutoImport\|imports:" vite.config.ts | head
```

<details><summary>Result</summary>

```
9:    variants: {
10:      variant: {
22:      variant: "default",
Table.vue
TableBody.vue
TableCaption.vue
TableCell.vue
TableEmpty.vue
TableFooter.vue
TableHead.vue
TableHeader.vue
TableRow.vue
index.ts
vite.config.ts:8:import AutoImport from 'unplugin-auto-import/vite'
vite.config.ts:12:import { VueRouterAutoImports } from 'vue-router/unplugin'
vite.config.ts:31:    AutoImport({
vite.config.ts:36:      imports: [
vite.config.ts:38:        VueRouterAutoImports,
```

</details>

## 🤖 Claude  <sub>2026-11-10 12:00:54</sub>

Page tests first.

**🔧 Bash** _Write page tests and confirm they fail_
```bash
mkdir -p ~/Desktop/projects/case-studies/insider_one/frontend/src/pages/campaigns/__tests__ && cd ~/Desktop/projects/case-studies/insider_one/frontend && cat > src/pages/campaigns/__tests__/index.test.ts <<'EOF'
// @vitest-environment happy-dom

import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { listCampaigns } from '@/lib/campaigns'

import CampaignsPage from '../index.vue'

const push = vi.fn()
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))
vi.mock('@/lib/campaigns', async (orig) => {
  const real = await orig<typeof import('@/lib/campaigns')>()
  return { ...real, listCampaigns: vi.fn() }
})
vi.mock('@/components/campaign/review-dialog.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      props: { campaignId: Number, open: Boolean },
      emits: ['approved', 'update:open'],
      setup(props, { emit }) {
        return () => props.open
          ? h('div', { 'data-test': 'review-stub', 'data-id': props.campaignId }, [
              h('button', { 'data-test': 'stub-approve', 'onClick': () => emit('approved', props.campaignId) }, 'approve'),
              h('button', { 'data-test': 'stub-close', 'onClick': () => emit('update:open', false) }, 'close'),
            ])
          : null
      },
    }),
  }
})

const list = vi.mocked(listCampaigns)

function make() {
  return mount(CampaignsPage, { global: { stubs: { RouterLink: RouterLinkStub } } })
}

describe('campaigns page', () => {
  beforeEach(() => {
    list.mockReset()
    push.mockReset()
  })

  it('lists campaigns with status badges', async () => {
    list.mockResolvedValue([
      { id: 2, status: 'draft', description: 'Istanbul users', languages: ['en', 'tr'], created_at: '2026-11-10T12:00:00Z', approved_at: null },
      { id: 1, status: 'approved', description: 'Ankara users', languages: ['tr'], created_at: '2026-11-09T12:00:00Z', approved_at: '2026-11-09T13:00:00Z' },
    ])
    const w = make()
    await flushPromises()
    const rows = w.findAll('[data-test="campaign-row"]')
    expect(rows).toHaveLength(2)
    expect(rows[0]!.text()).toContain('#2')
    expect(rows[0]!.text()).toContain('Istanbul users')
    expect(rows[0]!.text()).toContain('English, Turkish')
    expect(rows[0]!.text()).toContain('Draft')
    expect(rows[1]!.text()).toContain('Approved')
  })

  it('a draft opens the review window, an approved campaign opens its page', async () => {
    list.mockResolvedValue([
      { id: 2, status: 'draft', description: 'd', languages: ['tr'], created_at: '2026-11-10T12:00:00Z', approved_at: null },
      { id: 1, status: 'approved', description: 'a', languages: ['tr'], created_at: '2026-11-09T12:00:00Z', approved_at: '2026-11-09T13:00:00Z' },
    ])
    const w = make()
    await flushPromises()
    await w.findAll('[data-test="campaign-row"]')[0]!.trigger('click')
    exp
… [5165 more chars truncated]
```

<details><summary>Result</summary>

```
× shows an approved campaign read-only 6ms
     × a draft shows Review & approve and reloads after approval 2ms
     × shows not found 1ms
     × lists campaigns with status badges 8ms
     × a draft opens the review window, an approved campaign opens its page 1ms
     × reloads after approval and goes to the campaign 1ms
     × reloads after the window closes 0ms
     × shows the empty state 1ms
     × shows a load error 1ms
⎯⎯⎯⎯⎯⎯⎯ Failed Tests 9 ⎯⎯⎯⎯⎯⎯⎯
```

</details>

**🔧 Bash** _Implement campaign list and campaign pages, run tests_
```bash
cat > src/pages/campaigns/index.vue <<'EOF'
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import type { CampaignSummary } from '@/lib/campaigns'

import ReviewDialog from '@/components/campaign/review-dialog.vue'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { CampaignRequestError, languageName, listCampaigns } from '@/lib/campaigns'

const router = useRouter()
const campaigns = ref<CampaignSummary[] | null>(null)
const error = ref('')
const reviewId = ref<number | null>(null)
const reviewOpen = ref(false)

async function load() {
  error.value = ''
  try {
    campaigns.value = await listCampaigns()
  }
  catch (err) {
    error.value = err instanceof CampaignRequestError ? err.message : 'Something went wrong. Try again.'
  }
}

onMounted(load)

function languages(c: CampaignSummary) {
  return c.languages.map(languageName).join(', ')
}

function when(iso: string) {
  return new Date(iso).toLocaleString()
}

function openCampaign(c: CampaignSummary) {
  if (c.status === 'draft') {
    reviewId.value = c.id
    reviewOpen.value = true
  }
  else {
    router.push(`/campaigns/${c.id}`)
  }
}

function onReviewOpen(v: boolean) {
  reviewOpen.value = v
  if (!v)
    void load()
}

function onApproved(id: number) {
  router.push(`/campaigns/${id}`)
}
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-2">
      <h1 class="text-2xl font-bold tracking-tight">
        Campaigns
      </h1>
      <p class="text-muted-foreground">
        Drafts wait for your review. Approved campaigns are ready to send.
      </p>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <div v-else-if="campaigns && campaigns.length === 0" class="space-y-3 rounded-lg border border-dashed p-8 text-center">
      <p class="font-medium">
        No campaigns yet
      </p>
      <Button as-child variant="outline">
        <RouterLink to="/campaigns/new">
          Create a campaign
        </RouterLink>
      </Button>
    </div>

    <template v-else-if="campaigns">
      <div class="hidden rounded-lg border md:block">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-16">
                #
              </TableHead>
              <TableHead>Audience</TableHead>
              <TableHead>Languages</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="c in campaigns"
              :key="c.id"
              data-test="campaign-row"
              class=
… [6488 more chars truncated]
```

<details><summary>Result</summary>

```
× shows not found 2ms
⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯
CampaignRequestError: No campaign has this id.
      Tests  1 failed | 13 passed (14)
```

</details>

**🔧 Bash** _Fix test teardown, run full frontend suite, lint, types_
```bash
sed -i '' 's/  beforeEach(() => get.mockReset())/  beforeEach(() => {\n    get.mockReset()\n  })/' src/pages/campaigns/__tests__/id.test.ts && pnpm eslint --fix src >/dev/null 2>&1; pnpm vitest run 2>&1 | tail -4; CI=true pnpm lint 2>&1 | tail -5; pnpm vue-tsc -b 2>&1 | tail -5
```

<details><summary>Result</summary>

```
Tests  141 passed (141)
   Start at  15:01:59
   Duration  1.87s (import 39%, tests 24%, environment 20%, transform 16%, worker 1%)

$ eslint .
```

</details>

**🔧 Bash** _Ledger Task 8 and check compose ports_
```bash
cd ~/Desktop/projects/case-studies/insider_one && echo "Task 8: complete (tests: pnpm vitest run → 141/141; lint and vue-tsc clean)" >> .superpowers/sdd/06-review-approve/progress.md; grep -n "ports" -A2 docker-compose.yml compose.yaml 2>/dev/null | head -30; docker ps --format '{{.Names}} {{.Ports}}'
```

<details><summary>Result</summary>

```
16:    # No ports: Postgres is reachable only inside the compose network.
17-    # This avoids a conflict with a Postgres on the host's port 5432.
18-
--
29:    # No ports, for the same reason as Postgres (host port 6379).
30-
31-  api:
--
46:    ports:
47-      - "8080:8080"
48-    depends_on:
--
97:    ports:
98-      - "8081:8081"
99-
--
104:    ports:
105-      - "3000:8080"
106-    depends_on:
cc-dev-pg 127.0.0.1:55433->5432/tcp
insider_one-frontend-1 0.0.0.0:3000->8080/tcp, [::]:3000->8080/tcp
insider_one-api-1 0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp
insider_one-mock-provider-1 0.0.0.0:8081->8081/tcp, [::]:8081->8081/tcp
insider_one-redis-1 6379/tcp
insider_one-postgres-1 127.0.0.1:5433->5432/tcp
```

</details>

**🔧 Bash** _Read api and frontend compose service config_
```bash
sed -n 31,60p docker-compose.yml; sed -n 95,115p docker-compose.yml; docker compose version
```

<details><summary>Result</summary>

```
api:
    build:
      context: .
      target: go-service
      args:
        CMD: api
    env_file:
      - path: .env
        required: false
    environment:
      DATABASE_URL: postgres://campaign:campaign@postgres:5432/campaign?sslmode=disable
      LLM_MODE: ${LLM_MODE:-mock}
      ANTHROPIC_MODEL: ${ANTHROPIC_MODEL:-claude-opus-5-5}
      # Origins that may call the api from a browser (D-028).
      CORS_ORIGINS: ${CORS_ORIGINS:-http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000,http://127.0.0.1:5173}
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      # The api starts after the seed applies the migrations (D-015).
      seed:
        condition: service_completed_successfully

  # One-shot job. It applies the migrations, loads the users, then exits.
  # If users exist, it keeps them and exits 0. With SEED_RESET=true it
  # deletes all users and loads them again on each start.
  seed:
    build:
      context: .
      # Empty means a seed from the clock.
      MOCK_RANDOM: ${MOCK_RANDOM:-}
    ports:
      - "8081:8081"

  frontend:
    build:
      context: .
      target: frontend
    ports:
      - "3000:8080"
    depends_on:
      - api

volumes:
  pgdata:
  redisdata:
Docker Compose version v5.7.0
```

</details>

**🔧 Bash** _Find frontend build args in Dockerfile_
```bash
grep -n "ARG\|VITE" Dockerfile | head
```

<details><summary>Result</summary>

```
11:# ARG CMD comes after the download, so all binaries share this cache layer.
14:ARG CMD
30:ARG VITE_SERVER_API_URL=http://localhost:8080
31:ARG VITE_SERVER_API_PREFIX=/api
32:ARG VITE_SERVER_API_TIMEOUT=10000
33:ENV VITE_SERVER_API_URL=${VITE_SERVER_API_URL} \
34:    VITE_SERVER_API_PREFIX=${VITE_SERVER_API_PREFIX} \
35:    VITE_SERVER_API_TIMEOUT=${VITE_SERVER_API_TIMEOUT}
```

</details>

**🔧 Bash** _Start an isolated verify stack on other ports_
```bash
S=/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad; cat > $S/cc-verify.override.yml <<'EOF'
services:
  api:
    ports: !override
      - "18080:8080"
    environment:
      CORS_ORIGINS: http://localhost:13000
  mock-provider:
    ports: !override
      - "18081:8081"
  frontend:
    build:
      args:
        VITE_SERVER_API_URL: http://localhost:18080
    ports: !override
      - "13000:8080"
EOF
cd ~/Desktop/projects/case-studies/insider_one && env -u ANTHROPIC_API_KEY LLM_MODE=mock docker compose -p cc-verify -f docker-compose.yml -f $S/cc-verify.override.yml --env-file /dev/null up -d --build --wait > $S/up.log 2>&1; echo rc=$?; tail -5 $S/up.log
```

<details><summary>Result</summary>

```
rc=0
 Container cc-verify-frontend-1 Healthy 
 Container cc-verify-api-1 Healthy 
 Container cc-verify-mock-provider-1 Healthy 
 Container cc-verify-redis-1 Healthy 
 Container cc-verify-postgres-1 Healthy
```

</details>

**🔧 Bash** _Run segment preview and content stream end to end_
```bash
cd /tmp; B=http://localhost:18080/api; RULE=$(curl -s -X POST $B/segments/preview -H 'Content-Type: application/json' -d '{"brief":"Son 30 günde sepetine ürün ekleyip alışveriş yapmayan İstanbul'"'"'daki kullanıcılara %30 indirim hatırlatması gönder. Samimi ama kısa olsun."}' | python3 -c 'import sys,json; print(json.dumps(json.load(sys.stdin)["rule"]))'); echo "$RULE" | head -c 200; echo; curl -sN -X POST $B/content/generate -H 'Content-Type: application/json' -d "{\"brief\":\"Son 30 günde sepetine ürün ekleyip alışveriş yapmayan İstanbul'daki kullanıcılara %30 indirim hatırlatması gönder. Samimi ama kısa olsun.\",\"rule\":$RULE}" | grep -E "^event: (audience|final|campaign|error|done)" -A1 | grep -v "^--" | cut -c1-160
```

<details><summary>Result</summary>

```
{"op": "and", "conditions": [{"field": "city", "op": "eq", "value": "Istanbul"}, {"field": "last_cart_add_at", "op": "within_days", "value": 30}, {"op": "or", "conditions": [{"field": "last_purchase_a
event: audience
data: {"description":"city is Istanbul AND last cart add in the last 30 days AND (no purchase OR last purchase more than 30 days ago)","languages":[{"code":"tr"
event: final
data: {"language":"tr","content":{"sms":"Merhaba {{first_name}}, sepetin seni bekliyor! Bugün tamamla, %30 indirimi kaçırma.","email_subject":"{{first_name}}, s
event: campaign
data: {"id":1,"status":"draft"}
event: done
data: {}
```

</details>

**🔧 Bash** _Exercise list, sample, approve, repeat, put and CORS_
```bash
B=http://localhost:18080/api; curl -s $B/campaigns; echo; curl -s $B/campaigns/1/sample; echo; C=$(curl -s $B/campaigns/1 | python3 -c 'import sys,json; c=json.load(sys.stdin); print(json.dumps({"contents":[{k:x[k] for k in ("language","sms","email_subject","email_body")} for x in c["contents"]]}))'); curl -s -o /dev/null -w "approve %{http_code}\n" -X POST $B/campaigns/1/approve -H 'Content-Type: application/json' -d "$C"; curl -s -X POST $B/campaigns/1/approve -H 'Content-Type: application/json' -d "$C" -w "\nrepeat %{http_code}\n" | tail -c 120; curl -s -w "\nput %{http_code}\n" -X PUT $B/campaigns/1/contents -H 'Content-Type: application/json' -d "$C"; curl -s -i -X OPTIONS $B/campaigns/1/contents -H 'Origin: http://localhost:13000' -H 'Access-Control-Request-Method: PUT' | grep -i "allow-methods"
```

<details><summary>Result</summary>

```
{"campaigns":[{"id":1,"status":"draft","description":"city is Istanbul AND last cart add in the last 30 days AND (no purchase OR last purchase more than 30 days ago)","languages":["tr"],"created_at":"2026-11-10T12:02:57.810543Z","approved_at":null}]}

{"languages":[{"code":"tr","first_names":["Hüseyin","Çiğdem","Ufuk","Can","Deniz"]}]}

approve 200
rimden yararlanabilirsin.\n\nGörüşmek üzere!","sms_count":101,"email_count":148,"first_name_runes":7}]}

repeat 200
{"error":{"code":"campaign_not_draft","message":"Only a draft can change. This campaign is already approved."}}

put 409
Access-Control-Allow-Methods: GET, POST, PUT, OPTIONS
```

</details>

**🔧 Bash** _Look for a browser automation tool from earlier slices_
```bash
cd ~/Desktop/projects/case-studies/insider_one; ls frontend/node_modules | grep -iE "puppeteer|playwright"; ls ~/.cache/ms-playwright 2>/dev/null; grep -rl "puppeteer\|playwright\|remote-debugging" .superpowers/sdd/*/ 2>/dev/null | head -3; ls .superpowers/sdd/
```

<details><summary>Result</summary>

```
06-review-approve
```

</details>

**🔧 Bash** _Install puppeteer-core in the scratchpad for the browser check_
```bash
S=/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad; mkdir -p $S/e2e && cd $S/e2e && npm init -y >/dev/null && npm i puppeteer-core@24.10.0 >/dev/null 2>&1; echo rc=$?; ls node_modules | grep puppeteer
```

<details><summary>Result</summary>

```
rc=0
@puppeteer
puppeteer-core
```

</details>

**🔧 Bash** _Find brief input and button markup_
```bash
grep -n "Textarea\|Find audience\|<Button" ~/Desktop/projects/case-studies/insider_one/frontend/src/components/campaign/brief-audience.vue | head
```

<details><summary>Result</summary>

```
9:import { Textarea } from '@/components/ui/textarea'
52:      <Textarea
62:        <Button :disabled="!canSubmit || pending || running" @click="submit">
63:          Find audience
```

</details>

**🔧 Bash** _Run the headless Chrome flow at 1280 and 360 px_
```bash
cat > run.mjs <<'EOF'
import puppeteer from 'puppeteer-core'
const BRIEF = "Son 30 günde sepetine ürün ekleyip alışveriş yapmayan İstanbul'daki kullanıcılara %30 indirim hatırlatması gönder. Samimi ama kısa olsun."
const base = 'http://localhost:13000'
const browser = await puppeteer.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true })
const out = {}
for (const width of [1280, 360]) {
  const page = await browser.newPage()
  await page.setViewport({ width, height: 860 })
  const errors = []
  page.on('pageerror', e => errors.push(e.message))
  await page.goto(`${base}/campaigns/new`, { waitUntil: 'networkidle0' })
  await page.type('textarea', BRIEF)
  await page.evaluate(() => [...document.querySelectorAll('button')].find(b => b.textContent.includes('Find audience')).click())
  await page.waitForSelector('[data-test="review-button"]:not([disabled])', { timeout: 60000 })
  await page.click('[data-test="review-button"]')
  await page.waitForSelector('[data-test="sms-input"]', { timeout: 10000 })
  await new Promise(r => setTimeout(r, 400))
  const scrollOpen = await page.evaluate(() => document.documentElement.scrollWidth)
  await page.screenshot({ path: `review-sms-${width}.png` })
  const preview = await page.$eval('[data-test="sms-preview"]', e => e.textContent.trim())
  const counter = await page.$eval('[data-test="sms-counter"]', e => e.textContent.trim())
  await page.click('[data-test="channel-tab"][data-value="email"]')
  await page.waitForSelector('[data-test="subject-input"]')
  await page.screenshot({ path: `review-email-${width}.png` })
  await page.click('[data-test="approve"]')
  await page.waitForSelector('[data-test="send-now"]')
  await new Promise(r => setTimeout(r, 300))
  await page.screenshot({ path: `confirm-${width}.png` })
  await page.click('[data-test="send-now"]')
  await page.waitForFunction(() => /\/campaigns\/\d+$/.test(location.pathname), { timeout: 10000 })
  await page.waitForFunction(() => document.body.innerText.includes('Approved'), { timeout: 10000 })
  const url = page.url()
  await page.screenshot({ path: `campaign-${width}.png`, fullPage: true })
  await page.goto(`${base}/campaigns`, { waitUntil: 'networkidle0' })
  const scrollList = await page.evaluate(() => document.documentElement.scrollWidth)
  await page.screenshot({ path: `list-${width}.png`, fullPage: true })
  out[width] = { scrollOpen, scrollList, preview, counter, url, errors }
  await page.close()
}
console.log(JSON.stringify(out, null, 1))
await browser.close()
EOF
node run.mjs 2>&1 | tail -30
```

<details><summary>Result</summary>

```
{
 "360": {
  "scrollOpen": 360,
  "scrollList": 360,
  "preview": "Merhaba Hüseyin, sepetin seni bekliyor! Bugün tamamla, %30 indirimi kaçırma.",
  "counter": "76 / 160",
  "url": "http://localhost:13000/campaigns/3",
  "errors": []
 },
 "1280": {
  "scrollOpen": 1280,
  "scrollList": 1280,
  "preview": "Merhaba Hüseyin, sepetin seni bekliyor! Bugün tamamla, %30 indirimi kaçırma.",
  "counter": "76 / 160",
  "url": "http://localhost:13000/campaigns/2",
  "errors": []
 }
}
```

</details>

## 🤖 Claude  <sub>2026-11-10 12:03:50</sub>

Flow passes at both widths. Checking the screenshots for design.

**🔧 Read** `/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad/e2e/review-sms-1280.png`

<details><summary>Result</summary>

```
[image]
```

</details>

**🔧 Read** `/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad/e2e/review-email-360.png`

<details><summary>Result</summary>

```
[image]
```

</details>

**🔧 Read** `/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad/e2e/list-360.png`

<details><summary>Result</summary>

```
[image]
```

</details>

**🔧 Read** `/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad/e2e/confirm-1280.png`

<details><summary>Result</summary>

```
[image]
```

</details>

**🔧 Bash** _Stop verify stack and locate docs sections to update_
```bash
cd ~/Desktop/projects/case-studies/insider_one && S=/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/scratchpad; docker compose -p cc-verify -f docker-compose.yml -f $S/cc-verify.override.yml --env-file /dev/null down -v >/dev/null 2>&1; echo down=$?; grep -n "3. \*\*Approval\|TBD per slice\|Recipients. is the only\|segment/ \|delivery/" docs/architecture.md; tail -8 docs/decisions-log.md
```

<details><summary>Result</summary>

```
down=0
24:    segment/        rule DSL, validation, SQL builder
27:    delivery/       queue, rate limiter, retry, idempotency
53:   - A user is in the count of a channel only if the user has consent for that channel. The preview count applies consent with `FILTER` columns in `CountAudience`. `Recipients` is the only function that returns user rows. It always adds the consent condition for its channel. Slices 05 to 07 get recipients only through `Recipients`.
63:3. **Approval** — The marketer previews, edits and approves the content.
64:   - TBD per slice.
68:   - TBD per slice.
- **Why:** Each package keeps its own data. `llm.Request` gets no field that only the mock uses.

## D-035 Stream details of the content endpoint

- **Context:** The content endpoint sends `draft` events while Claude writes. The spec did not fix four details. They are the draft throttle after a retry, the last draft, a client that disconnects, and a response that cannot flush.
- **Options:** Reset the throttle after a retry or keep it. Always send the last draft or skip it when the text is the same. Send `error` and `done` after a disconnect or send nothing. Ignore a missing flusher or return an error.
- **Choice:** The draft throttle does not reset after a retry. The handler skips the last draft when the text did not change. After a client disconnect, it sends no `error` and no `done` event. If the response cannot flush, it returns `500` with `internal_error`. `net/http` can always flush, so this path is a guard only.
- **Why:** A throttle that resets would let a retry send a burst of drafts. A repeated draft adds no information. No client reads the events after a disconnect. A stream that cannot flush is not a stream, so the handler fails early and clearly.
```

</details>

**🔧 Bash** _Update architecture flow and add decisions D-036 to D-045_
```bash
python3 - <<'EOF'
p='docs/architecture.md'
s=open(p).read()
def rep(a,b):
    global s
    assert a in s, a
    s=s.replace(a,b,1)
rep("""    segment/        rule DSL, validation, SQL builder
""","""    segment/        rule DSL, validation, SQL builder, sample first names
    campaign/       draft and approved campaigns, edit check, approval
""")
rep("""`Recipients` is the only function that returns user rows. It always adds the consent condition for its channel. Slices 05 to 07 get recipients only through `Recipients`.""","""`Recipients` is the only function that returns user rows. It always adds the consent condition for its channel. Slices 05 to 07 get recipients only through `Recipients`. `SampleFirstNames` returns first names only, for the preview in step 3.""")
rep("""3. **Approval** — The marketer previews, edits and approves the content.
   - TBD per slice.""","""3. **Approval** — The marketer previews, edits and approves the content.
   - At the end of the content stream, the backend saves a draft campaign with the final languages. Failed languages are skipped. The stream sends a `campaign` event with the id, then `done` (D-036). No draft is saved after a client disconnect or when all languages fail.
   - "Regenerate content" sends the draft id. The backend replaces the content of that draft. It compares the brief and the rule (as `jsonb`) with the stored ones (D-037).
   - The review window loads the campaign and sample first names (`GET /api/campaigns/{id}/sample`). The browser counts the limits with the stored `first_name_runes`. The backend makes the decision.
   - `PUT /api/campaigns/{id}/contents` saves the edits. `POST /api/campaigns/{id}/approve` saves the last edits and approves. Both validate the stored rule again, run the audience query again and check the content with the new longest first name (D-038).
   - The approval locks the campaign row and checks the status in one transaction. A repeated approval returns the campaign and changes nothing (D-039).
   - Database triggers block a change of the content, the brief, the rule and the approval time after approval (D-040).
   - The campaign stores the counts at the last save, but no recipient list (D-041). The send step gets the recipients at send time.""")
open(p,'w').write(s)
EOF
cat >> docs/decisions-log.md <<'EOF'

## D-036 The content endpoint saves the draft

- **Context:** The user wants drafts. Claude creates a draft when it finishes the content.
- **Options:** The browser posts the final content to a create endpoint. The content endpoint saves the draft at the end of the stream.
- **Choice:** The content endpoint saves the draft after all languages are final or failed. It sends a `campaign` event before `done`. It saves no draft when all languages fail or after a client disconnect.
- **Why:** The stored content is always content that passed the validator. The browser cannot store content that Claude did not write. A disconnected client did not see the result, so a draf
… [6007 more chars truncated]
```

<details><summary>Result</summary>

```
docs/decisions-log.md:268: sentence-length: 26 words: **Choice:** Each save and each approval parses the stored rule again, runs the audience query again, and checks the content with the new longest first name.
```

</details>

**🔧 Bash** _Fix long sentence and view backlog entry format_
```bash
sed -i '' 's/- \*\*Choice:\*\* Each save and each approval parses the stored rule again, runs the audience query again, and checks the content with the new longest first name./- **Choice:** Each save and each approval parses the stored rule, runs the audience query again and checks the content with the new longest name./' docs/decisions-log.md && python3 -c "
import json;d=json.load(open('tasks/backlog.json'))
items=d if isinstance(d,list) else d['slices']
print(json.dumps([s for s in items if s['id']=='05-content-generation'][0],indent=2)[:1200])"; head -5 tasks/backlog.json
```

<details><summary>Result</summary>

```
{
  "id": "05-content-generation",
  "title": "Content generation",
  "spec": "docs/specs/05-content-generation/",
  "acceptance": [
    "FR-1. POST /api/content/generate with the example brief, the example rule and the default mock returns 200 with text/event-stream. The events are audience, then one or more draft events and one final event for tr, then done.",
    "FR-2. These bodies return 400 with code: invalid_request: a body over 16 KiB, a body that is not valid JSON, and a body with no brief or no rule. The LLM client gets no call.",
    "FR-3. A brief that fails the slice 04 checks returns 400 with code: invalid_brief. The LLM client gets no call.",
    "FR-4. A rule that fails the slice 04 validator returns 400 with code: invalid_rule and the errors in details. The cases include an unknown field and \"Istanbul' OR 1=1 --\" for city. The backend runs no query.",
    "FR-5. A rule with no reachable user returns 422 with code: empty_audience. The LLM client gets no call.",
    "FR-6. The backend makes one LLM call for each language in the audience. A rule with country in [TR, US] gives calls for tr and en. Each request names its language in the user message.",
    "FR-7. The 
[
  {
    "id": "00-skeleton",
    "title": "Repo skeleton",
    "spec": "docs/case.md",
```

</details>

**🔧 Bash** _Fill slice 06 acceptance checks in the backlog_
```bash
python3 - <<'EOF'
import json,re
reqs=[]
cur=None
for line in open('docs/specs/06-review-approve/requirements.md'):
    m=re.match(r'- \*\*((?:FR|NFR)-\d+)\.\*\* (.*)',line.rstrip())
    if m:
        cur=[m.group(1)+'. '+m.group(2)]; reqs.append(cur); continue
    if cur and line.startswith('  - '):
        cur.append(line.strip()[2:]); continue
    if not line.strip() or line.startswith('#'):
        cur=None
out=[]
for r in reqs:
    t=r[0]
    if len(r)>1: t=t+' '+'; '.join(r[1:])+'.'
    out.append(t.replace('`',''))
raw=open('tasks/backlog.json').read()
d=json.loads(raw)
items=d if isinstance(d,list) else d['slices']
for s in items:
    if s['id']=='06-review-approve':
        s['spec']='docs/specs/06-review-approve/'
        s['acceptance']=out
ensure_ascii = '\\u' in raw and 'ş' not in raw
open('tasks/backlog.json','w').write(json.dumps(d,indent=2,ensure_ascii=False)+'\n')
print(len(out))
EOF
git diff --stat tasks/backlog.json
```

<details><summary>Result</summary>

```
56
 tasks/backlog.json | 63 +++++++++++++++++++++++++++++++++++++++++++++++++++---
 1 file changed, 60 insertions(+), 3 deletions(-)
```

</details>

**🔧 Bash** _Run make verify after all tasks_
```bash
make verify > .superpowers/sdd/06-review-approve/verify9.log 2>&1; echo rc=$?; grep -E "^(ok|FAIL)|Tests |error" .superpowers/sdd/06-review-approve/verify9.log | head -20; echo "Task 9: e2e passed (compose -p cc-verify on ports 13000/18080 with an override because the user's stack holds 3000/8080; API flow + headless Chrome 1280/360, scrollWidth = viewport); docs, D-036..D-045, backlog done" >> .superpowers/sdd/06-review-approve/progress.md
```

<details><summary>Result</summary>

```
rc=0
ok  	campaigncomposer/cmd/api	(cached)
ok  	campaigncomposer/internal/campaign	(cached)
ok  	campaigncomposer/internal/config	(cached)
ok  	campaigncomposer/internal/content	(cached)
ok  	campaigncomposer/internal/httpapi	(cached)
ok  	campaigncomposer/internal/llm	(cached)
ok  	campaigncomposer/internal/mockprovider	(cached)
ok  	campaigncomposer/internal/seed	(cached)
ok  	campaigncomposer/internal/segment	(cached)
ok  	campaigncomposer/internal/store	(cached)
      Tests  141 passed (141)
dist/assets/errors-kDMof55T.js                       0.19 kB │ gzip:  0.17 kB
dist/assets/errors-DQFC1xB4.js                       0.31 kB │ gzip:  0.24 kB
dist/assets/custom-error-DSQS8iMY.js                 0.83 kB │ gzip:  0.50 kB
```

</details>

**🔧 Bash** _List changed files and read reviewer template header_
```bash
git status --short | head -40; sed -n 1,40p ~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/requesting-code-review/code-reviewer.md | head -20
```

<details><summary>Result</summary>

````
M backend/cmd/api/main.go
 M backend/internal/httpapi/body.go
 M backend/internal/httpapi/content.go
 M backend/internal/httpapi/cors.go
 M backend/internal/httpapi/cors_test.go
 M backend/internal/store/migrate_test.go
 M docs/architecture.md
 M docs/decisions-log.md
 M frontend/src/assets/index.css
 M frontend/src/components/campaign/__tests__/content-cards.test.ts
 M frontend/src/components/campaign/content-cards.vue
 M frontend/src/lib/__tests__/content.test.ts
 M frontend/src/lib/content.ts
 M frontend/src/pages/campaigns/[id].vue
 M frontend/src/pages/campaigns/index.vue
 M frontend/src/types/auto-import-components.d.ts
 M tasks/backlog.json
?? backend/internal/campaign/
?? backend/internal/httpapi/campaigns.go
?? backend/internal/httpapi/campaigns_test.go
?? backend/internal/httpapi/content_draft_test.go
?? backend/internal/segment/sample.go
?? backend/internal/segment/sample_test.go
?? backend/internal/store/migrations/00002_campaigns.sql
?? docs/plans/06-review-approve.md
?? docs/specs/06-review-approve/
?? frontend/src/components/campaign/__tests__/review-dialog.test.ts
?? frontend/src/components/campaign/message-preview.vue
?? frontend/src/components/campaign/review-dialog.vue
?? frontend/src/lib/__tests__/campaigns.test.ts
?? frontend/src/lib/__tests__/limits.test.ts
?? frontend/src/lib/campaigns.ts
?? frontend/src/lib/limits.ts
?? frontend/src/pages/campaigns/__tests__/
# Code Reviewer Prompt Template

Use this template when dispatching a code reviewer subagent.

**Purpose:** Review completed work against requirements and code quality standards before it cascades into more work.

```
Subagent (general-purpose):
  description: "Review code changes"
  prompt: |
    You are a Senior Code Reviewer with expertise in software architecture,
    design patterns, and best practices. Your job is to review completed work
    against its plan or requirements and identify issues before they cascade.

    ## What Was Implemented

    [DESCRIPTION]

    ## Requirement
… [8 more chars truncated]
````

</details>

**🔧 Agent** **Final whole-slice review** (reviewer)
```
You are a senior code reviewer. Review the finished slice 06 (review and approve) of the repo at ~/Desktop/projects/case-studies/insider_one. Read-only: do not edit any file.

The work is UNCOMMITTED on branch 06-review-approve. See it with `git status --short` and `git diff` (modified files) plus the untracked files listed by git status (backend/internal/campaign/, backend/internal/httpapi/campaigns*.go, content_draft_test.go, backend/internal/segment/sample*.go, backend/internal/store/migrations/00002_campaigns.sql, frontend/src/components/campaign/review-dialog.vue, message-preview.vue, frontend/src/lib/campaigns.ts, limits.ts, pages/campaigns/*, and their tests).

Authority, in order: CLAUDE.md hard rules; spec docs/specs/06-review-approve/goal.md and requirements.md; plan docs/plans/06-review-approve.md. The implementer's rulings (deliberate deviations) are the lines containing "Ruling:" in .superpowers/sdd/06-review-approve/progress.md — weigh them, flag any you disagree with.

What was built: the content SSE endpoint saves a draft campaign at the end (campaign event), regenerate replaces it; campaign store with row-lock approval and DB triggers that freeze approved campaigns; save/approve endpoints re-validate content against a fresh audience count; sample first names endpoint; Vue review dialog (SMS/Email tabs, language tabs, counters, preview, confirm popup, discard prompt), campaign list and campaign page.

Verification already done: `make verify` exit 0 (Go tests with Postgres incl. 10-way concurrent approve; 141 frontend tests; lint; build); a compose stack run with headless Chrome at 1280 px and 360 px completed the full flow with no horizontal scroll.

Check these deliberately (the plan's Review Focus — conditions the tests may not fully exercise):
1. An aborted/old content stream must not save a draft after the browser left (FR-6).
2. Postgres jsonb key reordering must not break "Regenerate content" (rule compare).
3. A longer first name joining the audience between draft and approval must be caught at approval.
4. The same draft approved in two tabs: one approval wins, the other gets 200 unchanged.
5. Closing the review window with unsaved edits by Escape or outside click must ask first.

Also look for: correctness bugs, transaction/locking mistakes (e.g. Update running the audience count inside a held row lock, timeouts), trigger gaps, the never-execute-LLM-output-as-SQL and consent hard rules, personal data leaking to the browser or logs, error mapping, frontend state bugs (stale edits across campaigns, campaignId handling on regenerate/conflict, double submit), accessibility basics, and anything a real marketer would hit.

Report findings ranked by severity: Critical / Important / Minor. For each: file:line, the concrete failure scenario (input/state → wrong result), and a suggested fix. Grade by what a user of the software would experience, not by whether the spec mentions the input. Add a short "Declined to judg
… [75 more chars truncated]
```

<details><summary>Result</summary>

```
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a10c2c4399cb2143a (internal ID - do not mention to user. Use SendMessage with to: 'a10c2c4399cb2143a', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/c7414ebc-9dab-4372-82ec-2a9965506baa/tasks/a10c2c4399cb2143a.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

</details>

## 🤖 Claude  <sub>2026-11-10 12:05:26</sub>

All 9 tasks are built and `make verify` passes. A final reviewer is checking the whole slice now. I'll fix what it finds and then report.

## 🤖 Claude  <sub>2026-11-10 12:08:34</sub>

Reviewer found 1 Critical and 4 Important issues. Fixing them now, each with a failing test first.

**🔧 Bash** _Write failing tests for lock and snapshot findings_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend && cat >> internal/campaign/service_test.go <<'EOF'

// With one pool connection, the row lock must not wait for the audience query.
func TestService_ApproveWithOneConnection(t *testing.T) {
	ctx, pool, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "hi {{first_name}}")
	cfg := pool.Config().Copy()
	cfg.MaxConns = 1
	one, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer one.Close()
	svc.Store = NewStore(one)
	svc.Count = func(ctx context.Context, w segment.Where) (segment.Audience, error) {
		return segment.CountAudience(ctx, one, w)
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, _, err := svc.Approve(cctx, id, body("hi {{first_name}}")); err != nil {
		t.Fatalf("approve with one connection: %v", err)
	}
}

// A failed check stores the new audience snapshot, so the window can show the new limit.
func TestService_FailedApproveStoresNewLongestName(t *testing.T) {
	ctx, pool, svc := seededService(t)
	l := trAudience(t, ctx, svc, ankaraRule)
	m := int(l.MaxFirstName)
	sms := "{{first_name}}" + strings.Repeat("a", 160-m-1)
	id := makeDraft(t, ctx, svc, ankaraRule, sms)
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, first_name,last_name,email,phone,language,country,city,birth_date,email_consent,sms_consent,created_at,order_count,total_spent)
VALUES (1000001, $1,'K','long@example.com','+905000000001','tr','TR','Ankara','1990-01-01',true,true,now(),0,0)`, strings.Repeat("Ş", m+2)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Approve(ctx, id, body(sms)); err == nil {
		t.Fatal("want an error")
	}
	c, _ := svc.Get(ctx, id)
	if c.Status != StatusDraft || c.Contents[0].FirstNameRunes != m+2 || c.Contents[0].SMS != sms {
		t.Fatalf("after failed approve: status %s runes %d (want %d), sms changed %v", c.Status, c.Contents[0].FirstNameRunes, m+2, c.Contents[0].SMS != sms)
	}
}

// A repeat approval needs no audience query, so a database error there cannot turn it into a failure.
func TestService_RepeatApproveSkipsCount(t *testing.T) {
	ctx, _, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "hi")
	if _, _, err := svc.Approve(ctx, id, body("hi")); err != nil {
		t.Fatal(err)
	}
	svc.Count = func(context.Context, segment.Where) (segment.Audience, error) {
		return segment.Audience{}, context.DeadlineExceeded
	}
	if _, repeat, err := svc.Approve(ctx, id, body("hi")); err != nil || !repeat {
		t.Fatalf("repeat: %v %v", repeat, err)
	}
}
EOF
TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55433/test?sslmode=disable' go test -count=1 -run 'OneConnection|StoresNewLongest|RepeatApproveSkips' ./internal/campaign/ 2>&1 | grep -E "^(---|ok|FAIL)|service_test" | head
```

<details><summary>Result</summary>

```
--- FAIL: TestService_ApproveWithOneConnection (2.16s)
    service_test.go:237: approve with one connection: count audience: context deadline exceeded
--- FAIL: TestService_FailedApproveStoresNewLongestName (0.12s)
    service_test.go:257: after failed approve: status draft runes 7 (want 9), sms changed false
FAIL
FAIL	campaigncomposer/internal/campaign	2.932s
FAIL
```

</details>

**🔧 Bash** _Move the audience count out of the row lock and store snapshot on failure_
```bash
python3 - <<'EOF'
p='internal/campaign/service.go'
s=open(p).read()
def rep(a,b):
    global s
    assert a in s, a
    s=s.replace(a,b,1)
rep('''import (
	"context"
	"encoding/json"
	"time"
''','''import (
	"context"
	"encoding/json"
	"errors"
	"time"
''')
rep('''func (s *Service) Save(ctx context.Context, id int64, body json.RawMessage) (Campaign, error) {
	c, _, err := s.Store.Update(ctx, id, false, s.check(ctx, body))
	return c, err
}''','''func (s *Service) Save(ctx context.Context, id int64, body json.RawMessage) (Campaign, error) {
	c, _, err := s.update(ctx, id, false, body)
	return c, err
}''')
rep('''	return s.Store.Update(ctx, id, true, s.check(ctx, body))
}''','''	return s.update(ctx, id, true, body)
}''')
i=s.index('func (s *Service) check(')
j=s.index('// SampleNames returns')
s=s[:i]+'''// update counts the audience first, then locks the row for the check and the
// write only. Thus the row lock never waits for the audience query, and a
// second connection is not needed while the lock is held. The stored rule of a
// draft never changes (ReplaceDraft keeps it), so the count stays valid.
func (s *Service) update(ctx context.Context, id int64, approve bool, body json.RawMessage) (Campaign, bool, error) {
	c, err := s.Store.Get(ctx, id)
	if err != nil {
		return Campaign{}, false, err
	}
	if c.Status != StatusDraft {
		// A repeat approval or ErrNotDraft; Update decides without a check.
		return s.Store.Update(ctx, id, approve, func(Campaign) ([]LanguageContent, error) { return nil, ErrNotDraft })
	}
	w, err := s.compile(c)
	if err != nil {
		return Campaign{}, false, err
	}
	cctx, cancel := context.WithTimeout(ctx, countTimeout)
	aud, err := s.Count(cctx, w)
	cancel()
	if err != nil {
		return Campaign{}, false, err
	}
	byLang := map[string]segment.LanguageCounts{}
	for _, l := range aud.Languages {
		byLang[l.Code] = l
	}
	out, repeat, err := s.Store.Update(ctx, id, approve, func(locked Campaign) ([]LanguageContent, error) {
		langs := make([]string, 0, len(locked.Contents))
		for _, lc := range locked.Contents {
			langs = append(langs, lc.Language)
		}
		contents, errs := CheckContents(body, langs, byLang)
		if len(errs) > 0 {
			return nil, &InvalidError{Errors: errs}
		}
		return contents, nil
	})
	var inv *InvalidError
	if errors.As(err, &inv) {
		// Keep the new audience snapshot, so the review window shows the limit
		// that failed. The content stays as it was.
		_ = s.Store.UpdateCounts(ctx, id, byLang)
	}
	return out, repeat, err
}

'''+s[j:]
open(p,'w').write(s)

p='internal/campaign/store.go'
s=open(p).read()
s=s.replace('''	"github.com/jackc/pgx/v5/pgxpool"
)''','''	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/segment"
)''',1)
s+='''
// UpdateCounts stores the audience snapshot of a draft: the counts and the
// longest first name of each language. A language missing from aud gets 0.
// It changes nothing when the campaign is not a draft.
func (s *Store) UpdateCounts(ctx context.Context, id int64, aud 
… [1097 more chars truncated]
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/campaign	3.762s
ok  	campaigncomposer/internal/httpapi	3.553s
```

</details>

**🔧 Bash** _Write failing frontend tests for review findings_
```bash
cd ~/Desktop/projects/case-studies/insider_one/frontend && python3 - <<'EOF'
p='src/components/campaign/__tests__/review-dialog.test.ts'
s=open(p).read().rstrip()
assert s.endswith('})')
s=s[:-2]+"""
  it('loads again on each open when nothing is unsaved', async () => {
    const w = await make()
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe('Hi {{first_name}}')
    await w.setProps({ open: false })
    await flushPromises()
    const regenerated = campaign()
    regenerated.contents[0] = { ...regenerated.contents[0]!, sms: 'New {{first_name}}' }
    getC.mockResolvedValue(regenerated)
    await w.setProps({ open: true })
    await flushPromises()
    expect(getC).toHaveBeenCalledTimes(2)
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe('New {{first_name}}')
  })

  it('ignores a late response for another campaign', async () => {
    let resolveA: (c: Campaign) => void = () => {}
    getC.mockImplementationOnce(() => new Promise((r) => {
      resolveA = r
    }))
    const w = mount(ReviewDialog, { props: { campaignId: 1, open: true }, attachTo: document.body })
    await flushPromises()
    getC.mockResolvedValueOnce(campaign({ id: 2 }))
    await w.setProps({ campaignId: 2 })
    await flushPromises()
    const a = campaign({ id: 1 })
    a.contents[0] = { ...a.contents[0]!, sms: 'From A' }
    resolveA(a)
    await flushPromises()
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe('Hi {{first_name}}')
  })

  it('takes the new longest name after a 422 and keeps the edits', async () => {
    approve.mockImplementation(() => Promise.reject(new CampaignRequestError('content_invalid', 'Some messages are not valid.', 422, [
      { path: 'contents[1].sms', message: 'sms: 161 runes after rendering, limit 160' },
    ])))
    await make()
    await pressTab('[data-test="lang-tab"][data-value="tr"]')
    await type('[data-test="sms-input"]', `{{first_name}}${'a'.repeat(140)}`) // 154 with 14
    const longer = campaign()
    longer.contents[1] = { ...longer.contents[1]!, first_name_runes: 21 }
    getC.mockResolvedValue(longer)
    await click('[data-test="approve"]')
    await click('[data-test="send-now"]')
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe(`{{first_name}}${'a'.repeat(140)}`)
    expect(text('[data-test="sms-counter"]')).toBe('161 / 160')
    await type('[data-test="sms-input"]', `{{first_name}}${'a'.repeat(141)}`)
    expect($('[data-test="approve"]')!.hasAttribute('disabled')).toBe(true)
  })
})
"""
open(p,'w').write(s+"\n")

p='src/components/campaign/__tests__/content-cards.test.ts'
s=open(p).read().rstrip()
s=s[:-2]+"""
  it('hides the review button during a regenerate and after a failed save', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'campaign', data: { id: 7, status: 'draft' } })
    r.get()({ event: 'done', data: {} })
    r.finish()
    await 
… [1537 more chars truncated]
```

<details><summary>Result</summary>

```
× hides the review button during a regenerate and after a failed save 4ms
     × loads again on each open when nothing is unsaved 100ms
     × ignores a late response for another campaign 51ms
     × takes the new longest name after a 422 and keeps the edits 119ms
⎯⎯⎯⎯⎯⎯⎯ Failed Tests 4 ⎯⎯⎯⎯⎯⎯⎯
      Tests  4 failed | 43 passed (47)
```

</details>

**🔧 Bash** _Fix reload-on-open, stale responses and limit refresh in dialog_
```bash
python3 - <<'EOF'
p='src/components/campaign/review-dialog.vue'
s=open(p).read()
def rep(a,b):
    global s
    assert a in s, a
    s=s.replace(a,b,1)
rep('''let loadedFor = 0
''','''let loadedFor = 0
let loadToken = 0
''')
rep('''async function load() {
  loadError.value = ''
  loadedFor = props.campaignId
  try {
    reset(await getCampaign(props.campaignId))
  }
  catch (err) {
    loadError.value = err instanceof CampaignRequestError ? err.message : 'Something went wrong. Try again.'
    return
  }
  try {
    samples.value = await getSampleNames(props.campaignId)
    sampleFailed.value = false
  }
  catch {
    samples.value = {}
    sampleFailed.value = true
  }
}

// Load on open. Edits stay while the same campaign is open again.
watch([open, () => props.campaignId], ([o, id]) => {
  if (o && (loadedFor !== id || !campaign.value))
    void load()
}, { immediate: true })
''','''async function load() {
  const id = props.campaignId
  const mine = ++loadToken
  loadError.value = ''
  loadedFor = id
  try {
    const c = await getCampaign(id)
    if (mine !== loadToken)
      return
    reset(c)
  }
  catch (err) {
    if (mine === loadToken)
      loadError.value = err instanceof CampaignRequestError ? err.message : 'Something went wrong. Try again.'
    return
  }
  try {
    const names = await getSampleNames(id)
    if (mine !== loadToken)
      return
    samples.value = names
    sampleFailed.value = false
  }
  catch {
    if (mine !== loadToken)
      return
    samples.value = {}
    sampleFailed.value = true
  }
}

// Reloads the counts and limits after a failed check, and keeps the edits.
async function refreshLimits() {
  const id = props.campaignId
  const mine = loadToken
  try {
    const c = await getCampaign(id)
    if (mine === loadToken && id === props.campaignId)
      campaign.value = c
  }
  catch {
    // The old limits stay. The server check still decides.
  }
}

// Load on each open, so that a regenerated draft shows its new text. Unsaved
// edits of the same campaign stay.
watch([open, () => props.campaignId], ([o, id]) => {
  if (o && (loadedFor !== id || !campaign.value || !dirty.value))
    void load()
}, { immediate: true })
''')
rep('''    banner.value = rest.length > 0 ? `${err.message} ${rest.join(' ')}` : ''
    return''','''    banner.value = rest.length > 0 ? `${err.message} ${rest.join(' ')}` : ''
    await refreshLimits()
    return''')
rep('''function applyError(err: unknown, sent: ContentInput[]) {''','''async function applyError(err: unknown, sent: ContentInput[]) {''')
s=s.replace('''    applyError(err, sent)''','''    await applyError(err, sent)''')
open(p,'w').write(s)
EOF
grep -n "applyError\|const dirty" src/components/campaign/review-dialog.vue
```

<details><summary>Result</summary>

```
134:const dirty = computed(() => Object.keys(edits).some(l =>
200:async function applyError(err: unknown, sent: ContentInput[]) {
231:    await applyError(err, sent)
253:    await applyError(err, sent)
```

</details>

**🔧 Bash** _Fix campaign id handling in cards and rerun frontend checks_
```bash
python3 - <<'EOF'
p='src/components/campaign/review-dialog.vue'
s=open(p).read()
w_start=s.index('// Load on each open, so that')
w_end=s.index('}, { immediate: true })', w_start)+len('}, { immediate: true })\n')
block=s[w_start:w_end]
s=s[:w_start]+s[w_end:]
d=s.index('const dirty = computed(')
d_end=s.index('\n\n', d)+2
s=s[:d_end]+block+'\n'+s[d_end:]
open(p,'w').write(s)

p='src/components/campaign/content-cards.vue'
s=open(p).read()
def rep(a,b):
    global s
    assert a in s, a
    s=s.replace(a,b,1)
rep('''const campaignId = ref<number | null>(null)
''','''// campaignId is the draft of the current run; the review button needs it.
// draftId is the draft that "Regenerate content" replaces.
const campaignId = ref<number | null>(null)
let draftId: number | null = null
''')
rep('''  if (e.event === 'campaign') {
    campaignId.value = e.data.id
    return
  }
  if (e.event === 'error' && e.data.language === undefined) {
    requestError.value = e.data.message''','''  if (e.event === 'campaign') {
    campaignId.value = e.data.id
    draftId = e.data.id
    return
  }
  if (e.event === 'error' && e.data.language === undefined) {
    if (e.data.code === 'campaign_conflict')
      draftId = null
    requestError.value = e.data.message''')
rep('''async function start(regenerate = false) {
  if (!regenerate) {
    campaignId.value = null
    reviewOpen.value = false
  }
  const draftId = campaignId.value ?? undefined
  controller?.abort()''','''async function start(regenerate = false) {
  if (!regenerate)
    draftId = null
  campaignId.value = null
  reviewOpen.value = false
  const replaceId = draftId ?? undefined
  controller?.abort()''')
rep('''    }, ctl.signal, draftId)''','''    }, ctl.signal, replaceId)''')
open(p,'w').write(s)
EOF
pnpm vitest run 2>&1 | grep -E "×|Tests " | head; pnpm vue-tsc -b 2>&1 | tail -3; CI=true pnpm lint 2>&1 | tail -3
```

<details><summary>Result</summary>

```
Tests  146 passed (146)
  3 errors and 0 warnings potentially fixable with the `--fix` option.

[ELIFECYCLE] Command failed with exit code 1.
```

</details>

**🔧 Bash** _Autofix lint and run make verify after the fix pass_
```bash
pnpm eslint --fix src >/dev/null 2>&1; CI=true pnpm lint 2>&1 | tail -3; pnpm vitest run 2>&1 | grep -E "Tests "; cd .. && make verify > .superpowers/sdd/06-review-approve/verify-final.log 2>&1; echo rc=$?; grep -E "^(ok|FAIL)|Tests " .superpowers/sdd/06-review-approve/verify-final.log
```

<details><summary>Result</summary>

```
$ eslint .
      Tests  146 passed (146)
rc=0
ok  	campaigncomposer/cmd/api	(cached)
ok  	campaigncomposer/internal/campaign	1.913s
ok  	campaigncomposer/internal/config	(cached)
ok  	campaigncomposer/internal/content	(cached)
ok  	campaigncomposer/internal/httpapi	2.498s
ok  	campaigncomposer/internal/llm	(cached)
ok  	campaigncomposer/internal/mockprovider	(cached)
ok  	campaigncomposer/internal/seed	(cached)
ok  	campaigncomposer/internal/segment	(cached)
ok  	campaigncomposer/internal/store	(cached)
      Tests  146 passed (146)
```

</details>

**🔧 Bash** _Update docs and progress, clean up temp services and workspace_
```bash
python3 - <<'EOF'
p='docs/decisions-log.md'
s=open(p).read()
a='''- **Choice:** The approval runs `SELECT ... FOR UPDATE`, checks the status, writes the content and sets `approved` in one transaction.'''
assert a in s
s=s.replace(a,'''- **Choice:** The service counts the audience first, with no lock. Then the approval runs `SELECT ... FOR UPDATE`, checks the status, writes the content and sets `approved` in one transaction. The lock never waits for the audience query.''')
a='''The counts in the campaign update. The browser sends no rule after the draft exists.'''
assert a in s
s=s.replace(a,'''The counts in the campaign update. A failed check also stores the new counts, so the window shows the new limit. The browser sends no rule after the draft exists.''')
open(p,'w').write(s)
p='docs/architecture.md'
s=open(p).read()
a='''   - The approval locks the campaign row and checks the status in one transaction.'''
assert a in s
s=s.replace(a,'''   - The service counts the audience before it locks the row. The approval then locks the campaign row and checks the status in one transaction.''')
open(p,'w').write(s)
EOF
python3 .claude/skills/ste-writing/scripts/ste_check.py docs/decisions-log.md docs/architecture.md | awk -F: '$2>243 || /architecture/' | head; cat >> tasks/progress.md <<'EOF'

## 2026-11-10 — 06-review-approve
- Done: spec `docs/specs/06-review-approve/` (with the mockup), plan `docs/plans/06-review-approve.md` and Tasks 1 to 9, run inline. User decisions: one review window with SMS and Email tabs, a second confirm before the send, a token-based design that works on a phone, failed languages are skipped, and Claude saves a draft when it finishes. Added migration `00002_campaigns`, `segment.SampleFirstNames`, `internal/campaign` (store, check, service), five campaign endpoints, the draft save in the content stream, the review window, `/campaigns` and `/campaigns/{id}`. Added D-036 to D-045. The final review found 1 Critical and 4 Important defects. All 5 are fixed, each with a test that failed first: a stale draft text after "Regenerate content", the row lock that waited for the audience query, a late response for another campaign, old limits after a 422, and the review button after a failed regenerate.
- Verified with: `make verify` exit 0 (Go tests with Postgres, 146 frontend tests, lint, build). A compose run with `-p cc-verify` on ports 13000 and 18080 (the running stack holds 3000 and 8080) gave the draft, approve `200`, repeat `200`, PUT `409` and CORS with `PUT`. Headless Chrome at 1280 px and 360 px completed the flow to "Approved" with no horizontal scroll.
- Next: user review of slice 06. The user sets `06-review-approve.passes`. Then slice 07-send-queue.
- Open questions:
  - Five Minor review items stay open. See the final message of the session.
  - Slice 07 must widen the `campaigns.status` CHECK and keep `approved_at` set for the later statuses.
EOF
cat >> .superpowers/sdd/06-review-approve/progress.md <<'EOF'
Final: fixed Cr
… [1121 more chars truncated]
```

<details><summary>Result</summary>

```
37
```

</details>

## 🤖 Claude  <sub>2026-11-10 12:11:12</sub>

Slice 06 is built and `make verify` passes. Nothing is committed yet, as the plan says: the 37 changed files are on branch `06-review-approve`, ready for you to review and commit.

**What a marketer can now do:**
- When Claude finishes the content, the backend saves the campaign as a draft. A "Review & approve" button appears on `/campaigns/new`.
- `/campaigns` lists every campaign with a "Draft" or "Approved" badge. Clicking a draft opens the review window from the mockup. Clicking an approved one opens `/campaigns/{id}`.
- In the window, the marketer can:
  - switch between the SMS and Email tabs and the language tabs
  - watch the character counters, which turn orange near the limit and red over it
  - preview the message with real first names from the audience, longest name first
  - "Save draft", or "Approve & send", which opens the confirm popup before anything happens
- The backend checks every save and approval again against the current audience and the limits. Once approved, the database blocks any further change. The same draft can't be approved twice, even from two tabs at once.

**How I checked it:**
- `make verify`: the Go tests ran against Postgres, including 10 approvals of one draft at the same time, plus 146 frontend tests, lint and the build.
- I ran a separate copy of the stack on ports 13000 and 18080, because your running stack holds 3000 and 8080. In headless Chrome at 1280 px and 360 px, the full flow reached "Approved" with no horizontal scroll. The API calls behaved as the spec says: approve `200`, a repeat approval `200`, then a save `409`.

**The final review found 1 Critical and 4 Important problems. All five are fixed, each with a test that failed first:**
1. **Critical:** after "Regenerate content", reopening the window showed the old text, and approving would have sent it. The window now reloads each time it opens, unless there are unsaved edits.
2. While an approval held its database lock, it waited for the audience query on a second connection. Under load this could time out with an error. The audience is now counted before the lock is taken.
3. On a slow network, a late response for one campaign could fill in another campaign's window. Late responses are now ignored.
4. After an approval failed because a user with a longer first name had joined the audience, the counter still used the old limit. The new limit is now saved and shown, and the edits stay.
5. After a failed regenerate, "Review & approve" still opened the old draft. The button now stays hidden until the new draft is saved.

**Decisions I made during the build (D-036 to D-045 in `docs/decisions-log.md`):**
- A subject line break can't be typed, because the browser removes line breaks from a single-line input. The backend still rejects one.
- All ways of closing the window (close button, Cancel, Escape, outside click) ask "Discard your changes?" when there are unsaved edits.
- An approved campaign opened in the window is read-only, with just a "Close" button.
- If something other than the content stream saves drafts, it has to be connected explicitly. The app wires it in, and the compose run proved it works.

**Five minor review items I left open:**
- If the browser disconnects at the exact moment the draft is being saved, a draft can still appear in the list.
- The database rule that locks approved campaigns can be bypassed by an edit that runs at the same moment as the approval. Deleting an approved campaign directly in SQL is also not blocked.
- A repeated approval with a broken request body returns `400` instead of `200`.
- If a draft is approved in another tab, "Save draft" shows an error but leaves the editor open instead of switching to read-only.
- Screen readers don't announce the red error dot on a language tab.

Next: review the slice and commit it. Once you approve, I'll set `06-review-approve.passes` to `true`. Slice 07 (send queue) will need to allow the status values after `approved`.

## 👤 User  <sub>2026-11-10 12:11:58</sub>

commit and merge to main

**🔧 Bash** _Check non-code changes and merge style_
```bash
git status --short | grep -v "^ M backend\|^?? backend\|frontend" ; git log --oneline main -3; git log --oneline --merges -5 --all | head; git log --format='%s%n%b' -1 7e3be6f
```

<details><summary>Result</summary>

```
M .claude/agent-memory/reviewer/MEMORY.md
 M docs/architecture.md
 M docs/decisions-log.md
 M tasks/backlog.json
 M tasks/progress.md
?? .claude/agent-memory/reviewer/review-approve-defects.md
?? docs/plans/06-review-approve.md
?? docs/specs/06-review-approve/
1a73133 04 segment completed
61ee9c2 fix(httpapi): reject trailing data and oversized preview bodies
c197fcb chore: remove SDD scratch report from git
chore(05): list slice 05 acceptance checks in the backlog
Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
```

</details>

**🔧 Bash** _Check fast-forward and tracked agent memory_
```bash
git merge-base --is-ancestor main HEAD && echo ff-possible; git log --oneline main..HEAD | wc -l; git log -1 --format=%s%n%b 1a73133 | head -5; git ls-files .claude/agent-memory | head
```

<details><summary>Result</summary>

```
ff-possible
      14
04 segment completed

.claude/agent-memory/reviewer/MEMORY.md
.claude/agent-memory/reviewer/frontend-trim-defects.md
.claude/agent-memory/reviewer/go-startup-defects.md
.claude/agent-memory/reviewer/http-config-validation-defects.md
.claude/agent-memory/reviewer/make-compose-defects.md
.claude/agent-memory/reviewer/seed-generator-defects.md
.claude/agent-memory/reviewer/ste-check-edge-cases.md
.claude/agent-memory/reviewer/test-db-harness-defects.md
```

</details>

**🔧 Bash** _Commit slice 06 in four commits and fast-forward main_
```bash
python3 - <<'EOF'
import json
raw=open('tasks/backlog.json').read(); d=json.loads(raw)
items=d if isinstance(d,list) else d['slices']
for s in items:
    if s['id']=='06-review-approve': s['passes']=True
open('tasks/backlog.json','w').write(json.dumps(d,indent=2,ensure_ascii=False)+'\n')
EOF
T="

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
git add docs/specs/06-review-approve docs/plans/06-review-approve.md && git commit -q -m "docs(06): add the review and approve spec, mockup and plan$T" && \
git add backend && git commit -q -m "feat(06): save drafts and approve campaigns

The content stream saves a draft at the end. Regenerate replaces it.
Save and approve check the content again against the current audience.
A row lock stops a double approval. Triggers freeze an approved campaign.
The sample endpoint returns first names only.$T" && \
git add frontend && git commit -q -m "feat(frontend): review window, campaign list and campaign page

The review window has SMS and Email tabs, live counters, a preview with
sample names and a second confirm before the approval.$T" && \
git add docs/architecture.md docs/decisions-log.md tasks .claude/agent-memory && git commit -q -m "docs(06): architecture flow, D-036 to D-045, backlog and progress$T" && \
git status --short | wc -l && git log --oneline -5 && git checkout -q main && git merge --ff-only -q 06-review-approve && git log --oneline -1 && git branch --show-current
```

<details><summary>Result</summary>

```
0
40f5db3 docs(06): architecture flow, D-036 to D-045, backlog and progress
e4cc561 feat(frontend): review window, campaign list and campaign page
293b095 feat(06): save drafts and approve campaigns
091322c docs(06): add the review and approve spec, mockup and plan
7e3be6f chore(05): list slice 05 acceptance checks in the backlog
40f5db3 docs(06): architecture flow, D-036 to D-045, backlog and progress
main
```

</details>

## 🤖 Claude  <sub>2026-11-10 12:12:23</sub>

I committed slice 06 and merged it into `main` on your machine. Nothing is pushed yet, so `origin/main` is still at the slice 04 commit.

The work is in four commits on `06-review-approve`:
- `091322c` docs(06): the spec, the mockup and the plan
- `293b095` feat(06): saving drafts and approving campaigns in the backend
- `e4cc561` feat(frontend): the review window, the campaign list and the campaign page
- `40f5db3` docs(06): the architecture flow, decisions D-036 to D-045, the backlog and the progress log

- **Merge:** `main` had no commits of its own since the branch started, so the merge was a fast-forward with no merge commit. You're now on `main` at `40f5db3`.
- **Slice 05 is included:** `06-review-approve` was branched from `05-content-generation`, which was never merged into `main`. So this merge also brought slice 05's commits into `main`.
- **Backlog:** I set `06-review-approve.passes` to `true` in the last docs commit, taking your request to merge as your approval of the slice.
- **Not re-run:** `make verify` last passed after the review fixes, before the commits. The only change made after that run is the backlog flag, and I didn't run it again on `main`.

Do you want me to push `main`, and delete the merged `05-content-generation` and `06-review-approve` branches?


---

## Subagent transcripts


### Subagent: Builder check on slice 06 spec (general-purpose)

#### 👤 User  <sub>2026-11-10 11:36:11</sub>

You are a talented full-stack software engineer (Go backend, Vue 3 + shadcn-vue frontend). You've been handed a design document and asked to put together a plan to build what it describes. You will not be able to talk to the person who wrote it.

Read ~/Desktop/projects/case-studies/insider_one/docs/specs/06-review-approve/goal.md. You may read the repo it lives in (CLAUDE.md, docs/specs/04-*/ and 05-*/, docs/decisions-log.md, backend/, frontend/src/) for context. Do not edit anything.

Report only the questions you'd still need answered before you could plan this well: places where you'd have to guess and a wrong guess would matter. For each, give the question and why the answer changes the plan. Rank them by importance.

Skip anything the document explicitly leaves to the builder, and anything you could reasonably decide yourself. Skip nitpicks about wording.

If nothing important is left, say "nothing important left".

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one; ls docs/specs/*; cat docs/specs/06-review-approve/goal.md; cat tasks/backlog.json | head -150
```

<details><summary>Result</summary>

```
docs/specs/01-frontend-trim:
goal.md
requirements.md

docs/specs/02-seed-users:
goal.md
requirements.md

docs/specs/03-mock-provider:
goal.md
requirements.md

docs/specs/04-segment-rules:
goal.md
requirements.md

docs/specs/05-content-generation:
goal.md
requirements.md

docs/specs/06-review-approve:
goal.md
review-mockup.html
# Goal: Review and approve

## Why

The case flow has five steps. Step 3 says that the marketer previews the content, edits it if necessary, and approves it. Claude suggests, and the human makes the last decision.

This slice builds step 3. The marketer opens a review window, checks each message with a sample user, edits the text, and approves. A second confirm popup comes before the approval. The backend then stores the campaign. Slice 07 sends the stored campaign.

## Input

- The case brief, section "Kullanıcı akışı" step 3, and the sections "Kitle" and "İçerik".
- Slice 04: the brief sanitization, the rule validator, the audience query. See `docs/specs/04-segment-rules/goal.md`.
- Slice 05: the content types, the variable parser, the renderer, the validator, and the content cards on `/campaigns/new`. See `docs/specs/05-content-generation/goal.md`.
- The mockup that the user accepted on 2026-11-10. A copy is in `docs/specs/06-review-approve/review-mockup.html`.
- User decisions of 2026-11-10:
  - The review is one window. The marketer reviews, edits and approves in it.
  - SMS and Email are tabs. The Email tab shows the subject and the body.
  - One button "Approve & send" approves the campaign. It opens a second confirm popup first.
  - The design is clean, uses design tokens, works on a phone and is easy to use.
  - If the content of a language failed in slice 05, the marketer can still approve. The campaign skips that language.
  - A later slice can add a retry for one failed language. This slice does not.

## Output

1. A migration that adds the campaign tables.
2. A Go package `internal/campaign`. It validates an approval request and s
… [27030 more chars truncated]
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one; cat docs/specs/05-content-generation/goal.md; ls backend/internal/*; ls backend/migrations 2>/dev/null || find backend -name "*.sql"
```

<details><summary>Result</summary>

```
# Goal: Content generation

## Why

The case flow has five steps. Step 2 says that Claude writes the SMS and the email, and that the content streams onto the page. The case also gives two limits: 160 characters for SMS and 80 characters for the email subject. The content supports `{{first_name}}` variables.

This slice builds step 2. Claude writes a draft for each language of the audience. The backend checks the draft against the limits after it fills in the variables. The operator sees the text arrive while Claude writes it, and sees the final text only after the backend accepts it.

## Input

- The case brief, sections "Kullanıcı akışı" step 2 and "İçerik".
- Slice 04: the brief sanitization, the rule validator, the rule description, the audience count, `internal/llm` and the panel on `/campaigns/new`. See `docs/specs/04-segment-rules/goal.md`.
- D-022: the segment response has a `languages` array. This slice uses it.
- User decisions of 2026-11-10:
  - The steps run in sequence. Content generation needs the result of the segment step.
  - The operator clicks one time. Content generation starts automatically when the segment step gives an audience.
  - The tone and style instructions in the brief go into the content prompt. There is no separate tone step or tone field. Example: "sıcak ve samimi".
  - The email body is plain text in this slice. A later slice can add formatting.

## Output

1. A Go package `internal/content`. It has the content types, the variable parser, the renderer, the validator, the prompt, the tool schema and the mock answers.
2. A streaming method in `internal/llm`, for the Anthropic client and the mock client.
3. The endpoint `POST /api/content/generate` in `internal/httpapi`. It streams Server-Sent Events (SSE).
4. On `/campaigns/new`: one content card for each language. The card shows the SMS, the email subject and the email body. The text fills in while Claude writes it.

This slice does not store content. Slice 06 adds the campaign recor
… [15175 more chars truncated]
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one; wc -c docs/specs/06-review-approve/review-mockup.html; sed -e 's/<style>.*<\/style>//' docs/specs/06-review-approve/review-mockup.html | grep -v '^\s*$' | head -400; grep -n "Header\|Allow" backend/internal/httpapi/cors.go
```

<details><summary>Result</summary>

```
16351 docs/specs/06-review-approve/review-mockup.html
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Review campaign</title>
<style>
  /* ---------- Tokens (same names as shadcn-vue in the app) ---------- */
  :root {
    --background: #ffffff;
    --foreground: #0a0a0a;
    --muted: #f4f4f5;
    --muted-foreground: #71717a;
    --border: #e4e4e7;
    --input: #e4e4e7;
    --primary: #18181b;
    --primary-foreground: #fafafa;
    --accent: #f4f4f5;
    --destructive: #dc2626;
    --success: #16a34a;
    --warning: #d97706;
    --overlay: rgb(0 0 0 / .45);
    --ring: #a1a1aa;
    --bubble: #e9e9eb;
    --radius: 10px;
    --radius-sm: 6px;
    --space-1: 4px; --space-2: 8px; --space-3: 12px; --space-4: 16px; --space-6: 24px;
    --text-xs: 12px; --text-sm: 14px; --text-base: 15px; --text-lg: 18px;
    --shadow: 0 10px 30px rgb(0 0 0 / .12), 0 2px 6px rgb(0 0 0 / .06);
    --font: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --background: #0b0b0d; --foreground: #fafafa; --muted: #18181b;
      --muted-foreground: #a1a1aa; --border: #27272a; --input: #27272a;
      --primary: #fafafa; --primary-foreground: #18181b; --accent: #1f1f23;
      --overlay: rgb(0 0 0 / .65); --bubble: #2a2a2e;
    }
  }
  * { box-sizing: border-box; }
  body { margin: 0; font-family: var(--font); font-size: var(--text-base); color: var(--foreground); background: var(--muted); }
  /* Page behind the dialog */
  .page { padding: var(--space-6); max-width: 960px; margin: 0 auto; opacity: .5; }
  .page h1 { font-size: 22px; margin: 0 0 var(--space-2); }
  .page .stub { height: 80px; border-radius: var(--radius); background: var(--background); border: 1px solid var(--border); margin-bottom: var(--space-3); }
  .overlay { position: fixed; inset: 0; background: var(--overlay); display: grid; place-items: center; paddin
… [14861 more chars truncated]
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one; git diff tasks/backlog.json; python3 -c "
import json;d=json.load(open('tasks/backlog.json'))
for s in d[6:]: print(json.dumps(s,indent=1))"; ls docs/plans; tail -40 tasks/progress.md; grep -n "^## D-0" docs/decisions-log.md | tail -15
```

<details><summary>Result</summary>

```
diff --git a/tasks/backlog.json b/tasks/backlog.json
index f79475f..b1a1716 100644
--- a/tasks/backlog.json
+++ b/tasks/backlog.json
@@ -187,7 +187,7 @@
       "NFR-9. The SSE response works with the api server settings. No server write timeout cuts a stream of 90 seconds.",
       "NFR-10. make verify passes."
     ],
-    "passes": false
+    "passes": true
   },
   {
     "id": "06-review-approve",
{
 "id": "06-review-approve",
 "title": "Review and approve",
 "spec": null,
 "acceptance": [],
 "passes": false
}
{
 "id": "07-send-queue",
 "title": "Send queue",
 "spec": null,
 "acceptance": [],
 "passes": false
}
{
 "id": "08-dashboard-metrics",
 "title": "Dashboard and metrics",
 "spec": null,
 "acceptance": [],
 "passes": false
}
00-skeleton.md
01-frontend-trim.md
02-seed-users.md
03-mock-provider.md
04-segment-rules.md
05-content-generation.md
- Next: Task 2 (unused UI components, packages and files). The template `frontend/README.md` must be replaced, because it names removed code (NFR-5).
- Open questions: `__pycache__/` in `.gitignore` (from slice 00).

## 2026-11-06 — 01-frontend-trim
- Done: Task 2 of `docs/plans/01-frontend-trim.md`. Deleted 29 UI components (kept the 34 on the list and `kbd`). Deleted the decision 3 files that no kept file uses. Removed 10 packages. `universal-cookie` stays, because `useCookies` in `default.vue` needs it as a peer. Replaced the template `frontend/README.md` and the `index.html` title (NFR-5). The reviewer found 2 defects. Both are fixed: a coverage glob for the deleted `src/services/`, and theme-editor exports in `constants/themes.ts`.
- Verified with: `pnpm test:run` (55 tests), `pnpm build`, `CI=true pnpm lint`. The NFR-4 import grep has no matches. The NFR-5 text grep matches only the negative assertions in `__tests__`. `pnpm peers check` shows only the two issues that the template also has.
- Next: Task 3 (frontend in Docker and compose, pnpm in the Makefile, delete the template).
- Open questions: the Task 2 Step 4 
… [7633 more chars truncated]
```

</details>

**🔧 SubagentHandback** ```json
{
  "message": "The spec is detailed. I found three questions that change the plan. Everything else is either assigned to the builder or something I can decide and log in `docs/decisions-log.md`.\n\n**1. Who writes the acceptance checks for slice 06?**\n- Every earlier spec folder has a `requirements.md` with numbered FR and NFR checks, and the plan and its tests are built from that list. `docs/specs/06-review-approve/` has only `goal.md` and the mockup.\n- The `06-review-approve` entry in `tasks/backlog.json` has `\"spec\": null` and `\"acceptance\": []`. `passes` can only go to `true` against that list.\n- Why it matters: if I write `requirements.md` from `goal.md`, that is an extra task at the start of the plan and the user must approve the list first. If the user writes it, I should wait, because the checks decide the test tasks and which edge cases get tests. Examples: the concurrent idempotency-key race, the `campaign_contents` update block, and the 360 px no-scroll check.\n\n**2. Can the marketer leave out a language, and what happens when a language has no reachable users left at approval?**\n- The spec skips a language only when its content failed in slice 05. The review window has no \"leave out this language\" control. If the marketer does not like the English copy, the only choice is to edit it.\n- Approval also re-runs the audience. If a language that has a `final` event now has 0 reachable users, the backend returns `422 content_invalid` on `contents[i].language`. That error has no field to show under. The UI gives no way to remove the language, so the marketer must click \"Find audience\" again, which deletes all edits and makes a new idempotency key.\n- Why it matters: if the answer is \"a per-language include/skip toggle\", that is new UI, new confirm-popup summary logic and a change to the \"skipped language\" footer text. If the answer is \"the backend quietly skips such a language\", the validation rules and the error contract change. If the spec is right as written, I only need an error banner for errors that have no field. Each answer leads to a different frontend and validation task list.\n\n**3. The mockup shows personal data that the spec does not allow. What should the preview show instead?**\n- The approved mockup's \"Preview as\" select shows surnames (\"Ayşe Yılmaz\", \"Abdurrahmangazi K.\"). The email preview has a \"To: <name>@example.com\" line and a fixed sender (\"ORNEKSHOP\", \"Örnek Shop <hello@ornek.com>\").\n- The spec says only first names reach the browser, and the sample query reads only `first_name`, `language` and the consent columns.\n- Why it matters: if the user wants the mockup's look, the sample endpoint must return more fields, which breaks a stated privacy rule. If not, I drop the surname and the \"To:\" line, or I make a fake address from the first name. The sender name and address also have no source yet: no env var, no config, and the spec allows no new env var. I need to know whether a hard-co
… [477 more chars truncated]
```

<details><summary>Result</summary>

```
{"success":true,"message":"Report delivered to your caller."}
```

</details>


### Subagent: Final whole-slice review (reviewer)

#### 👤 User  <sub>2026-11-10 12:05:24</sub>

You are a senior code reviewer. Review the finished slice 06 (review and approve) of the repo at ~/Desktop/projects/case-studies/insider_one. Read-only: do not edit any file.

The work is UNCOMMITTED on branch 06-review-approve. See it with `git status --short` and `git diff` (modified files) plus the untracked files listed by git status (backend/internal/campaign/, backend/internal/httpapi/campaigns*.go, content_draft_test.go, backend/internal/segment/sample*.go, backend/internal/store/migrations/00002_campaigns.sql, frontend/src/components/campaign/review-dialog.vue, message-preview.vue, frontend/src/lib/campaigns.ts, limits.ts, pages/campaigns/*, and their tests).

Authority, in order: CLAUDE.md hard rules; spec docs/specs/06-review-approve/goal.md and requirements.md; plan docs/plans/06-review-approve.md. The implementer's rulings (deliberate deviations) are the lines containing "Ruling:" in .superpowers/sdd/06-review-approve/progress.md — weigh them, flag any you disagree with.

What was built: the content SSE endpoint saves a draft campaign at the end (campaign event), regenerate replaces it; campaign store with row-lock approval and DB triggers that freeze approved campaigns; save/approve endpoints re-validate content against a fresh audience count; sample first names endpoint; Vue review dialog (SMS/Email tabs, language tabs, counters, preview, confirm popup, discard prompt), campaign list and campaign page.

Verification already done: `make verify` exit 0 (Go tests with Postgres incl. 10-way concurrent approve; 141 frontend tests; lint; build); a compose stack run with headless Chrome at 1280 px and 360 px completed the full flow with no horizontal scroll.

Check these deliberately (the plan's Review Focus — conditions the tests may not fully exercise):
1. An aborted/old content stream must not save a draft after the browser left (FR-6).
2. Postgres jsonb key reordering must not break "Regenerate content" (rule compare).
3. A longer first name joining the audience between draft and approval must be caught at approval.
4. The same draft approved in two tabs: one approval wins, the other gets 200 unchanged.
5. Closing the review window with unsaved edits by Escape or outside click must ask first.

Also look for: correctness bugs, transaction/locking mistakes (e.g. Update running the audience count inside a held row lock, timeouts), trigger gaps, the never-execute-LLM-output-as-SQL and consent hard rules, personal data leaking to the browser or logs, error mapping, frontend state bugs (stale edits across campaigns, campaignId handling on regenerate/conflict, double submit), accessibility basics, and anything a real marketer would hit.

Report findings ranked by severity: Critical / Important / Minor. For each: file:line, the concrete failure scenario (input/state → wrong result), and a suggested fix. Grade by what a user of the software would experience, not by whether the spec mentions the input. Add a short "Declined to judge" list for anything you could not decide. Keep the report under 900 words.

**🔧 Bash** _Read spec, plan, rulings_
```bash
cat docs/specs/06-review-approve/*.md docs/plans/06-review-approve.md | head -400; grep -n "Ruling:" .superpowers/sdd/06-review-approve/progress.md
```

<details><summary>Result</summary>

```
# Goal: Review and approve

## Why

The case flow has five steps. Step 3 says that the marketer previews the content, edits it if necessary, and approves it. Claude suggests, and the human makes the last decision.

This slice builds step 3. When Claude finishes the content, the backend saves the campaign as a draft. The marketer opens the draft in a review window, checks each message with a sample user, edits the text, and approves. A second confirm popup comes before the approval. Slice 07 sends the approved campaign.

## Input

- The case brief, section "Kullanıcı akışı" step 3, and the sections "Kitle" and "İçerik".
- Slice 04: the brief sanitization, the rule validator, the audience query. See `docs/specs/04-segment-rules/goal.md`.
- Slice 05: the content endpoint, the content types, the variable parser, the renderer, the validator, and the content cards on `/campaigns/new`. See `docs/specs/05-content-generation/goal.md`.
- The mockup that the user accepted on 2026-11-10. A copy is in `docs/specs/06-review-approve/review-mockup.html`.
- User decisions of 2026-11-10:
  - The review is one window. The marketer reviews, edits and approves in it.
  - SMS and Email are tabs. The Email tab shows the subject and the body.
  - One button "Approve & send" approves the campaign. It opens a second confirm popup first.
  - The design is clean, uses design tokens, works on a phone and is easy to use.
  - If the content of a language failed in slice 05, the marketer can still approve. The campaign skips that language.
  - A later slice can add a retry for one failed language. This slice does not.
  - Campaigns have drafts. When Claude finishes the content, the backend creates a draft.
  - A click on a draft opens its review window. The marketer approves the draft there, and then the campaign goes out.

## Output

1. A migration that adds the campaign tables.
2. A Go package `internal/campaign`. It creates, reads, updates and approves campaigns.
3. A change to `POST /api/conte
… [23550 more chars truncated]
```

</details>

**🔧 Bash** _Rest of requirements, focus, file sizes_
```bash
sed -n 60,200p docs/specs/06-review-approve/requirements.md; grep -n -A40 "Review Focus" docs/plans/06-review-approve.md | head -60; ls backend/internal/campaign; wc -l backend/internal/campaign/* backend/internal/httpapi/campaigns.go backend/internal/segment/sample.go backend/internal/store/migrations/00002_campaigns.sql frontend/src/components/campaign/review-dialog.vue frontend/src/lib/campaigns.ts frontend/src/lib/limits.ts frontend/src/pages/campaigns/*.vue
```

<details><summary>Result</summary>

```
- **FR-36.** "Save draft" sends one `PUT` with all languages and shows "Draft saved." on success.
- **FR-37.** "Cancel" with unsaved edits asks "Discard your changes?". With no edits, it closes at once.
- **FR-38.** "Approve & send" opens the confirm popup. "Back" closes it and keeps the edits. "Send now" sends one approval with the edited content of all languages.
- **FR-39.** While the approval runs, "Send now" and "Back" are off. A double click sends one request.
- **FR-40.** On `200`, the page goes to `/campaigns/{id}`.
- **FR-41.** On `422 content_invalid`, each error shows under its field, on the correct language tab. An error with no field shows in a banner.
- **FR-42.** For an approved campaign, the review window shows the content read-only, with no "Save draft" and no "Approve & send".
- **FR-43.** `/campaigns` lists the campaigns with the number, the description, the languages, the status badge and the time. A click on a draft opens the review window. A click on an approved campaign goes to `/campaigns/{id}`. With no campaigns, the page shows "No campaigns yet".
- **FR-44.** `/campaigns/{id}` shows the campaign number, the status badge, the description and one read-only card for each language with its counts. A draft also shows "Review & approve". A `404` shows "Campaign not found" and a link to `/campaigns`.

## Non-functional requirements

- **NFR-1.** Tests use the mock LLM. No test reads `ANTHROPIC_API_KEY`.
- **NFR-2.** Logs contain no brief text, no content text and no API key.
- **NFR-3.** Save and approve use `content.Validate` and `content.Render` from slice 05. The browser counter has a unit test with the same cases as FR-9 of slice 05.
- **NFR-4.** `internal/campaign` imports nothing from `internal/` except `content`, `segment` and `store`.
- **NFR-5.** The campaign tables store no user row. Only the sample endpoint sends first names to the browser. No endpoint sends a surname, an email address or a phone number.
- **NFR-6.** The review window a
… [6946 more chars truncated]
```

</details>

**🔧 Bash** _Read migration, store, service_
```bash
cat -n backend/internal/store/migrations/00002_campaigns.sql backend/internal/campaign/store.go backend/internal/campaign/service.go
```

<details><summary>Result</summary>

```
1	-- +goose Up
     2	CREATE TABLE campaigns (
     3	    id                bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
     4	    brief             text NOT NULL,
     5	    rule              jsonb NOT NULL,
     6	    description       text NOT NULL,
     7	    status            text NOT NULL CHECK (status IN ('draft', 'approved')),
     8	    skipped_languages text[] NOT NULL DEFAULT '{}',
     9	    created_at        timestamptz NOT NULL,
    10	    updated_at        timestamptz NOT NULL,
    11	    approved_at       timestamptz,
    12	    CHECK ((status = 'draft') = (approved_at IS NULL))
    13	);
    14	
    15	CREATE INDEX campaigns_created_at_idx ON campaigns (created_at DESC, id DESC);
    16	
    17	CREATE TABLE campaign_contents (
    18	    campaign_id      bigint NOT NULL REFERENCES campaigns (id) ON DELETE CASCADE,
    19	    language         text NOT NULL CHECK (language IN ('tr', 'en')),
    20	    sms              text NOT NULL,
    21	    email_subject    text NOT NULL,
    22	    email_body       text NOT NULL,
    23	    sms_count        bigint NOT NULL CHECK (sms_count >= 0),
    24	    email_count      bigint NOT NULL CHECK (email_count >= 0),
    25	    first_name_runes integer NOT NULL CHECK (first_name_runes >= 0),
    26	    PRIMARY KEY (campaign_id, language)
    27	);
    28	
    29	-- Only the content of a draft can change.
    30	-- +goose StatementBegin
    31	CREATE FUNCTION campaign_contents_draft_only() RETURNS trigger AS $$
    32	DECLARE
    33	    st text;
    34	BEGIN
    35	    SELECT status INTO st FROM campaigns
    36	    WHERE id = CASE WHEN TG_OP = 'DELETE' THEN OLD.campaign_id ELSE NEW.campaign_id END;
    37	    IF st IS NOT NULL AND st <> 'draft' THEN
    38	        RAISE EXCEPTION 'campaign content cannot change after approval';
    39	    END IF;
    40	    IF TG_OP = 'UPDATE' AND NEW.campaign_id <> OLD.campaign_id THEN
    41	        RAISE EXCEPTION 'campaign content cannot move to another campaign';
    4
… [16109 more chars truncated]
```

</details>

**🔧 Bash** _Read check, types, handlers_
```bash
cat -n backend/internal/campaign/check.go backend/internal/campaign/campaign.go backend/internal/httpapi/campaigns.go
```

**🔧 Bash** _Diff of modified backend files_
```bash
git diff backend/ | head -500
```

<details><summary>Result</summary>

```
1	package campaign
     2	
     3	import (
     4		"bytes"
     5		"encoding/json"
     6		"fmt"
     7		"sort"
     8		"strings"
     9	
    10		"campaigncomposer/internal/content"
    11		"campaigncomposer/internal/segment"
    12	)
    13	
    14	// InvalidError carries the validation errors of a save or an approval.
    15	type InvalidError struct{ Errors []content.FieldError }
    16	
    17	func (e *InvalidError) Error() string {
    18		parts := make([]string, 0, len(e.Errors))
    19		for _, fe := range e.Errors {
    20			parts = append(parts, fe.Path+": "+fe.Message)
    21		}
    22		return "campaign: invalid content: " + strings.Join(parts, "; ")
    23	}
    24	
    25	var entryKeys = map[string]bool{"language": true, "sms": true, "email_subject": true, "email_body": true}
    26	
    27	// CheckContents decodes body ({"contents": [...]}) and checks it against the
    28	// languages of the draft and the current audience. It returns the trimmed
    29	// contents with the new counts and first_name_runes, sorted by language.
    30	func CheckContents(body json.RawMessage, draftLangs []string, aud map[string]segment.LanguageCounts) ([]LanguageContent, []content.FieldError) {
    31		var errs []content.FieldError
    32		fail := func(path, format string, args ...any) {
    33			errs = append(errs, content.FieldError{Path: path, Message: fmt.Sprintf(format, args...)})
    34		}
    35	
    36		var top map[string]json.RawMessage
    37		if t := bytes.TrimSpace(body); len(t) == 0 || t[0] != '{' || json.Unmarshal(body, &top) != nil {
    38			fail("", "must be a JSON object")
    39			return nil, errs
    40		}
    41		var entries []json.RawMessage
    42		if raw, ok := top["contents"]; !ok || json.Unmarshal(raw, &entries) != nil || len(entries) == 0 {
    43			fail("contents", "must be an array with one entry for each language")
    44			return nil, errs
    45		}
    46	
    47		inDraft := map[string]bool{}
    48		for _, l := range draftLangs {
    49			inD
… [11425 more chars truncated]
```

</details>

<details><summary>Result</summary>

```
diff --git a/backend/cmd/api/main.go b/backend/cmd/api/main.go
index a1b4710..cc7010d 100644
--- a/backend/cmd/api/main.go
+++ b/backend/cmd/api/main.go
@@ -13,6 +13,7 @@ import (
 
 	"github.com/jackc/pgx/v5/pgxpool"
 
+	"campaigncomposer/internal/campaign"
 	"campaigncomposer/internal/config"
 	"campaigncomposer/internal/httpapi"
 	"campaigncomposer/internal/llm"
@@ -80,7 +81,19 @@ func run() error {
 		Now:        time.Now,
 		LLMTimeout: 20 * time.Second,
 	})
-	httpapi.RegisterContent(mux, newContentGenerator(llmClient, segment.Default(), count))
+	campaigns := campaign.NewStore(pool)
+	gen := newContentGenerator(llmClient, segment.Default(), count)
+	gen.Drafts = campaigns
+	httpapi.RegisterContent(mux, gen)
+	httpapi.RegisterCampaigns(mux, &campaign.Service{
+		Store: campaigns,
+		Allow: segment.Default(),
+		Count: campaign.AudienceCounter(count),
+		Sample: func(ctx context.Context, w segment.Where, langs []string, n int) (map[string][]string, error) {
+			return segment.SampleFirstNames(ctx, pool, w, langs, n)
+		},
+		Now: time.Now,
+	})
 
 	// No WriteTimeout: the content SSE stream can last 90 s (NFR-9).
 	srv := &http.Server{
diff --git a/backend/internal/httpapi/body.go b/backend/internal/httpapi/body.go
index 39aa5c0..d7da370 100644
--- a/backend/internal/httpapi/body.go
+++ b/backend/internal/httpapi/body.go
@@ -16,7 +16,12 @@ var errTrailingData = errors.New("httpapi: trailing data after JSON body")
 // extra bytes come after a valid value, for example whitespace padding.
 // Data after the value, other than whitespace, also fails.
 func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
-	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
+	return decodeBodyLimit(w, r, v, maxBodyBytes)
+}
+
+// decodeBodyLimit is decodeBody with a given byte limit.
+func decodeBodyLimit(w http.ResponseWriter, r *http.Request, v any, limit int64) error {
+	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))
 	if err != ni
… [11472 more chars truncated]
```

</details>

**🔧 Bash** _Pool config, sample, count_
```bash
sed -n 30,80p backend/cmd/api/main.go; grep -n "MaxConns\|pool_max" -r backend | head; cat -n backend/internal/segment/sample.go; grep -n "func CountAudience" -A30 backend/internal/segment/*.go | head -50
```

<details><summary>Result</summary>

```
slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("api stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// pgxpool.New connects lazily. It fails only on an invalid DATABASE_URL.
	// If Postgres is down, the process keeps running and /readyz reports it.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	// Bound the startup ping, so that an unreachable host cannot delay the listener.
	pingCtx, cancelPing := context.WithTimeout(ctx, startupPingTimeout)
	err = pool.Ping(pingCtx)
	cancelPing()
	if err != nil {
		slog.Warn("initial db ping failed", "err", err)
	}

	mux := http.NewServeMux()
	httpapi.RegisterHealth(mux, pool)

	llmClient, err := llm.New(cfg, mockDefault)
	if err != nil {
		return err
	}
	if m, ok := llmClient.(*llm.Mock); ok {
		// Slow, small fragments, so that the demo shows the stream.
		m.SetStream(16, 30*time.Millisecond)
	}
	slog.Info("llm configured", "model", cfg.AnthropicModel, "llm_mode", cfg.LLMMode)
	count := func(ctx context.Context, w segment.Where) (segment.Audience, error) {
		return segment.CountAudience(ctx, pool, w)
	}
	httpapi.RegisterSegments(mux, &httpapi.SegmentPreviewer{
		LLM:        llmClient,
		Allow:      segment.Default(),
		Count:      count,
     1	package segment
     2	
     3	import (
     4		"context"
     5		"fmt"
     6	
     7		"github.com/jackc/pgx/v5"
     8		"github.com/jackc/pgx/v5/pgxpool"
     9	)
    10	
    11	// sampleQuery builds the sample query for w. The language list and the limit
    12	// take the two placeholders after the arguments of w. It reads only id,
    13	// first_name, language, the consent columns and the columns of the rule.
    14	func sampleQuery(w Where, _ int) string {
    15		langArg := len(w.Args()) + 1
    16		limitArg := langArg +
… [4236 more chars truncated]
```

</details>

**🔧 Bash** _Read concurrent approve test_
```bash
grep -n "Concurrent" -A60 backend/internal/campaign/service_test.go | head -80; grep -rn "func testPool\|pgxpool.New\|MaxConns" backend/internal/campaign/*_test.go backend/internal/store/*.go | head
```

<details><summary>Result</summary>

```
145:func TestService_ApproveConcurrent(t *testing.T) {
146-	ctx, _, svc := seededService(t)
147-	id := makeDraft(t, ctx, svc, ankaraRule, "hi {{first_name}}")
148-	var wg sync.WaitGroup
149-	var mu sync.Mutex
150-	firsts := 0
151-	approvedAt := map[string]bool{}
152-	for i := 0; i < 10; i++ {
153-		wg.Add(1)
154-		go func() {
155-			defer wg.Done()
156-			c, repeat, err := svc.Approve(ctx, id, body(fmt.Sprintf("hi %d {{first_name}}", i)))
157-			mu.Lock()
158-			defer mu.Unlock()
159-			if err != nil {
160-				t.Errorf("approve: %v", err)
161-				return
162-			}
163-			if c.ID != id {
164-				t.Errorf("id = %d", c.ID)
165-			}
166-			if !repeat {
167-				firsts++
168-			}
169-			approvedAt[c.ApprovedAt.String()] = true
170-		}()
171-	}
172-	wg.Wait()
173-	if firsts != 1 || len(approvedAt) != 1 {
174-		t.Fatalf("first approvals = %d, approved_at values = %d; want 1 and 1", firsts, len(approvedAt))
175-	}
176-}
177-
178-func TestService_SaveApproved(t *testing.T) {
179-	ctx, _, svc := seededService(t)
180-	id := makeDraft(t, ctx, svc, ankaraRule, "hi")
181-	if _, _, err := svc.Approve(ctx, id, body("hi")); err != nil {
182-		t.Fatal(err)
183-	}
184-	if _, err := svc.Save(ctx, id, body("hi")); !errors.Is(err, ErrNotDraft) {
185-		t.Fatalf("err = %v, want ErrNotDraft", err)
186-	}
187-}
188-
189-func TestService_DBTimeout(t *testing.T) {
190-	ctx, _, svc := seededService(t)
191-	id := makeDraft(t, ctx, svc, ankaraRule, "hi")
192-	svc.Count = func(context.Context, segment.Where) (segment.Audience, error) {
193-		return segment.Audience{}, context.DeadlineExceeded
194-	}
195-	if _, _, err := svc.Approve(ctx, id, body("changed")); !errors.Is(err, context.DeadlineExceeded) {
196-		t.Fatalf("err = %v, want DeadlineExceeded", err)
197-	}
198-	c, _ := svc.Get(ctx, id)
199-	if c.Status != StatusDraft || c.Contents[0].SMS != "hi" {
200-		t.Fatalf("campaign changed: %+v", c)
201-	}
202-}
203-
204-func TestService_SampleNames(t *testing.T) {
205-	ctx, _, svc := seededService(t)
ba
… [76 more chars truncated]
```

</details>

**🔧 Bash** _Read review dialog_
```bash
cat -n frontend/src/components/campaign/review-dialog.vue
```

**🔧 Bash** _Read frontend lib and diffs_
```bash
cat -n frontend/src/lib/campaigns.ts frontend/src/lib/limits.ts frontend/src/components/campaign/message-preview.vue; git diff frontend/src/lib/content.ts frontend/src/components/campaign/content-cards.vue
```

<details><summary>Result</summary>

```
1	<script setup lang="ts">
     2	import { computed, reactive, ref, watch } from 'vue'
     3	import { toast } from 'vue-sonner'
     4	
     5	import type { Campaign, ContentInput } from '@/lib/campaigns'
     6	import type { Content } from '@/lib/content'
     7	
     8	import { Alert, AlertDescription } from '@/components/ui/alert'
     9	import {
    10	  AlertDialog,
    11	  AlertDialogContent,
    12	  AlertDialogDescription,
    13	  AlertDialogFooter,
    14	  AlertDialogHeader,
    15	  AlertDialogTitle,
    16	} from '@/components/ui/alert-dialog'
    17	import { Button } from '@/components/ui/button'
    18	import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
    19	import { Input } from '@/components/ui/input'
    20	import { Label } from '@/components/ui/label'
    21	import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
    22	import { Spinner } from '@/components/ui/spinner'
    23	import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
    24	import { Textarea } from '@/components/ui/textarea'
    25	import { approveCampaign, CampaignRequestError, getCampaign, getSampleNames, languageName, saveContents } from '@/lib/campaigns'
    26	import { checkContent, LIMITS, renderedLength } from '@/lib/limits'
    27	
    28	import MessagePreview from './message-preview.vue'
    29	
    30	type Field = keyof Content
    31	type Channel = 'sms' | 'email'
    32	
    33	const props = defineProps<{ campaignId: number }>()
    34	const emit = defineEmits<{ approved: [id: number] }>()
    35	const open = defineModel<boolean>('open', { default: false })
    36	
    37	const FALLBACK_NAME = 'Alex'
    38	const VAR_LABEL = '{{first_name}}'
    39	const FIELDS: Field[] = ['sms', 'email_subject', 'email_body']
    40	
    41	const campaign = ref<Campaign | null>(null)
    42	const loadError = ref('')
    43	const edits = reactive<Record<string, Content>>(
… [19169 more chars truncated]
```

</details>

<details><summary>Result</summary>

```
1	import type { Content } from '@/lib/content'
     2	
     3	import { apiFetch } from '@/lib/api-client'
     4	
     5	export type CampaignStatus = 'draft' | 'approved'
     6	export interface LanguageContent extends Content { language: string, sms_count: number, email_count: number, first_name_runes: number }
     7	export interface Campaign {
     8	  id: number
     9	  status: CampaignStatus
    10	  brief: string
    11	  description: string
    12	  skipped_languages: string[]
    13	  created_at: string
    14	  updated_at: string
    15	  approved_at: string | null
    16	  contents: LanguageContent[]
    17	}
    18	export interface CampaignSummary { id: number, status: CampaignStatus, description: string, languages: string[], created_at: string, approved_at: string | null }
    19	export interface FieldError { path: string, message: string }
    20	export type ContentInput = Content & { language: string }
    21	
    22	export const LANGUAGE_NAMES: Record<string, string> = { tr: 'Turkish', en: 'English' }
    23	
    24	export function languageName(code: string): string {
    25	  return LANGUAGE_NAMES[code] ?? code.toUpperCase()
    26	}
    27	
    28	const FALLBACK_MESSAGE = 'Something went wrong. Try again.'
    29	
    30	export class CampaignRequestError extends Error {
    31	  code: string
    32	  details: FieldError[]
    33	  status: number
    34	
    35	  constructor(code: string, message: string, status: number, details: FieldError[]) {
    36	    super(message)
    37	    this.name = 'CampaignRequestError'
    38	    this.code = code
    39	    this.status = status
    40	    this.details = details
    41	  }
    42	}
    43	
    44	function toError(err: unknown): CampaignRequestError {
    45	  const e = err as { data?: { error?: { code?: unknown, message?: unknown, details?: unknown } }, statusCode?: unknown } | null
    46	  const body = e?.data?.error
    47	  const code = typeof body?.code === 'string' ? body.code : 'request_failed'
 
… [10721 more chars truncated]
```

</details>

**🔧 Bash** _Read campaign pages_
```bash
sed -n 1,80p frontend/src/pages/campaigns/index.vue; sed -n 1,70p "frontend/src/pages/campaigns/[id].vue"
```

<details><summary>Result</summary>

```
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import type { CampaignSummary } from '@/lib/campaigns'

import ReviewDialog from '@/components/campaign/review-dialog.vue'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { CampaignRequestError, languageName, listCampaigns } from '@/lib/campaigns'

const router = useRouter()
const campaigns = ref<CampaignSummary[] | null>(null)
const error = ref('')
const reviewId = ref<number | null>(null)
const reviewOpen = ref(false)

async function load() {
  error.value = ''
  try {
    campaigns.value = await listCampaigns()
  }
  catch (err) {
    error.value = err instanceof CampaignRequestError ? err.message : 'Something went wrong. Try again.'
  }
}

onMounted(load)

function languages(c: CampaignSummary) {
  return c.languages.map(languageName).join(', ')
}

function when(iso: string) {
  return new Date(iso).toLocaleString()
}

function openCampaign(c: CampaignSummary) {
  if (c.status === 'draft') {
    reviewId.value = c.id
    reviewOpen.value = true
  }
  else {
    router.push(`/campaigns/${c.id}`)
  }
}

function onReviewOpen(v: boolean) {
  reviewOpen.value = v
  if (!v)
    void load()
}

function onApproved(id: number) {
  router.push(`/campaigns/${id}`)
}
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-2">
      <h1 class="text-2xl font-bold tracking-tight">
        Campaigns
      </h1>
      <p class="text-muted-foreground">
        Drafts wait for your review. Approved campaigns are ready to send.
      </p>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <div v-else-if="campaigns 
… [2246 more chars truncated]
```

</details>

**🔧 Bash** _Dialog usage and regenerate tests_
```bash
grep -n "ReviewDialog\|review-dialog" -A6 frontend/src/pages/campaigns/index.vue "frontend/src/pages/campaigns/[id].vue" | sed -n 1,60p; grep -n "regenerat\|campaign_conflict\|draft_not_saved\|stays off" frontend/src/components/campaign/__tests__/content-cards.test.ts | head -20
```

<details><summary>Result</summary>

```
frontend/src/pages/campaigns/[id].vue:7:import ReviewDialog from '@/components/campaign/review-dialog.vue'
frontend/src/pages/campaigns/[id].vue-8-import { Alert, AlertDescription } from '@/components/ui/alert'
frontend/src/pages/campaigns/[id].vue-9-import { Badge } from '@/components/ui/badge'
frontend/src/pages/campaigns/[id].vue-10-import { Button } from '@/components/ui/button'
frontend/src/pages/campaigns/[id].vue-11-import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
frontend/src/pages/campaigns/[id].vue-12-import { CampaignRequestError, getCampaign, languageName } from '@/lib/campaigns'
frontend/src/pages/campaigns/[id].vue-13-
--
frontend/src/pages/campaigns/[id].vue:129:      <ReviewDialog
frontend/src/pages/campaigns/[id].vue-130-        v-if="campaign.status === 'draft'"
frontend/src/pages/campaigns/[id].vue-131-        :open="reviewOpen"
frontend/src/pages/campaigns/[id].vue-132-        :campaign-id="campaign.id"
frontend/src/pages/campaigns/[id].vue-133-        @update:open="onReviewOpen"
frontend/src/pages/campaigns/[id].vue-134-        @approved="load"
frontend/src/pages/campaigns/[id].vue-135-      />
frontend/src/pages/campaigns/index.vue:7:import ReviewDialog from '@/components/campaign/review-dialog.vue'
frontend/src/pages/campaigns/index.vue-8-import { Alert, AlertDescription } from '@/components/ui/alert'
frontend/src/pages/campaigns/index.vue-9-import { Badge } from '@/components/ui/badge'
frontend/src/pages/campaigns/index.vue-10-import { Button } from '@/components/ui/button'
frontend/src/pages/campaigns/index.vue-11-import { Card, CardContent } from '@/components/ui/card'
frontend/src/pages/campaigns/index.vue-12-import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
frontend/src/pages/campaigns/index.vue-13-import { CampaignRequestError, languageName, listCampaigns } from '@/lib/campaigns'
--
frontend/src/pages/campaigns/index.vue:159:    <ReviewDialog
frontend/src/
… [692 more chars truncated]
```

</details>

**🔧 Bash** _Check regenerate and Escape tests_
```bash
sed -n 240,310p frontend/src/components/campaign/__tests__/content-cards.test.ts; grep -n "Escape\|outside" frontend/src/components/campaign/__tests__/review-dialog.test.ts | head
```

<details><summary>Result</summary>

```
expect(btn().attributes('disabled')).toBeUndefined()
    await btn().trigger('click')
    expect(w.find('[data-test="review-stub"]').attributes('data-id')).toBe('7')
    await w.find('[data-test="stub-approve"]').trigger('click')
    expect(push).toHaveBeenCalledWith('/campaigns/7')
  })

  it('no review button without a campaign event', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'error', data: { language: 'tr', code: 'llm_error', message: 'x' } })
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(w.find('[data-test="review-button"]').exists()).toBe(false)
  })

  it('regenerate sends the campaign id and a new run does not', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'campaign', data: { id: 7, status: 'draft' } })
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(gen.mock.calls[0]![4]).toBeUndefined()
    open()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(gen.mock.calls[1]![4]).toBe(7)
    open()
    await w.setProps({ runId: 2 })
    await flushPromises()
    expect(gen.mock.calls[2]![4]).toBeUndefined()
  })

  it('shows an error with no language in the content area', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'final', data: { language: 'tr', content, limits } })
    r.get()({ event: 'error', data: { code: 'draft_not_saved', message: 'The draft could not be saved. Try again.' } })
    await flushPromises()
    expect(w.find('[role="alert"], .alert, [data-slot="alert"]').text()).toContain('The draft could not be saved.')
    expect(w.find('[data-language="tr"]').exists()).toBe(true)
  })
})
206:  it('asks before closing with Escape', async () => {
209:    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
```

</details>

**🔧 Bash** _Find ContentCards usage_
```bash
grep -rn "ContentCards\|content-cards" frontend/src --include=*.vue | grep -v __tests__; grep -rn "<ContentCards" -B2 -A6 frontend/src --include=*.vue
```

<details><summary>Result</summary>

```
(eval):1: no matches found: --include=*.vue
(eval):1: no matches found: --include=*.vue
```

</details>

**🔧 Bash** _Find ContentCards usage_
```bash
grep -rn "ContentCards" frontend/src | grep -v __tests__ | grep -v d.ts
```

<details><summary>Result</summary>

```
frontend/src/components/campaign/brief-audience.vue:12:import ContentCards from './content-cards.vue'
frontend/src/components/campaign/brief-audience.vue:166:    <ContentCards
```

</details>

**🔧 Bash** _Read brief-audience props passing_
```bash
sed -n 15,90p frontend/src/components/campaign/brief-audience.vue; sed -n 160,175p frontend/src/components/campaign/brief-audience.vue
```

<details><summary>Result</summary>

```
const brief = ref('')
const length = computed(() => brief.value.length)
const canSubmit = computed(() => brief.value.trim() !== '' && length.value <= MAX_BRIEF)

const mutation = useMutation({ mutationFn: previewSegment, meta: { silent: true } })
const pending = computed(() => mutation.isPending.value)
const result = computed(() => (mutation.isError.value ? undefined : mutation.data.value))
const errorMessage = computed(() => (mutation.isError.value ? apiErrorMessage(mutation.error.value) : ''))

const runId = ref(0)
const running = ref(false)
const contentBrief = ref('')
const contentRule = ref<unknown>(null)

function submit() {
  if (!canSubmit.value || pending.value || running.value)
    return
  const sent = brief.value
  // Unmounts the cards, which aborts any open content request.
  runId.value = 0
  running.value = false
  mutation.mutate(sent, {
    onSuccess: (data) => {
      if (data.audience.languages.length === 0)
        return
      contentBrief.value = sent
      contentRule.value = data.rule
      runId.value++
    },
  })
}
</script>

<template>
  <div class="space-y-6">
    <div class="space-y-2">
      <Textarea
        v-model="brief"
        class="min-h-32"
        placeholder="Example: Send a winter sale to customers in Turkey who bought in the last 90 days."
        aria-label="Campaign brief"
      />
      <div class="flex items-center justify-between">
        <span class="text-sm" :class="length > MAX_BRIEF ? 'text-destructive' : 'text-muted-foreground'">
          {{ length }} / {{ MAX_BRIEF }}
        </span>
        <Button :disabled="!canSubmit || pending || running" @click="submit">
          Find audience
        </Button>
      </div>
    </div>

    <p v-if="pending" class="text-sm text-muted-foreground" role="status">
      Generating segment…
    </p>

    <Alert v-if="errorMessage" variant="destructive">
      <AlertDescription>{{ errorMessage }}</AlertDescription>
    </Alert>

    <Card v-if="result && !pending">
      <Car
… [732 more chars truncated]
```

</details>

**🔧 Bash** _Read list template_
```bash
sed -n 80,158p frontend/src/pages/campaigns/index.vue
```

<details><summary>Result</summary>

```
</p>
      <Button as-child variant="outline">
        <RouterLink to="/campaigns/new">
          Create a campaign
        </RouterLink>
      </Button>
    </div>

    <template v-else-if="campaigns">
      <div class="hidden rounded-lg border md:block">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-16">
                #
              </TableHead>
              <TableHead>Audience</TableHead>
              <TableHead>Languages</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="c in campaigns"
              :key="c.id"
              data-test="campaign-row"
              class="cursor-pointer"
              tabindex="0"
              @click="openCampaign(c)"
              @keydown.enter="openCampaign(c)"
            >
              <TableCell class="font-medium tabular-nums">
                #{{ c.id }}
              </TableCell>
              <TableCell class="max-w-md truncate">
                {{ c.description }}
              </TableCell>
              <TableCell>{{ languages(c) }}</TableCell>
              <TableCell>
                <Badge :variant="c.status === 'draft' ? 'outline' : 'default'">
                  {{ c.status === 'draft' ? 'Draft' : 'Approved' }}
                </Badge>
              </TableCell>
              <TableCell class="text-muted-foreground">
                {{ when(c.created_at) }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <div class="space-y-3 md:hidden">
        <Card
          v-for="c in campaigns"
          :key="c.id"
          class="cursor-pointer py-4"
          tabindex="0"
          @click="openCampaign(c)"
          @keydown.enter="openCampaign(c)"
        >
          <CardContent class="space-y-2 px-4">
            <div class="flex items-ce
… [582 more chars truncated]
```

</details>

**🔧 Bash** _Run STE check on changed docs_
```bash
python3 .claude/skills/ste-writing/scripts/ste_check.py docs/architecture.md docs/decisions-log.md docs/specs/06-review-approve/*.md docs/plans/06-review-approve.md 2>&1 | tail -20; git diff --stat docs
```

<details><summary>Result</summary>

```
docs/specs/06-review-approve/goal.md:137: present-perfect: has unsaved
docs/specs/06-review-approve/requirements.md:16: sentence-length: 26 words: **FR-8.** A request with the CODE of an approved campaign, or with a different brief or rule, gives an CODE event with no CODE and CODE.
docs/specs/06-review-approve/requirements.md:25: sentence-length: 27 words: **FR-14.** These bodies return CODE for save and approve: a body over 64 KiB, a body that is not valid JSON, and a body with no CODE.
docs/plans/06-review-approve.md:42: sentence-length: 29 words: **The marketer clicks "Regenerate content" after a slow first run, and the old stream saves a draft late.** The old stream must not save after the page aborted it.
docs/plans/06-review-approve.md:44: sentence-length: 31 words: **A long Turkish name, for example CODE, joins the audience between the draft and the approval.** The approval must use the new CODE, so an SMS that fit before can fail.
docs/plans/06-review-approve.md:76: sentence-length: 27 words: Then each of these fails: an update of CODE, a delete of the content row, an insert of an CODE row, an update of CODE, and CODE.
docs/plans/06-review-approve.md:306: sentence-length: 28 words: [ ] **Step 5: Postgres check** CODE (Postgres, seed users, real CODE as the saver, default mock): after the stream, CODE returns a draft with a CODE row.
docs/plans/06-review-approve.md:351: sentence-length: 26 words: one test for each row of the spec error table and for each FR-14 body case (CODE, CODE, CODE, over 64 KiB, not JSON, no CODE)
docs/plans/06-review-approve.md:405: sentence-length: 29 words: CODE: the FR-9 cases of slice 05 (CODE 20, SMS 154 runes + one variable = 160 passes, 155 fails, subject 80 passes and 81 fails, CODE counts 1).
docs/plans/06-review-approve.md:492: sentence-length: 34 words: [ ] **Step 1: CODE.** Replace "TBD per slice" in Flow step 3 with the draft save, the review window, the save and approve endpoints, the row lock, the triggers and the sampl
… [610 more chars truncated]
```

</details>

**🔧 Bash** _STE check for architecture and log_
```bash
python3 .claude/skills/ste-writing/scripts/ste_check.py docs/architecture.md docs/decisions-log.md 2>&1 | head; git diff docs/decisions-log.md | grep -n "^+" | grep -i "row lock\|count\|first_name_runes" | head
```

<details><summary>Result</summary>

```
27:+- **Options:** Trust the browser counters. Check with the `first_name_runes` of the draft. Run the audience query again.
28:+- **Choice:** Each save and each approval parses the stored rule, runs the audience query again and checks the content with the new longest name. The counts in the campaign update. The browser sends no rule after the draft exists.
31:+## D-039 A row lock stops a double approval
34:+- **Options:** An idempotency key from the browser. A row lock and a status check.
45:+## D-041 The campaign stores counts, not recipients
48:+- **Options:** Store the recipient list at approval. Store only the counts.
49:+- **Choice:** Each content row stores the SMS count, the email count and `first_name_runes` of the last save. The send step gets the recipients with `segment.Recipients` at send time.
```

</details>

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/.claude/agent-memory/reviewer/review-approve-defects.md`
```
---
name: review-approve-defects
description: Defects found in the slice 06 review (draft cache, async load race, count under row lock, stale first_name_runes after 422)
metadata:
  type: project
---

Patterns found in the slice 06 review on 2026-11-10. Check for them in later slices (07 send, 08 dashboard).

- Component cache keyed by id only ("reload only when id changes"). A regenerate keeps the same id but changes the server data, so the dialog shows stale content. Check every "loadedFor === id" style guard.
- Async load with no stale-response guard. Two opens in a row: the older response can overwrite the newer one. Check that code compares the id after `await`.
- A second pool query (audience count) inside a transaction that holds `SELECT ... FOR UPDATE`. With N >= pool size concurrent requests this starves the pool. Waiters also hit the 2 s statement_timeout and get 503, not the idempotent 200.
- After a server 422 because the audience changed, the browser keeps the old counts and `first_name_runes`. The counter says OK, the server says no.
- Trigger reads the parent status without `FOR SHARE`. Cascade delete of an approved parent passes the child trigger.

**Why:** the tests pinned the happy concurrency path but not these.
**How to apply:** grep for row locks plus external calls, and for id-only caches in Vue components. Related: [[http-config-validation-defects]].
```

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/.claude/agent-memory/reviewer/review-approve-defects.md (file state is current in your context — no need to Read it back)
```

</details>

**🔧 Bash** _Add memory index entry_
```bash
cd ~/Desktop/projects/case-studies/insider_one/.claude/agent-memory/reviewer && echo "- [Review and approve defects](review-approve-defects.md) — id-only dialog cache, stale async load, count under row lock, stale limits after 422" >> MEMORY.md
```

<details><summary>Result</summary>

```
(Bash completed with no output)
```

</details>

**🔧 SubagentHandback** ```json
{
  "message": "# Slice 06 review (review and approve)\n\nI found 1 critical defect, 4 important defects and 5 minor defects. Three of the five Review Focus items (2, 3 and 5) work as the plan says. Item 1 has a small timing gap (Minor 1). Item 4 can give a 503 instead of the 200 (Important 1). The hard rules hold:\n- The backend parses and compiles the stored rule again on each save and approval.\n- No LLM text reaches SQL.\n- The sample query reads first names only.\n- The logs contain no brief and no content.\n\n## Critical\n\n1. `frontend/src/components/campaign/review-dialog.vue:102-105`, together with `content-cards.vue:99-104`. \"Regenerate content\" keeps the same `campaignId`, and `ReviewDialog` stays mounted. The watch loads only when `loadedFor !== id || !campaign`.\n   - Failure: the marketer opens \"Review & approve\", closes it and clicks \"Regenerate content\". The backend replaces the draft. The marketer opens \"Review & approve\" again. The window shows the OLD text, counts and `first_name_runes`. \"Send now\" sends the old text, and the backend approves it. This overwrites the new content. The marketer sends the copy that they rejected.\n   - Fix: load again on each open when the window has no unsaved edits. Or give the dialog `:key=\"campaignId + '-' + runCount\"` in content-cards.\n\n## Important\n\n1. `backend/internal/campaign/store.go:200-211` and `service.go:86-88`. `Update` holds the row lock (`FOR UPDATE`) on connection A. Then `check` runs `CountAudience` on a second pool connection.\n   - Failure 1: the pool has `max(4, NumCPU)` connections. With that many concurrent saves or approvals, the waiters hold all the connections. The lock holder cannot get a connection for `Count`, so the requests wait until the 2 s timeouts and fail with 503.\n   - Failure 2: a large audience makes `Count` take about 1.5 s. The second tab waits on `FOR UPDATE` under `statement_timeout = 2s` and gets 503. Focus item 4 expects 200 with the campaign unchanged.\n   - The 10-way test passes only because the test machine has many CPUs.\n   - Fix: read the campaign and run `Count` before you open the transaction. Then lock the row, check the status, validate and write. Or use `SET LOCAL lock_timeout` larger than `statement_timeout`.\n2. `review-dialog.vue:81-99`. `load()` does not check that the response belongs to the current id.\n   - Failure: on `/campaigns`, the marketer opens draft A on a slow network, closes it, then opens draft B. The response for A arrives last. The window shows A's text under B's id. \"Approve & send\" approves B with A's copy when the languages match.\n   - Fix: after each `await`, return if `loadedFor !== props.campaignId`.\n3. `review-dialog.vue:173-190` and `:60`. After a 422 because a longer name joined the audience (FR-17), the window keeps the old `first_name_runes`.\n   - Failure: the counter shows a pass, for example `150 / 160`, and the field shows the server error. One keystroke clears the server error (`onInpu
… [3263 more chars truncated]
```

<details><summary>Result</summary>

```
{"success":true,"message":"Report delivered to your caller."}
```

</details>

