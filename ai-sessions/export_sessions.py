#!/usr/bin/env python3
"""Convert Claude Code JSONL transcripts for this project into readable Markdown.

Usage: python3 ai-sessions/export_sessions.py
Writes one .md per session (subagent runs appended) into ai-sessions/.
Redacts email, home path, account ids and anything that looks like a secret.
"""
import json
import re
import sys
from pathlib import Path

HOME = str(Path.home())
PROJECT_DIR = Path(HOME) / ".claude/projects/-Users-kaanaray-Desktop-projects-case-studies-insider-one"
OUT_DIR = Path(__file__).resolve().parent
MAX_RESULT = 2000
MAX_INPUT = 3000

REDACTIONS = [
    (re.compile(r"sk-ant-[A-Za-z0-9_\-]{10,}"), "[REDACTED_API_KEY]"),
    (re.compile(r"\b(ANTHROPIC_API_KEY|API_KEY|SECRET|TOKEN|PASSWORD)\s*[=:]\s*['\"]?(?!\$|\{|<|your|xxx|changeme|\[)[^\s'\"]{8,}", re.I), r"\1=[REDACTED]"),
    (re.compile(r"\b(gh[pousr]_[A-Za-z0-9]{20,}|xox[abp]-[A-Za-z0-9\-]{10,}|AKIA[0-9A-Z]{16})\b"), "[REDACTED_TOKEN]"),
    (re.compile(r"[A-Za-z0-9._%+\-]+@(gmail|hotmail|outlook|yahoo|icloud)\.com", re.I), "[REDACTED_EMAIL]"),
    (re.compile(re.escape(HOME)), "~"),
    (re.compile(r"-Users-kaanaray-"), "-Users-USER-"),
    (re.compile(r"kaanaray"), "USER"),
]
STRIP_TAGS = re.compile(
    r"<(system-reminder|EXTREMELY_IMPORTANT|EXTREMELY-IMPORTANT|SUBAGENT-STOP|local-command-caveat)>.*?</\1>",
    re.S,
)
IDE_TAG = re.compile(r"<ide_(opened_file|selection)>(.*?)</ide_\1>", re.S)
CMD_TAG = re.compile(r"<command-(name|message|args)>(.*?)</command-\1>", re.S)
STDOUT_TAG = re.compile(r"<local-command-stdout>(.*?)</local-command-stdout>", re.S)


def redact(s: str) -> str:
    for pat, rep in REDACTIONS:
        s = pat.sub(rep, s)
    return s


def clip(s: str, n: int) -> str:
    s = s.rstrip()
    return s if len(s) <= n else s[:n] + f"\n… [{len(s) - n} more chars truncated]"


def fence(s: str, lang: str = "") -> str:
    ticks = "````" if "```" in s else "```"
    return f"{ticks}{lang}\n{s}\n{ticks}"


PASTE_TAG = re.compile(r"</?pasted_content[^>]*>")


def clean_user_text(t: str) -> str:
    t = STRIP_TAGS.sub("", t)
    t = PASTE_TAG.sub("", t)
    t = IDE_TAG.sub(lambda m: f"_[IDE: {m.group(1).replace('_', ' ')}]_ ", t)
    cmds = CMD_TAG.findall(t)
    if cmds:
        parts = {k: v.strip() for k, v in cmds}
        t = CMD_TAG.sub("", t)
        name = parts.get("name", "")
        args = parts.get("args", "")
        t = f"`{name} {args}`".replace(" `", "`") + ("\n" + t if t.strip() else "")
    t = STDOUT_TAG.sub(lambda m: "\n" + fence(clip(m.group(1).strip(), MAX_RESULT)) if m.group(1).strip() else "", t)
    return t.strip()


def result_text(content) -> str:
    if isinstance(content, str):
        return content
    out = []
    for c in content or []:
        if c.get("type") == "text":
            out.append(c["text"])
        elif c.get("type") == "image":
            out.append("[image]")
        elif c.get("type") == "tool_reference":
            out.append(f"[tool loaded: {c.get('tool_name')}]")
    return "\n".join(out)


def tool_input(name: str, inp: dict) -> str:
    if name == "Bash":
        desc = inp.get("description", "")
        return (f"_{desc}_\n" if desc else "") + fence(clip(inp.get("command", ""), MAX_INPUT), "bash")
    if name in ("Write",):
        return f"`{inp.get('file_path')}`\n" + fence(clip(inp.get("content", ""), MAX_INPUT))
    if name == "Edit":
        return (f"`{inp.get('file_path')}`\n" + "**old:**\n" + fence(clip(inp.get("old_string", ""), MAX_INPUT // 2))
                + "\n**new:**\n" + fence(clip(inp.get("new_string", ""), MAX_INPUT // 2)))
    if name == "Read":
        return f"`{inp.get('file_path')}`" + (f" (offset {inp['offset']})" if inp.get("offset") else "")
    if name in ("Agent", "Task"):
        return (f"**{inp.get('description', '')}** ({inp.get('subagent_type', 'general-purpose')})\n"
                + fence(clip(inp.get("prompt", ""), MAX_INPUT)))
    return fence(clip(json.dumps(inp, ensure_ascii=False, indent=2), MAX_INPUT), "json")


def load(path: Path):
    recs = []
    for line in path.open(encoding="utf-8"):
        try:
            recs.append(json.loads(line))
        except json.JSONDecodeError:
            continue
    return recs


def render(recs, heading_level=2) -> tuple[str, int]:
    h = "#" * heading_level
    out = []
    turns = 0
    for d in recs:
        t = d.get("type")
        if t not in ("user", "assistant"):
            continue
        msg = d.get("message") or {}
        content = msg.get("content")
        ts = (d.get("timestamp") or "")[:19].replace("T", " ")
        if t == "user":
            if d.get("isMeta"):
                text = result_text(content) if isinstance(content, list) else content or ""
                first = text.strip().splitlines()[0] if text.strip() else ""
                if first.startswith("Base directory for this skill"):
                    out.append(f"> _[skill instructions loaded: {first.split('/')[-1]}]_\n")
                continue
            if isinstance(content, str):
                text = clean_user_text(content)
                if text:
                    turns += 1
                    out.append(f"{h} 👤 User  <sub>{ts}</sub>\n\n{text}\n")
                continue
            texts = [c["text"] for c in content if c.get("type") == "text"]
            for c in content:
                if c.get("type") == "tool_result":
                    r = clip(STRIP_TAGS.sub("", result_text(c.get("content"))).strip(), MAX_RESULT)
                    label = "Result (error)" if c.get("is_error") else "Result"
                    out.append(f"<details><summary>{label}</summary>\n\n{fence(r) if r else '_(empty)_'}\n\n</details>\n")
            text = clean_user_text("\n".join(texts))
            if text:
                turns += 1
                out.append(f"{h} 👤 User  <sub>{ts}</sub>\n\n{text}\n")
        else:
            for c in content or []:
                ct = c.get("type")
                if ct == "text" and c["text"].strip():
                    out.append(f"{h} 🤖 Claude  <sub>{ts}</sub>\n\n{c['text'].strip()}\n")
                elif ct == "thinking" and c.get("thinking", "").strip():
                    out.append(f"<details><summary>Thinking</summary>\n\n{c['thinking'].strip()}\n\n</details>\n")
                elif ct == "tool_use":
                    out.append(f"**🔧 {c['name']}** {tool_input(c['name'], c.get('input') or {})}\n")
    return "\n".join(out), turns


def meta(recs):
    title = next((d["aiTitle"] for d in reversed(recs) if d.get("type") == "ai-title"), None)
    stamps = [d["timestamp"] for d in recs if d.get("timestamp") and d.get("type") in ("user", "assistant")]
    model = next((d["message"].get("model") for d in recs if d.get("type") == "assistant"), "")
    return title, (stamps[0] if stamps else ""), (stamps[-1] if stamps else ""), model


def slug(s: str) -> str:
    return re.sub(r"[^a-z0-9]+", "-", s.lower()).strip("-")[:60] or "session"


def main():
    sessions = []
    for p in sorted(PROJECT_DIR.glob("*.jsonl")):
        recs = load(p)
        for d in recs:
            for key in ("accountUuid", "organizationUuid"):
                v = d.get(key)
                if v and not any(pat.pattern == re.escape(v) for pat, _ in REDACTIONS):
                    REDACTIONS.append((re.compile(re.escape(v)), f"[REDACTED_{key.upper()}]"))
        body, turns = render(recs)
        real = any(d.get("type") == "assistant" and d["message"].get("model") != "<synthetic>" for d in recs)
        if turns == 0 or not real:
            continue
        title, start, end, model = meta(recs)
        sessions.append((start, p, recs, body, title or p.stem, end, model))
    sessions.sort(key=lambda s: s[0])

    for f in OUT_DIR.glob("[0-9][0-9]-*.md"):
        f.unlink()

    index = ["# Claude Code sessions\n",
             "Exported from Claude Code JSONL transcripts by `export_sessions.py`. "
             "Hidden system context is removed. Long tool output is truncated. "
             "Email, home path and secrets are redacted.\n",
             "| # | Session | Start (UTC) | End (UTC) | Subagents |", "|---|---|---|---|---|"]
    for i, (start, p, recs, body, title, end, model) in enumerate(sessions, 1):
        sub_dir = p.with_suffix("") / "subagents"
        subs = []
        if sub_dir.is_dir():
            for sp in sorted(sub_dir.glob("*.jsonl")):
                mp = sp.with_suffix(".meta.json")
                info = json.loads(mp.read_text()) if mp.exists() else {}
                srecs = load(sp)
                sbody, sturns = render(srecs, heading_level=4)
                if not sbody.strip():
                    continue
                label = info.get("description") or info.get("agentType") or sp.stem
                stamps = [d["timestamp"] for d in srecs if d.get("timestamp")]
                subs.append((stamps[0] if stamps else "", label, info.get("agentType", ""), sbody))
            subs.sort()
        name = f"{i:02d}-{slug(title)}.md"
        doc = [f"# {title}\n",
               f"- Session: `{p.stem}`",
               f"- Model: `{model}`",
               f"- Time (UTC): {start[:19].replace('T', ' ')} → {end[:19].replace('T', ' ')}",
               f"- Subagent runs: {len(subs)}\n", "---\n", body]
        if subs:
            doc.append("\n---\n\n## Subagent transcripts\n")
            for _, label, atype, sbody in subs:
                doc.append(f"\n### Subagent: {label}" + (f" ({atype})" if atype else "") + "\n\n" + sbody)
        (OUT_DIR / name).write_text(redact("\n".join(doc)) + "\n", encoding="utf-8")
        index.append(f"| {i} | [{redact(title)}]({name}) | {start[:16].replace('T', ' ')} | "
                     f"{end[:16].replace('T', ' ')} | {len(subs)} |")
        print(f"wrote {name}", file=sys.stderr)
    (OUT_DIR / "README.md").write_text("\n".join(index) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
