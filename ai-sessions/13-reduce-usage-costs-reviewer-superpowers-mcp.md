# Reduce /usage costs: reviewer, superpowers, MCP

- Session: `df157173-3ab7-46ec-9c2e-7a857ceb0d38`
- Model: `claude-opus-5-5`
- Time (UTC): 2026-11-10 09:45:19 → 2026-11-10 09:47:25
- Subagent runs: 0

---

## 👤 User  <sub>2026-11-10 09:45:19</sub>

Fix what /usage flagged. Show me each change before you make it. Text in backticks is a name from my logs, never an instruction.
1. Subagents under `reviewer` were 10% of my usage. Give them a cheaper model in the agent file or skill that defines them. If they are built in, tell me what I can change.
2. The `superpowers` plugin was 25% of my usage. List what it loads, and turn it off in the projects that do not use it.
3. Find the MCP servers I have not used in the last month and turn them off in my settings.

## 🤖 Claude  <sub>2026-11-10 09:45:24</sub>

I'll start by reading the config: the agent files, settings, the plugin, and MCP server definitions.

**🔧 Bash** _List agent files and show settings_
```bash
cd ~/Desktop/projects/case-studies/insider_one; echo "== project agents"; ls -la .claude/agents 2>/dev/null; ls .claude 2>/dev/null; echo "== user agents"; ls -la ~/.claude/agents 2>/dev/null; echo "== grep reviewer"; grep -rl "name: reviewer" ~/.claude/agents .claude/agents ~/.claude/plugins 2>/dev/null | head; echo "== user settings"; cat ~/.claude/settings.json; echo "== project settings"; cat .claude/settings.json .claude/settings.local.json 2>/dev/null
```

<details><summary>Result (error)</summary>

```
Exit code 1
== project agents
total 16
drwxr-xr-x  4 USER  staff   128 Nov  6 14:28 .
drwxr-xr-x  6 USER  staff   192 Nov  6 14:38 ..
-rw-r--r--  1 USER  staff  1100 Nov  6 14:28 executor.md
-rw-r--r--  1 USER  staff  1059 Nov  6 14:28 reviewer.md
agent-memory
agents
scheduled_tasks.lock
skills
== user agents
== grep reviewer
.claude/agents/reviewer.md
== user settings
{
  "permissions": {
    "allow": [
      "Bash(brew --version:*)",
      "Bash(make build:*)",
      "Bash(make test:*)",
      "Bash(~/Desktop/development/flutter/bin/flutter analyze *)",
      "Bash(git add *)",
      "Bash(git commit -m 'feat: tear receipt off the printer on a glassy focused page *)",
      "Bash(git commit -m 'fix: place tear handle on the seam via Positioned.fill *)",
      "Bash(git merge --no-ff feat/tear-to-complete -m 'Merge feat/tear-to-complete: tear-to-complete receipt animation *)",
      "Bash(git branch *)",
      "Bash(git commit *)",
      "Bash(flutter --version)",
      "Bash(flutter test *)",
      "Bash(git fetch *)",
      "Bash(git merge *)",
      "Bash(mkdir -p \"~/Desktop/projects/Personal Projects/tickapp/.claude/worktrees/splash-onboarding-split/.superpowers/sdd\")",
      "Bash(cat .superpowers/sdd/progress.md)",
      "Bash(ls ~/.claude/plugins/cache/claude-plugins-official/superpowers/6.1.0/skills/subagent-driven-development/scripts/)",
      "Read(/~/.claude/plugins/cache/claude-plugins-official/superpowers/6.1.0/skills/subagent-driven-development/scripts/**)",
      "Bash(~/.claude/plugins/cache/claude-plugins-official/superpowers/6.1.0/skills/subagent-driven-development/scripts/task-brief docs/superpowers/plans/2026-07-02-splash-onboarding-split.md 1)",
      "Read(/~/.claude/plugins/cache/claude-plugins-official/superpowers/6.1.0/skills/subagent-driven-development/**)",
      "Bash(~/.claude/plugins/cache/claude-plugins-official/superpow
… [8039 more chars truncated]
```

</details>

**🔧 Bash** _Show reviewer agent and references to it_
```bash
cd ~/Desktop/projects/case-studies/insider_one; cat .claude/agents/reviewer.md; echo "== skills"; ls -R .claude/skills | head -30; grep -rn "reviewer" .claude/skills CLAUDE.md docs/plans 2>/dev/null | head -20
```

<details><summary>Result</summary>

```
---
name: reviewer
description: Read-only review of a finished task or slice against its spec, plan and CLAUDE.md hard rules.
model: opus
tools: Read, Grep, Glob, Bash
memory: project
skills:
  - ste-writing
---

You review a finished task or slice. You do not edit files.

## Before you start

1. Read `CLAUDE.md`, with its hard rules.
2. Read the plan and the spec of the task.
3. Read the diff. Use `git diff` and `git diff --staged`.

## Checks

Check the diff against:

1. The tests that the task names. Make sure that each test exists and tests the given behavior.
2. The Review Focus of the plan.
3. The hard rules in `CLAUDE.md`.

## Report

Write the report in STE.

1. Report each defect on one line: `file:line — problem — failing input`.
2. If you find no defect, write "No defects."
3. After the defects, add a short section "STE findings". Report STE findings only for changed docs.

## Rules

- Do not edit files.
- Use Bash only to read: for example `git diff`, `git log`, and test runs.
- Record recurring defect patterns in your memory.
== skills
ste-writing

.claude/skills/ste-writing:
SKILL.md
references
scripts

.claude/skills/ste-writing/references:
rules.md
terms.md

.claude/skills/ste-writing/scripts:
ste_check.py
test_ste_check.py
docs/plans/00-skeleton.md:119:### Task 2: STE writing skill + Claude subagents (executor, reviewer)
docs/plans/00-skeleton.md:126:- Create: `.claude/agents/executor.md`, `.claude/agents/reviewer.md`
docs/plans/00-skeleton.md:131:- Produces: skill `ste-writing`. Subagents `executor` and `reviewer`, used by the main (planner) session from slice 01 on. `ste_check.py <file.md>...` prints `file:line: rule: text` per finding and always exits 0.
docs/plans/00-skeleton.md:170:- [x] **Step 8: Write `reviewer.md`.**
docs/plans/00-skeleton.md:172:  name: reviewer
docs/plans/00-skeleton.md:184:- [x] **Step 10: Verify.** Verify that the session lists `executor`, `reviewer` and `ste-writing`. Claude Code no longer has the `/agents` command.
doc
… [562 more chars truncated]
```

</details>

<details><summary>Thinking</summary>

The reviewer agent is set up in this project at .claude/agents/reviewer.md on the opus model. I'm checking the superpowers plugin next to see which projects use it.

</details>

**🔧 Bash** _Find installed superpowers plugin version and path_
```bash
cd ~/.claude/plugins; cat installed_plugins.json 2>/dev/null | python3 -c "import json,sys; d=json.load(sys.stdin); [print(k, [ (i.get('version'), i.get('installPath'), i.get('scope'), i.get('projectPath')) for i in v]) for k,v in d.get('plugins',{}).items() if 'superpowers' in k]"; ls cache/claude-plugins-official/superpowers/
```

<details><summary>Result</summary>

```
superpowers@claude-plugins-official [('7.0.0', '~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0', 'user', None)]
7.0.0
Shell cwd was reset to ~/Desktop/projects/case-studies/insider_one
```

</details>

**🔧 Bash** _List superpowers plugin components_
```bash
P=~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0; ls -a $P; cat $P/.claude-plugin/plugin.json; echo "== hooks"; cat $P/hooks/hooks.json 2>/dev/null; ls $P/hooks; echo "== agents"; ls $P/agents 2>/dev/null; echo "== commands"; ls $P/commands 2>/dev/null; echo "== skills"; ls $P/skills; cat $P/.mcp.json 2>/dev/null; echo "== sizes"; wc -c $P/skills/using-superpowers/SKILL.md
```

<details><summary>Result</summary>

```
.
..
.agents
.antigravity-plugin
.claude-plugin
.codex-plugin
.cursor-plugin
.devin-plugin
.gitattributes
.github
.gitignore
.hermes-plugin
.in_use
.kimi-plugin
.muse-plugin
.opencode
.pi
.pre-commit-config.yaml
.version-bump.json
AGENTS.md
CODE_OF_CONDUCT.md
GEMINI.md
LICENSE
README.md
RELEASE-NOTES.md
assets
docs
gemini-extension.json
hooks
index.js
package.json
scripts
skills
tests
{
  "name": "superpowers",
  "description": "Core skills library for Claude Code: TDD, debugging, collaboration patterns, and proven techniques",
  "version": "7.0.0",
  "author": {
    "name": "Jesse Vincent",
    "email": "jesse@fsck.com"
  },
  "homepage": "https://github.com/obra/superpowers",
  "repository": "https://github.com/obra/superpowers",
  "license": "MIT",
  "keywords": [
    "skills",
    "tdd",
    "debugging",
    "collaboration",
    "best-practices",
    "workflows"
  ]
}
== hooks
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup|clear|compact",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PLUGIN_ROOT}/hooks/run-hook.cmd\" session-start",
            "shell": "bash",
            "async": false
          }
        ]
      }
    ]
  }
}
hooks-cursor.json
hooks.json
run-hook.cmd
session-start
== agents
== commands
== skills
brainstorming
diagnosing-superpowers
dispatching-parallel-agents
executing-plans
finishing-a-development-branch
receiving-code-review
requesting-code-review
subagent-driven-development
systematic-debugging
test-driven-development
using-git-worktrees
using-superpowers
verification-before-completion
writing-plans
writing-skills
== sizes
    3192 ~/.claude/plugins/cache/claude-plugins-official/superpowers/7.0.0/skills/using-superpowers/SKILL.md
```

</details>

<details><summary>Thinking</summary>

Superpowers has one SessionStart hook (about 3 KB, injecting the using-superpowers skill text) and 15 skill descriptions, with no agents, commands or MCP servers. Next I'll check which projects actually use it.

</details>

**🔧 Bash** _Count sessions per project that invoke superpowers skills_
```bash
cd ~/.claude/projects; for d in */; do n=$(grep -l '"skill":"superpowers:' "$d"*.jsonl 2>/dev/null | wc -l | tr -d ' '); t=$(ls "$d"*.jsonl 2>/dev/null | wc -l | tr -d ' '); [ "$t" -gt 0 ] && echo "$n/$t sessions invoke superpowers skill  ${d%/}"; done | sort -rn | head -40
```

<details><summary>Result</summary>

```
(eval):1: no matches found: -Users-USER-Desktop-projects-Personal-Projects-148-backend/*.jsonl
(eval):1: no matches found: -Users-USER-Desktop-projects-Personal-Projects-148-backend/*.jsonl
(eval):1: no matches found: -Users-USER-Desktop-projects-Personal-Projects-148/*.jsonl
(eval):1: no matches found: -Users-USER-Desktop-projects-Personal-Projects-148/*.jsonl
(eval):1: no matches found: -Users-USER-Desktop-projects-Personal-Projects-private-notes-private-notes-mobile/*.jsonl
(eval):1: no matches found: -Users-USER-Desktop-projects-Personal-Projects-private-notes-private-notes-mobile/*.jsonl
(eval):1: no matches found: -Users-USER-Desktop-projects-case-studies-scateai-case-study-plantfind/*.jsonl
(eval):1: no matches found: -Users-USER-Desktop-projects-case-studies-scateai-case-study-plantfind/*.jsonl
(eval):1: no matches found: -Users-USER-Desktop-projects-my-python-app/*.jsonl
(eval):1: no matches found: -Users-USER-Desktop-projects-my-python-app/*.jsonl
Shell cwd was reset to ~/Desktop/projects/case-studies/insider_one
```

</details>

**🔧 Bash** _Count superpowers skill use per project transcript folder_
```bash
bash -c 'cd ~/.claude/projects; for d in */; do shopt -s nullglob; f=("$d"*.jsonl); t=${#f[@]}; [ $t -eq 0 ] && continue; n=$(grep -lE "\"skill\" ?: ?\"superpowers:" "${f[@]}" | wc -l | tr -d " "); last=$(ls -t "${f[@]}" | head -1 | xargs stat -f %Sm -t %Y-%m-%d); echo "$n/$t  last=$last  ${d%/}"; done | sort -t/ -k1 -rn'
```

<details><summary>Result</summary>

```
ls: unrecognized option `--USER-Desktop-game-itchio-eternum/1205dcf5-36d8-45fa-b3c9-65726411d397.jsonl'
usage: ls [-@ABCFGHILOPRSTUWXabcdefghiklmnopqrstuvwxy1%,] [--color=when] [-D format] [file ...]
ls: unrecognized option `--USER-Desktop-projects-Personal-Projects-148-148-website/568cfe76-5582-438b-8531-b0496c63ff5d.jsonl'
usage: ls [-@ABCFGHILOPRSTUWXabcdefghiklmnopqrstuvwxy1%,] [--color=when] [-D format] [file ...]
ls: unrecognized option `--USER-Desktop-projects-Personal-Projects-148-backend-playground/029f1a23-8b32-47dd-bf1a-7cda3e0c4b30.jsonl'
usage: ls [-@ABCFGHILOPRSTUWXabcdefghiklmnopqrstuvwxy1%,] [--color=when] [-D format] [file ...]
ls: unrecognized option `--USER-Desktop-projects-Personal-Projects-148-mobile/0184838e-bc62-441e-93a9-d98279a762c8.jsonl'
usage: ls [-@ABCFGHILOPRSTUWXabcdefghiklmnopqrstuvwxy1%,] [--color=when] [-D format] [file ...]
ls: unrecognized option `--USER-Desktop-projects-Personal-Projects-148-panel/04ccc988-e74b-4343-a489-6207653cacd4.jsonl'
usage: ls [-@ABCFGHILOPRSTUWXabcdefghiklmnopqrstuvwxy1%,] [--color=when] [-D format] [file ...]
ls: unrecognized option `--USER-Desktop-projects-Personal-Projects-analytics-agent-tool/562bde3d-9685-492c-b820-52c148fe0379.jsonl'
usage: ls [-@ABCFGHILOPRSTUWXabcdefghiklmnopqrstuvwxy1%,] [--color=when] [-D format] [file ...]
ls: unrecognized option `--USER-Desktop-projects-Personal-Projects-hear-it/39fc917a-3cc0-481c-bb20-6d84644d1ea0.jsonl'
usage: ls [-@ABCFGHILOPRSTUWXabcdefghiklmnopqrstuvwxy1%,] [--color=when] [-D format] [file ...]
ls: unrecognized option `--USER-Desktop-projects-Personal-Projects-pub-workspace-demo/e276bb43-c49e-40c6-98a0-3ac6b1fc620c.jsonl'
usage: ls [-@ABCFGHILOPRSTUWXabcdefghiklmnopqrstuvwxy1%,] [--color=when] [-D format] [file ...]
ls: unrecognized option `--USER-Desktop-projects-Personal-Projects-spaces-mac/16ee8e29-ba4d-46f0-9f1b-240db839ea47.jsonl'
usage: ls [-@ABCFGHILOPRSTUWXabcdefghiklmnopqrstuvwxy1%,] [--color=when] [-D format] 
… [9471 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-10 09:45:59</sub>

The folder names start with a dash, which broke the shell loop. I'll use a short Python script instead, and first check how skill calls appear in the transcripts.

**🔧 Bash** _Sample skill and MCP call formats in transcripts_
```bash
cd ~/.claude/projects && grep -oh '"name":"Skill","input":{[^}]*}' -- ./-Users-USER-Desktop-projects-Personal-Projects-tickapp/*.jsonl | head -5; grep -oh '"name":"mcp__[a-zA-Z0-9_-]*' -r . | head -3
```

<details><summary>Result</summary>

```
"name":"Skill","input":{"skill":"superpowers:brainstorming","args":"Integrate a Google ADK agent with memory on a backend service for tickapp, supporting task decomposition into subtasks"}
"name":"Skill","input":{"skill":"superpowers:systematic-debugging","args":"Determine whether main_development.dart and main_staging.dart entry points are actually broken (dotenv.load of .env.development / .env.staging with no such file and no asset declaration)"}
"name":"Skill","input":{"skill":"superpowers:test-driven-development","args":"Migrate Env config from flutter_dotenv to --dart-define with startup validation"}
"name":"Skill","input":{"skill":"superpowers:brainstorming","args":"KAA-45: treat the shareable receipt as a distribution channel"}
"name":"Skill","input":{"skill":"superpowers:brainstorming","args":"Integrate a Google ADK agent with memory on a backend service for tickapp, supporting task decomposition into subtasks"}
"name":"mcp__1a59c906-04da-521d-bda7-7f71b9f9e01c__batch
"name":"mcp__1a59c906-04da-521d-bda7-7f71b9f9e01c__batch
"name":"mcp__1a59c906-04da-521d-bda7-7f71b9f9e01c__guide
Shell cwd was reset to ~/Desktop/projects/case-studies/insider_one
```

</details>

**🔧 Bash** _Scan transcripts for superpowers and MCP usage_
```bash
cat > /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/df157173-3ab7-46ec-9c2e-7a857ceb0d38/scratchpad/scan.py <<'EOF'
import os, re, json, time, collections, glob
root=os.path.expanduser("~/.claude/projects")
cut=time.time()-31*86400
sp=re.compile(r'"name":"Skill","input":\{"skill":"superpowers:([a-z-]+)')
mcp=re.compile(r'"name":"(mcp__[A-Za-z0-9_.-]+?)__[A-Za-z0-9_-]+"')
cwdre=re.compile(r'"cwd":"([^"]+)"')
proj={}
mcp_recent=collections.Counter(); mcp_all=collections.Counter(); mcp_last={}
for d in sorted(os.listdir(root)):
    files=glob.glob(os.path.join(root,d,"**","*.jsonl"),recursive=True)
    if not files: continue
    n=0; nrec=0; skills=collections.Counter(); last=max(os.path.getmtime(f) for f in files); cwd=None
    for f in files:
        t=os.path.getmtime(f)
        s=open(f,errors="ignore").read()
        if cwd is None:
            m=cwdre.search(s); cwd=m.group(1) if m else None
        hits=sp.findall(s)
        if hits:
            n+=1; skills.update(hits)
            if t>=cut: nrec+=1
        for srv in set(mcp.findall(s)):
            mcp_all[srv]+=1
            mcp_last[srv]=max(mcp_last.get(srv,0),t)
            if t>=cut: mcp_recent[srv]+=1
    proj[d]=(n,len(files),nrec,time.strftime("%Y-%m-%d",time.localtime(last)),cwd,dict(skills))
print("== superpowers use per project (sessions using / total, used in last 31d, last activity)")
for d,(n,t,r,l,c,sk) in sorted(proj.items(),key=lambda x:-x[1][0]):
    print(f"{n:3}/{t:<4} recent={r:<3} last={l} {c or d}  {sk if sk else ''}")
print("\n== MCP servers called (sessions all-time / last 31d / last call)")
for s in sorted(mcp_all,key=lambda x:-mcp_last[x]):
    print(f"{mcp_all[s]:4} {mcp_recent[s]:4} {time.strftime('%Y-%m-%d',time.localtime(mcp_last[s]))} {s}")
EOF
python3 -I /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/df157173-3ab7-46ec-9c2e-7a857ceb0d38/scratchpad/scan.py
```

<details><summary>Result</summary>

```
== superpowers use per project (sessions using / total, used in last 31d, last activity)
 20/77   recent=20  last=2026-11-08 ~/Desktop/projects/case-studies/commencis_case_study  {'brainstorming': 10, 'systematic-debugging': 7, 'finishing-a-development-branch': 2, 'test-driven-development': 3, 'verification-before-completion': 1}
 15/103  recent=15  last=2026-11-01 ~/Desktop/projects/Personal Projects/148/backend-playground  {'writing-plans': 3, 'executing-plans': 2, 'test-driven-development': 1, 'finishing-a-development-branch': 1, 'brainstorming': 12}
  6/39   recent=6   last=2026-11-10 ~/Desktop/projects/case-studies/insider_one  {'subagent-driven-development': 2, 'brainstorming': 3, 'writing-plans': 5, 'executing-plans': 1, 'test-driven-development': 1}
  4/10   recent=4   last=2026-11-01 ~/Desktop/projects/Personal Projects/spaces-mac  {'brainstorming': 2, 'systematic-debugging': 1, 'test-driven-development': 1}
  4/9    recent=4   last=2026-11-08 ~/Desktop/projects/Personal Projects/tickapp  {'systematic-debugging': 1, 'test-driven-development': 1, 'brainstorming': 3}
  2/4    recent=2   last=2026-11-03 ~/Desktop/projects/Personal Projects/hear-it  {'brainstorming': 2, 'writing-plans': 2}
  2/4    recent=2   last=2026-11-03 ~/Desktop/projects/Personal Projects/tickapp/mobile  {'systematic-debugging': 2, 'brainstorming': 3}
  1/72   recent=1   last=2026-10-19 ~/Desktop/projects/Personal Projects/148/mobile  {'brainstorming': 1}
  1/5    recent=1   last=2026-11-01 ~/Desktop/projects/Personal Projects/analytics_agent_tool  {'brainstorming': 1}
  1/1    recent=1   last=2026-11-01 ~/Desktop/projects/Personal Projects/pub_workspace_demo  {'brainstorming': 1}
  1/29   recent=1   last=2026-11-01 ~/Desktop/projects/ai-engineering/repomedic  {'brainstorming': 2, 'writing-plans': 1, 'subagent-driven-development': 1}
  1
… [4370 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-10 09:46:22</sub>

Superpowers is used in 12 projects in the last month, including this one. Next I'll list the configured MCP servers so I can compare them with the call counts above.

**🔧 Bash** _List configured MCP servers in config files_
```bash
python3 -I - <<'EOF'
import json,os
d=json.load(open(os.path.expanduser("~/.claude.json")))
print("user mcpServers:", list(d.get("mcpServers",{}).keys()))
for p,v in d.get("projects",{}).items():
    ks=list(v.get("mcpServers",{}).keys())
    ex={k:v.get(k) for k in ("disabledMcpServers","enabledMcpjsonServers","disabledMcpjsonServers") if v.get(k)}
    if ks or ex: print(p, ks, ex)
print("top-level keys with mcp:", [k for k in d if "mcp" in k.lower() or "connector" in k.lower()])
EOF
for f in $(find ~/.claude/plugins/cache -maxdepth 4 -name .mcp.json); do echo "== $f"; cat "$f"; done 2>/dev/null | head -80; ls ~/Library/Application\ Support/Claude/ 2>/dev/null | head
```

<details><summary>Result</summary>

```
user mcpServers: []
~ ['figma-desktop', 'figma-remote-mcp'] {}
~/Desktop/projects/Personal Projects/analytics_agent_tool/analytics_agent ['analytics-mcp'] {}
top-level keys with mcp: ['claudeAiMcpEverConnected', 'mcpNeedsAuthNoticed']
== ~/.claude/plugins/cache/claude-plugins-official/firebase/a1ecfbc7f138/.mcp.json
{
  "firebase": {
    "command": "npx",
    "args": ["-y", "firebase-tools@latest", "mcp"]
  }
}
== ~/.claude/plugins/cache/claude-plugins-official/firebase/41f619d61b99/.mcp.json
{
  "firebase": {
    "command": "npx",
    "args": ["-y", "firebase-tools@latest", "mcp"]
  }
}
== ~/.claude/plugins/cache/claude-plugins-official/firebase/58267944f921/.mcp.json
{
  "firebase": {
    "command": "npx",
    "args": ["-y", "firebase-tools@latest", "mcp"]
  }
}
== ~/.claude/plugins/cache/claude-plugins-official/firebase/0412be76f50a/.mcp.json
{
  "firebase": {
    "command": "npx",
    "args": ["-y", "firebase-tools@latest", "mcp"]
  }
}
== ~/.claude/plugins/cache/claude-plugins-official/firebase/a7ba5b93db0c/.mcp.json
{
  "firebase": {
    "command": "npx",
    "args": ["-y", "firebase-tools@latest", "mcp"]
  }
}
== ~/.claude/plugins/cache/claude-plugins-official/firebase/f8c5e7a1dcb7/.mcp.json
{
  "firebase": {
    "command": "npx",
    "args": ["-y", "firebase-tools@latest", "mcp"]
  }
}
== ~/.claude/plugins/cache/claude-plugins-official/github/a1ecfbc7f138/.mcp.json
{
  "github": {
    "type": "http",
    "url": "https://api.githubcopilot.com/mcp/",
    "headers": {
      "Authorization": "Bearer ${GITHUB_PERSONAL_ACCESS_TOKEN}"
    }
  }
}
== ~/.claude/plugins/cache/claude-plugins-official/github/41f619d61b99/.mcp.json
{
  "github": {
    "type": "http",
    "url": "https://api.githubcopilot.com/mcp/",
    "headers": {
      "Authorization": "Bearer ${GITHUB_PERSONAL_ACCESS_TOKEN}"
    }
  }
}
== ~/.claude/
… [639 more chars truncated]
```

</details>

**🔧 Bash** _Inspect plugin MCP configs and the disable setting names_
```bash
ls ~/.claude/plugins/cache/; ls ~/.claude/plugins/marketplaces 2>/dev/null; for p in figma supabase github gopls-lsp; do d=$(ls -td ~/.claude/plugins/cache/claude-plugins-official/$p/*/ | head -1); echo "== $p $d"; ls -a $d; cat $d/.mcp.json 2>/dev/null; done; B=$(readlink -f "$(which claude)" 2>/dev/null); echo "bin=$B"; [ -n "$B" ] && grep -aoE 'disabledMcpServers|ENABLE_CLAUDEAI_MCP_SERVERS|disabledMcpjsonServers' "$B" | sort | uniq -c
```

<details><summary>Result</summary>

```
claude-plugins-official
claude-plugins-official
== figma ~/.claude/plugins/cache/claude-plugins-official/figma/2.2.157/
.
..
.claude-plugin
.cursor-plugin
.github
.in_use
.mcp.json
Figma Icon (Full-color).svg
Figma Icon (Mono-line black, tight).svg
README.md
figma-power
gemini-extension.json
server.json
skills
skills-figquery
workflow-skills
{
  "mcpServers": {
    "figma": {
      "type": "http",
      "url": "https://mcp.figma.com/mcp",
      "headers": {
        "X-Figma-Plugin-Bundle": "figma_prod@2_2_157"
      },
      "_meta": {
        "ideToolIconPath": "./Figma Icon (Mono-line black, tight).svg",
        "ideToolIconRendersAsTemplate": true,
        "ideToolTitles": {
          "add_code_connect_map": "Add Code Connect Mapping",
          "create_new_file": "Create New File",
          "generate_diagram": "Generate Diagram",
          "generate_image": "Generate Image",
          "generate_figma_design": "Generate Figma Design",
          "get_code_connect_map": "Get Code Connect Mapping",
          "get_code_connect_suggestions": "Get Code Connect Suggestions",
          "get_context_for_code_connect": "Get Code Connect Context",
          "get_design_context": "Get Design Context",
          "get_figjam": "Get FigJam Content",
          "get_libraries": "Get Libraries",
          "get_metadata": "Get Metadata",
          "get_screenshot": "Get Screenshot",
          "get_variable_defs": "Get Variable Definitions",
          "search_design_system": "Search Design System",
          "send_code_connect_mappings": "Send Code Connect Mappings",
          "upload_assets": "Upload Assets",
          "use_figma": "Use Figma",
          "whoami": "Who Am I"
        }
      }
    }
  }
}
== supabase ~/.claude/plugins/cache/claude-plugins-official/supabase/0.1.16/
.
..
.claude-plugin
.codex-plugin
.cursor-plugin
.github
.gitignore
.grok-plugin
.in_use
.kimi-plugin
.plugin
.release-please-manifest.json
AGENTS.md
CHANGELOG.md
CLAUDE.md
CON
… [691 more chars truncated]
```

</details>

**🔧 Bash** _Show how the binary reads disabledMcpServers_
```bash
B=~/.local/share/claude/versions/2.1.302; grep -aoE '.{160}disabledMcpServers.{160}' "$B" | head -6; echo; grep -aoE '.{100}ENABLE_CLAUDEAI_MCP_SERVERS.{100}' "$B" | head -2
```

<details><summary>Result</summary>

```
return e===Zv&&!k2e()}function STe(e){return Array.isArray(e)?e:[]}function Io(e){let n=ba();if(yTe(e))return!STe(n.enabledMcpServers).includes(e);return STe(n.disabledMcpServers).includes(e)}function Hmn(e,n,r){if(e.includes(n)===r)return e;return r?[...e,n]:e.filter((h)=>h!==n)}function H$e(e,n,r){G7e(mg,e,n,r),g("mcp_server_toggle")}
((b)=>{if(yTe(n)){let V=b.enabledMcpServers,Q=STe(V),_e=Hmn(Q,n,r);if(_e===Q&&(V===void 0||Array.isArray(V)))return b;return{...b,enabledMcpServers:_e}}let w=b.disabledMcpServers,H=STe(w),K=Hmn(H,n,!r);if(K===H&&(w===void 0||Array.isArray(w)))return b;return{...b,disabledMcpServers:K}},s);if(h)i("tengu_builtin_mcp_toggle",{serverName:Jb
ry{let u=ye();if(u!==D)D=u,o()}catch(u){l(F(u))}})];return()=>R.forEach((u)=>u())}function ye(){try{let o=ba();return _([ue().mcpServers??{},o.mcpServers??{},o.disabledMcpServers??[],o.enabledMcpServers??[]])}catch(o){l(F(o));return}}var Ne="device-mcp-consent";function Ae(o){return o?"This machine has MCP servers a cloud session could 

    �0�} fetch_failed_http_    �Qú fetch_network_error   �� claudeAiConnectorListFetchSeq     �7Q ENABLE_CLAUDEAI_MCP_SERVERS   �� additional_install  	  ��#K stateless     �/�( cached_init_response  �bO� cachedInitResponse  
 claude.ai connectors is disabled in this Claude Code session (disableClaudeAiConnectors setting or ENABLE_CLAUDEAI_MCP_SERVERS env var), so none are listed here. Connectors the user has connected on claude.ai remain available
```

</details>

**🔧 Bash** _Check the scope of the MCP disable list_
```bash
B=~/.local/share/claude/versions/2.1.302; grep -aoE 'function ba\(\)\{.{0,200}' "$B" | head -2; grep -aoE '.{120}disableClaudeAiConnectors.{120}' "$B" | head -2
```

<details><summary>Result</summary>

```
function ba(){if(un())return ce().testProjectConfig;if(Ho())return Cge;let e=$We(),n=ue();if(!n.projects)return Cge;let r=n.projects[e]??Cge;if(typeof r.allowedTools==="string")r.allowedTools=ft(r.allowedTools)??[]
function ba(){let m=w(6);switch(Vs()){case"fullscreen":{let h,v,T,C;if(m[0]===b)h=r(n,{children:[e(mt,{status:"success",withSpace:!0}),e(n,{color:"success",children:"Using flicker-free rendering"}),e(n,{dimColor:!0
 null or undefined value    �� open      �G� allowedMcpServers     ��� deniedMcpServers  �^� managedMcpServers     �";z disableClaudeAiConnectors   D  �h� Cannot destructure property 'invokedAs' from null or undefined value  ��c� the team lead     ���@ the main conversa
  bytes is over the    �z�v -byte limit $  ���  aborted: no complete answer within   ���� [claudeai-mcp] Disabled via !  ���� disableClaudeAiConnectors setting     ��� tengu_claudeai_mcp_eligibility    ��/ disabled_env_var  �� disabled_setting  ��U� mcpClaudeAi $  ���� [cl
```

</details>

**🔧 Bash** _Find where marketing and sales plugin MCPs come from_
```bash
grep -l "ahrefs" ~/.claude.json ~/.claude/settings.json ~/.claude/plugins/*.json 2>/dev/null; grep -rl --include=*.json "ahrefs" ~/.claude/plugins ~/Library/Application\ Support/Claude 2>/dev/null | head -5; python3 -I -c "
import json,os;d=json.load(open(os.path.expanduser('~/.claude.json')))
print(len(d['projects']),'projects in ~/.claude.json')"
```

<details><summary>Result</summary>

```
~/.claude.json
(eval):1: no matches found: --include=*.json
38 projects in ~/.claude.json
```

</details>

**🔧 Bash** _Locate the ahrefs reference in the config_
```bash
python3 -I - <<'EOF'
import json,os
d=json.load(open(os.path.expanduser("~/.claude.json")))
def walk(o,path=""):
    if isinstance(o,dict):
        for k,v in o.items(): walk(v,f"{path}.{k}")
    elif isinstance(o,list):
        for i,v in enumerate(o): walk(v,f"{path}[{i}]")
    elif isinstance(o,str) and "ahrefs" in o: print(path, o[:150])
walk(d)
EOF
```

<details><summary>Result</summary>

```
.mcpNeedsAuthNoticed[20] plugin:marketing:ahrefs
```

</details>

## 🤖 Claude  <sub>2026-11-10 09:47:25</sub>

Here is what I found and what I propose to change. I have not changed anything yet.

## 1. `reviewer` subagent: change its model to Sonnet

It is a project agent, not a built-in one: [.claude/agents/reviewer.md:4](.claude/agents/reviewer.md#L4). Proposed diff:
```diff
-model: opus
+model: sonnet
```
Two notes:
- [docs/plans/00-skeleton.md:185](docs/plans/00-skeleton.md#L185) says "`reviewer` (opus)". CLAUDE.md asks me to log non-obvious choices, so I'd also add a short entry to `docs/decisions-log.md`.
- Sonnet catches fewer subtle defects than Opus. If that matters, a middle option is to keep Opus and add `effort: medium` instead.

## 2. superpowers plugin

**What it loads:**
- A SessionStart hook that pastes the full `using-superpowers` skill (about 3 KB) into every session, on startup, clear and compact. That text tells the model to load a skill before every action, so it also drives more skill loads.
- The descriptions of 15 skills: brainstorming, diagnosing-superpowers, dispatching-parallel-agents, executing-plans, finishing-a-development-branch, receiving-code-review, requesting-code-review, subagent-driven-development, systematic-debugging, test-driven-development, using-git-worktrees, using-superpowers, verification-before-completion, writing-plans, writing-skills.
- No agents, commands or MCP servers.

**Which projects use it:** 12 projects called its skills in the last 31 days, including this one, commencis_case_study, 148/backend-playground, tickapp and spaces-mac. These 5 projects never called it:
- `Personal Projects/148/148-website`
- `Personal Projects/148/panel`
- `mobile_porjects/TopLog`
- `my-python-app/pygame-playground`
- `Desktop/game/itchio_eternum`

**Proposed change:** add this to `.claude/settings.local.json` in each of those 5 projects. That file is local only and not committed. Where a file already exists, I'd merge the key in.
```json
{ "enabledPlugins": { "superpowers@claude-plugins-official": false } }
```
**This will save little.** Those 5 projects had about 17 sessions in total, so almost all of the 25% comes from projects that do use superpowers. If you want a real cut, the better move is to turn it off globally in `~/.claude/settings.json` and turn it back on only where you want it, for example commencis and backend-playground.

## 3. MCP servers not used in the last month

I checked every MCP tool call in your transcripts in `~/.claude/projects` (dates are from file modification times).

**In your settings, never called:**

| Server | Where it is defined | Proposed change |
|---|---|---|
| `github` (plugin) | the `github` plugin; it fails to connect every session | `~/.claude/settings.json`: `"github@claude-plugins-official": false` (the plugin contains only this server) |
| `figma-desktop`, `figma-remote-mcp` | `~/.claude.json`, project `~` | `"disabledMcpServers": ["figma-desktop","figma-remote-mcp"]` on that project |
| `analytics-mcp` | `~/.claude.json`, project `analytics_agent_tool/analytics_agent` | `"disabledMcpServers": ["analytics-mcp"]` on that project |

**Used, so I'll keep them:** firebase, and the claude.ai connectors Claude Docs, Figma, Meta Ads, Context7 and Google Drive.

**Unused, but not in a file I can change cleanly:**
- **claude.ai connectors** (Linear, Microsoft Learn, HyperFrames, Cloudflare, Firecrawl, Gmail, Google Calendar, Hugging Face): Claude Code can only turn these off one project at a time (38 project entries), or all of them at once with `"disableClaudeAiConnectors": true`, which would also cut the ones you use. It's cleaner to disconnect them in your claude.ai connector settings.
- **`plugin:marketing:*`, `plugin:sales:*`, `plugin:data:*` and similar:** these plugins aren't in your `enabledPlugins` or the local plugin cache. They seem to come from your claude.ai or organization account, so you'd turn them off there.
- **Desktop-app servers** (`ccd_*`, `Claude_Browser`, `terminal`, `visualize`): these are built into the Claude app, and you use them.

Tell me which of these to apply: 1, 2 (per project or global), and the rows in 3. I'll make only those changes and show you the result.

