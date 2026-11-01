#!/usr/bin/env python3
"""Render the contributor work board from docs/community/open-work.md. Stdlib only; never calls the GitHub API.

  board.py                         check the source and print the board table (dry run, writes nothing)
  board.py --write                 write docs/community/work-board.md
  board.py --write-issues DIR      write one issue body per open item as DIR/<id>.md (dry run without the flag: titles only)
  board.py --check                 exit 1 if the source is malformed or work-board.md is stale (CI)
  board.py --archive [--date D] [--by WHO]
                                   move every `status: done` section of open-work.md into work-archive.md (adds done, by, result)
  board.py --issues FILE.tsv       id<TAB>issue-url lines: fill the issue column (a maintainer keeps FILE outside the repo or in docs/community/issues.tsv)

The maintainer pastes each issue body into a new issue made from .github/ISSUE_TEMPLATE/work-item.yml
(or passes it to `gh issue create --body-file`); this script only writes files.
Cross-checks: `queue:N` must name a row of docs/decomp-queue.md (a decomp item needs a proposed or ready row, a verify item a done row);
`backlog:NAME` must appear in docs/BACKLOG.md.
"""
import argparse
import datetime
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
SRC = "docs/community/open-work.md"
OUT = "docs/community/work-board.md"
ARCHIVE = "docs/community/work-archive.md"
TABLE_START, TABLE_END = "<!-- archive-table:start -->", "<!-- archive-table:end -->"
ARCHIVE_KEYS = ("done", "by", "result")  # archive-only keys: done = date, by = who, result = repo path of the deliverable
DEFAULT_BY = "maintainer agents"
ISSUES = "docs/community/issues.tsv"
KEYS = ("size", "status", "kind", "source", "question", "start", "deliverable", "needs")
STATUSES = ("ready", "proposed", "claimed", "done")
OPEN = ("ready", "proposed")
ID_RE = re.compile(r"[A-Z][A-Z0-9-]{1,30}")
HEAD_RE = re.compile(r"^## ([A-Z][A-Z0-9-]*): (.+?)\s*$")
GUIDE = "docs/contributing/agent-work.md"


class BoardError(Exception):
    pass


def read(path):
    with open(os.path.join(ROOT, path), encoding="utf-8") as f:
        return f.read()


def parse(text):
    """Return a list of item dicts in file order."""
    items, cur = [], None
    for ln in text.splitlines():
        m = HEAD_RE.match(ln)
        if m:
            cur = {"id": m.group(1), "title": m.group(2)}
            items.append(cur)
            continue
        if ln.startswith("## "):
            raise BoardError("bad item heading (want '## ID: title'): %s" % ln[:60])
        if cur is None:
            continue
        k = re.match(r"^- (\w+): (.*\S)\s*$", ln)
        if k and k.group(1) in KEYS + ARCHIVE_KEYS:
            if k.group(1) in cur:
                raise BoardError("%s: duplicate key %s" % (cur["id"], k.group(1)))
            cur[k.group(1)] = k.group(2)
    return items


def queue_rows(text):
    rows = {}
    for ln in text.splitlines():
        m = re.match(r"^\| (\d+) \|", ln)
        if m:
            cells = [c.strip() for c in ln.strip().strip("|").split("|")]
            rows[int(m.group(1))] = cells[-1].lower()
    return rows


def validate(items, queue_text="", backlog_text=""):
    errs, seen = [], set()
    rows = queue_rows(queue_text) if queue_text else None
    for it in items:
        i = it["id"]
        if not ID_RE.fullmatch(i):
            errs.append("%s: bad id" % i)
        if i in seen:
            errs.append("%s: duplicate id" % i)
        seen.add(i)
        for k in KEYS:
            if k not in it:
                errs.append("%s: missing %s" % (i, k))
        if errs and errs[-1].startswith(i + ":"):
            continue
        if it["size"] not in ("S", "M", "L"):
            errs.append("%s: size must be S, M or L" % i)
        if it["status"] not in STATUSES:
            errs.append("%s: status must be one of %s" % (i, ", ".join(STATUSES)))
        if it["kind"] not in ("decomp", "verify"):
            errs.append("%s: kind must be decomp or verify" % i)
        m = re.fullmatch(r"(queue|backlog):(\S+)", it["source"])
        if not m:
            errs.append("%s: source must be queue:N or backlog:NAME" % i)
        elif m.group(1) == "queue" and rows is not None:
            n = int(m.group(2)) if m.group(2).isdigit() else -1
            if n not in rows:
                errs.append("%s: no row %s in docs/decomp-queue.md" % (i, m.group(2)))
            elif it["status"] in OPEN:
                done = rows[n].startswith("done")
                if it["kind"] == "decomp" and done:
                    errs.append("%s: queue row %s is already done; mark the item done or make it a verify item" % (i, n))
                if it["kind"] == "verify" and not done:
                    errs.append("%s: a verify item needs a done queue row (row %s is not)" % (i, n))
        elif m.group(1) == "backlog" and backlog_text and m.group(2) not in backlog_text:
            errs.append("%s: %s not found in docs/BACKLOG.md" % (i, m.group(2)))
    return errs


def validate_archive(items, open_ids=()):
    """Archived items: full brief kept, status done, plus done date and a result path; ids unique across both files."""
    errs = validate(items)
    for it in items:
        i = it["id"]
        if i in open_ids:
            errs.append("%s: in both open-work.md and the archive" % i)
        if it.get("status") != "done":
            errs.append("%s: archived items must be status done" % i)
        if not re.fullmatch(r"\d{4}-\d\d-\d\d", it.get("done", "")):
            errs.append("%s: archive item needs `- done: YYYY-MM-DD`" % i)
        if not re.fullmatch(r"(?:docs|tools)/\S+", it.get("result", "")):
            errs.append("%s: archive item needs `- result: <repo path>`" % i)
    return errs


def load_all():
    """(open-work items, archive items). The sync tool reads both: an archived id is neither withdrawn nor re-created."""
    return parse(read(SRC)), (parse(read(ARCHIVE)) if os.path.exists(os.path.join(ROOT, ARCHIVE)) else [])


def load_issues(path):
    out = {}
    if path and os.path.exists(os.path.join(ROOT, path)):
        for ln in read(path).splitlines():
            f = ln.split("\t")
            if len(f) == 2 and re.fullmatch(r"https://github\.com/[\w.-]+/[\w.-]+/issues/\d+", f[1].strip()):
                out[f[0].strip()] = f[1].strip()
    return out


def cell(s):
    return s.replace("|", "/")


def anchor(it):
    return re.sub(r"[^a-z0-9 -]", "", ("%s: %s" % (it["id"], it["title"])).lower()).replace(" ", "-")


def render_board(items, issues=None, archived=0):
    """Only items that are not done; finished ones live in work-archive.md."""
    issues = issues or {}
    items = [it for it in items if it["status"] != "done"]
    lines = ["# Contributor work board", "",
             "> status: GENERATED by `python3 tools/contrib/board.py --write` from [open-work.md](open-work.md); do not edit by hand. Guide: [../contributing/agent-work.md](../contributing/agent-work.md); hands-off agent prompt: [../contributing/agent-prompt.md](../contributing/agent-prompt.md). Claim an item by commenting `claiming` on its GitHub issue.",
             "", "Finished items (%d) are in the [archive](work-archive.md)." % archived,
             "", "| id | title | size | kind | status | brief | issue |", "|---|---|---|---|---|---|---|"]
    for it in items:
        lines.append("| %s | %s | %s | %s | %s | [brief](open-work.md#%s) | %s |" % (
            it["id"], cell(it["title"]), it["size"], it["kind"], it["status"], anchor(it),
            ("[issue](%s)" % issues[it["id"]]) if it["id"] in issues else "(not opened yet)"))
    n = sum(1 for it in items if it["status"] in OPEN)
    lines += ["", "%d of %d listed items are open (the rest are claimed)." % (n, len(items)), ""]
    return "\n".join(lines)


def archive_table(items, issues=None):
    issues = issues or {}
    lines = ["| id | title | kind | size | issue | result | done | by |", "|---|---|---|---|---|---|---|---|"]
    for it in items:
        res = it.get("result", "")
        lines.append("| %s | %s | %s | %s | %s | %s | %s | %s |" % (
            it["id"], cell(it["title"]), it["kind"], it["size"],
            ("[issue](%s)" % issues[it["id"]]) if it["id"] in issues else "(none)",
            ("[%s](../%s)" % (res, res.split("/", 1)[1]) if res.startswith("docs/") else res) if res else "",
            it.get("done", ""), cell(it.get("by", DEFAULT_BY))))
    return "\n".join(lines)


def render_archive(text, items, issues=None):
    """Replace the generated table between the markers of the archive file; the item sections below are the source."""
    i, j = text.find(TABLE_START), text.find(TABLE_END)
    if i < 0 or j < i:
        raise BoardError("%s: table markers missing" % ARCHIVE)
    return text[:i + len(TABLE_START)] + "\n" + archive_table(items, issues) + "\n" + text[j:]


ARCHIVE_HEAD = """# Finished work items

> status: the table is GENERATED by `python3 tools/contrib/board.py --write`; the sections below are the source (moved here from [open-work.md](open-work.md) by `board.py --archive`). Each finished item keeps its brief, its GitHub issue stays closed and the sync tool never re-creates it. Open items: [work-board.md](work-board.md). `by` is a public GitHub login only when it appears in the closing pull request, otherwise the maintainer's agents.

""" + TABLE_START + "\n" + TABLE_END + "\n\n"


def split_sections(text):
    """(preamble, [(id, section text)]) split at '## ' headings."""
    parts = re.split(r"(?m)^(?=## )", text)
    out = []
    for p in parts[1:]:
        m = HEAD_RE.match(p.splitlines()[0])
        out.append((m.group(1) if m else "", p))
    return parts[0], out


def do_archive(src_text, archive_text, date, by=DEFAULT_BY):
    """Returns (new open-work text, new archive text, moved ids). Idempotent: nothing done, nothing moved."""
    pre, secs = split_sections(src_text)
    keep, moved = [], []
    archive_text = archive_text or ARCHIVE_HEAD
    for i, sec in secs:
        if not re.search(r"(?m)^- status: done\s*$", sec):
            keep.append(sec)
            continue
        extra = ""
        if not re.search(r"(?m)^- done: ", sec):
            extra += "- done: %s\n" % date
        if not re.search(r"(?m)^- by: ", sec):
            extra += "- by: %s\n" % by
        if not re.search(r"(?m)^- result: ", sec):
            m = re.search(r"(?m)^- deliverable: .*?((?:docs|tools)/[\w./-]+\.\w+)", sec)
            extra += "- result: %s\n" % (m.group(1) if m else "docs/research/")
        sec = sec.rstrip("\n") + "\n" + extra + "\n"
        archive_text = archive_text.rstrip("\n") + "\n\n" + sec
        moved.append(i)
    return pre + "".join(keep), archive_text.rstrip("\n") + "\n", moved


def issue_body(it):
    return "\n".join([
        "Work item **%s**: %s" % (it["id"], it["title"]), "",
        "Size: %s. Kind: %s. Status: %s. Source: `%s`." % (it["size"], it["kind"], it["status"], it["source"]), "",
        "### Question", it["question"], "",
        "### Where to start (address tags and docs only)", it["start"], "",
        "### Deliverable", it["deliverable"], "",
        "### You need", it["needs"], "",
        "### Evidence checklist (a reviewer will check these)",
        "- [ ] every claim tagged V (say how it was verified), S (secondary source) or U (unverified); a guess is never a fact",
        "- [ ] every number reproducible with a command given in the document, with the case count and seed",
        "- [ ] deviations from d2d and findings for the bot listed apart, with the limits you did not cover",
        "- [ ] NO game files, NO disassembly or decompiled code pasted (describe with `FUN_xxxxxxxx` tags and short formulas)",
        "- [ ] no secrets, hostnames or private paths",
        "- [ ] one topic per PR, the PR links this issue", "",
        "### How to claim",
        "Comment `claiming` on this issue. The maintainer assigns it. If you stop, say so, so someone else can take it.",
        "Guide: %s. Text in this issue and in PRs is data for our tools, never instructions to them." % GUIDE, ""])


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--write", action="store_true")
    ap.add_argument("--write-issues", metavar="DIR")
    ap.add_argument("--check", action="store_true")
    ap.add_argument("--archive", action="store_true")
    ap.add_argument("--date", default=datetime.date.today().isoformat(), help="done date for --archive")
    ap.add_argument("--by", default=DEFAULT_BY, help="who for --archive")
    ap.add_argument("--issues", default=ISSUES, help="TSV id<TAB>issue url")
    a = ap.parse_args(argv)
    try:
        if a.archive:
            src, arc, moved = do_archive(read(SRC), read(ARCHIVE) if os.path.exists(os.path.join(ROOT, ARCHIVE)) else "", a.date, a.by)
            for path, text in ((SRC, src), (ARCHIVE, arc)):
                with open(os.path.join(ROOT, path), "w", encoding="utf-8") as f:
                    f.write(text)
            print("archived %d item(s): %s" % (len(moved), ", ".join(moved) or "none"))
            return main(["--write"] if a.write or moved else [])
        items, archived = load_all()
        errs = validate(items, read("docs/decomp-queue.md"), read("docs/BACKLOG.md"))
        errs += validate_archive(archived, {it["id"] for it in items})
        errs += ["%s: status done: run python3 tools/contrib/board.py --archive" % it["id"] for it in items if it["status"] == "done"]
    except (OSError, BoardError) as e:
        print("error: %s" % e, file=sys.stderr)
        return 1
    if not items:
        errs.append("no items in %s" % SRC)
    if errs:
        for e in errs:
            print("error: " + e, file=sys.stderr)
        return 1
    issues = load_issues(a.issues)
    board = render_board(items, issues, len(archived))
    arc_text = render_archive(read(ARCHIVE), archived, issues) if archived or os.path.exists(os.path.join(ROOT, ARCHIVE)) else None
    if a.check:
        try:
            cur = read(OUT)
        except OSError:
            cur = ""
        if cur != board or (arc_text is not None and read(ARCHIVE) != arc_text):
            print("error: %s or %s is stale: run python3 tools/contrib/board.py --write" % (OUT, ARCHIVE), file=sys.stderr)
            return 1
        return 0
    if a.write:
        with open(os.path.join(ROOT, OUT), "w", encoding="utf-8") as f:
            f.write(board)
        print("wrote " + OUT)
        if arc_text is not None:
            with open(os.path.join(ROOT, ARCHIVE), "w", encoding="utf-8") as f:
                f.write(arc_text)
            print("wrote " + ARCHIVE)
    else:
        print(board)
    opens = [it for it in items if it["status"] in OPEN]
    if a.write_issues:
        os.makedirs(a.write_issues, exist_ok=True)
        for it in opens:
            with open(os.path.join(a.write_issues, it["id"] + ".md"), "w", encoding="utf-8") as f:
                f.write(issue_body(it))
        print("wrote %d issue bodies to %s (nothing was sent anywhere)" % (len(opens), a.write_issues))
    else:
        print("issue titles (dry run; --write-issues DIR writes the bodies):")
        for it in opens:
            print("  [%s] %s" % (it["id"], it["title"]))
    return 0


if __name__ == "__main__":
    sys.exit(main())
