#!/usr/bin/env python3
"""Check Markdown files against a subset of ASD-STE100 rules.

Usage: ste_check.py <file.md>...

Prints one line per finding: file:line: rule: text
Always exits 0. The findings are warnings for a human or agent to review.
"""

import re
import sys
from typing import NamedTuple

MAX_WORDS = 25

FENCE = re.compile(r"^\s*(`{3,}|~{3,})")
INLINE_CODE = re.compile(r"`[^`]*`")
LIST_MARKER = re.compile(r"^\s*(?:[-*+]|\d+[.)])\s+")
BLOCKQUOTE = re.compile(r"^\s*>\s?")
# Sentence end: . ! ? with optional closing quotes or brackets, then whitespace.
SENTENCE_END = re.compile(r"([.!?][\"')\]]*)\s+")
NOT_SEMICOLON = re.compile(r"&#?\w+;|https?://\S+")
PRESENT_PERFECT = [
    re.compile(r"\b(has|have|had) been\b", re.IGNORECASE),
    re.compile(r"\b(has|have) [a-z]+ed\b", re.IGNORECASE),
]


class Finding(NamedTuple):
    line: int
    rule: str
    text: str


def prose_lines(text):
    """Yield (line number, line) outside fenced code, with inline code replaced.

    A fence closes only on a line with the same marker character, at least the
    same length, and no info string (CommonMark rule).
    """
    fence = None
    for number, line in enumerate(text.splitlines(), start=1):
        match = FENCE.match(line)
        if fence is None:
            if match:
                fence = match.group(1)
                continue
        else:
            marker = match.group(1) if match else ""
            closes = (
                marker
                and marker[0] == fence[0]
                and len(marker) >= len(fence)
                and not line.strip()[len(marker):].strip()
            )
            if closes:
                fence = None
            continue
        yield number, INLINE_CODE.sub("CODE", line)


def blocks(lines):
    """Group prose lines into blocks that can hold wrapped sentences.

    A blank line, a heading, a table row or a new list item ends a block.
    Headings and table rows are not part of any block.
    Yields (first line number, text).
    """
    start, parts = None, []
    for number, line in lines:
        stripped = BLOCKQUOTE.sub("", line).strip()
        is_item = bool(LIST_MARKER.match(stripped))
        if not stripped or stripped.startswith(("|", "#")) or is_item:
            if parts:
                yield start, " ".join(parts)
            start, parts = None, []
            if not stripped or not is_item:
                continue
            stripped = LIST_MARKER.sub("", stripped)
        if start is None:
            start = number
        parts.append(stripped)
    if parts:
        yield start, " ".join(parts)


def check_text(text):
    lines = list(prose_lines(text))
    findings = []

    for number, line in lines:
        stripped = line.strip()
        if ";" in NOT_SEMICOLON.sub("", stripped):
            findings.append(Finding(number, "semicolon", stripped))
        for pattern in PRESENT_PERFECT:
            match = pattern.search(stripped)
            if match:
                findings.append(Finding(number, "present-perfect", match.group(0)))
                break

    for number, block in blocks(lines):
        for sentence in SENTENCE_END.sub("\\1\n", block).split("\n"):
            count = len(sentence.split())
            if count > MAX_WORDS:
                findings.append(
                    Finding(number, "sentence-length", f"{count} words: {sentence}")
                )

    findings.sort(key=lambda finding: finding.line)
    return findings


def main(paths):
    for path in paths:
        try:
            with open(path, encoding="utf-8") as handle:
                text = handle.read()
        except (OSError, UnicodeDecodeError) as error:
            print(f"{path}:0: read-error: {error}")
            continue
        for finding in check_text(text):
            print(f"{path}:{finding.line}: {finding.rule}: {finding.text}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
