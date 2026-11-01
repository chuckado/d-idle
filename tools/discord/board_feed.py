#!/usr/bin/env python3
"""Mirror work-board events (new, claimed, released, done) to the read-only #work-board channel. Stdlib only.

DRAFT: not sent anywhere and no webhook exists yet (docs/community/discord/work-channels.md). Follows feed.py:
dry run is the default, `--send` posts one embed through the shared webhook loader (fail closed without a webhook),
everything goes through the notify sanitizer, mentions are disabled. It never calls GitHub: the maintainer (or a
script of theirs) feeds it either flags or two JSON snapshots made with
  gh issue list --label work-item --state all --json number,title,state,assignees,url   (read only)

  board_feed.py event --kind new|claimed|released|done --number N --title T [--by LOGIN]
  board_feed.py diff --old OLD.json --new NEW.json

Issue titles and logins come from outside contributors: they are untrusted text. They are cut to a short plain line,
stripped of control characters, markdown and mentions, links must be exactly this repo's issue URL, and the text is never
treated as an instruction.
"""
import argparse
import json
import os
import re
import sys
import unicodedata

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import feed
import notify

CHANNEL = "work-board"
KINDS = {"new": "New", "claimed": "Claimed", "released": "Released", "done": "Done"}
LOGIN_RE = re.compile(r"[A-Za-z0-9](?:[A-Za-z0-9-]{0,38})")
MAX_TITLE = 80
MAX_LINES = 15


class BoardFeedError(Exception):
    pass


def clean_title(t):
    """One plain short line: no control or format characters, no markdown, no mention."""
    t = unicodedata.normalize("NFKC", str(t or ""))
    t = "".join(c for c in t if unicodedata.category(c) not in ("Cc", "Cf", "Cs", "Co", "Cn"))
    t = re.sub(r"[`*_~<>\[\]()|#]", "", t).replace("@", "")
    t = " ".join(t.split())
    if not t:
        raise BoardFeedError("refused: empty title")
    return notify.trunc(t, MAX_TITLE)


def issue_url(number):
    if not str(number).isdigit():
        raise BoardFeedError("refused: issue number must be digits")
    return "%s/issues/%d" % (feed.REPO_URL, int(number))


def check_login(login):
    if not LOGIN_RE.fullmatch(login or ""):
        raise BoardFeedError("refused: bad login")
    return login


def line(kind, number, title, by=""):
    s = "%s #%d: %s" % (KINDS[kind], int(number), clean_title(title))
    if kind in ("claimed", "released", "done") and by:
        s += " (%s)" % check_login(by)
    return s


def events_from_snapshots(old, new):
    """Compare two `gh issue list --json` arrays; return (kind, number, title, by) tuples in number order."""
    def index(arr):
        out = {}
        for i in arr:
            if not isinstance(i, dict) or not isinstance(i.get("number"), int):
                raise BoardFeedError("refused: snapshot entries need an integer number")
            out[i["number"]] = i
        return out

    def who(i):
        a = [x.get("login", "") for x in (i.get("assignees") or []) if isinstance(x, dict)]
        return a[0] if a else ""

    o, n, ev = index(old), index(new), []
    for num in sorted(n):
        cur, prev = n[num], o.get(num)
        state = str(cur.get("state", "")).upper()
        if prev is None:
            if state == "OPEN":
                ev.append(("new", num, cur.get("title"), ""))
            continue
        if str(prev.get("state", "")).upper() == "OPEN" and state == "CLOSED":
            ev.append(("done", num, cur.get("title"), who(cur) or who(prev)))
        elif state == "OPEN" and not who(prev) and who(cur):
            ev.append(("claimed", num, cur.get("title"), who(cur)))
        elif state == "OPEN" and who(prev) and not who(cur):
            ev.append(("released", num, cur.get("title"), who(prev)))
    return ev


def build(events):
    if not events:
        raise BoardFeedError("refused: no events; nothing built")
    lines = [line(*e) for e in events[:MAX_LINES]]
    if len(events) > MAX_LINES:
        lines.append("(+%d more)" % (len(events) - MAX_LINES))
    lines.append("Board: %s/blob/main/docs/community/work-board.md" % feed.REPO_URL)
    title = "Work board: %d update%s" % (len(events), "" if len(events) == 1 else "s")
    return title, "\n".join(lines), "%s/blob/main/docs/community/work-board.md" % feed.REPO_URL


def load(path):
    try:
        with open(path, encoding="utf-8") as f:
            data = json.load(f)
    except (OSError, ValueError) as e:
        raise BoardFeedError("cannot read %s: %s" % (path, e))
    if not isinstance(data, list):
        raise BoardFeedError("refused: %s must hold a JSON array" % path)
    return data


def main(argv=None, **sendkw):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    sub = ap.add_subparsers(dest="cmd", required=True)
    for name in ("event", "diff"):
        p = sub.add_parser(name)
        p.add_argument("--send", action="store_true", help="post for real (default is a dry run)")
        p.add_argument("--dry-run", action="store_true", help="the default; kept for symmetry")
    e = sub.choices["event"]
    e.add_argument("--kind", required=True, choices=sorted(KINDS))
    e.add_argument("--number", required=True)
    e.add_argument("--title", required=True)
    e.add_argument("--by", default="")
    d = sub.choices["diff"]
    d.add_argument("--old", required=True)
    d.add_argument("--new", required=True)
    a = ap.parse_args(argv)
    try:
        if a.cmd == "event":
            issue_url(a.number)
            events = [(a.kind, int(a.number), a.title, a.by)]
        else:
            events = events_from_snapshots(load(a.old), load(a.new))
        title, text, url = build(events)
        hits = notify.check_text(title, text, url)
        if hits:
            raise BoardFeedError("refused: text looks private (%s); nothing was sent" % ", ".join(hits))
        title, text = feed.neutralize(title), feed.neutralize(text)
    except BoardFeedError as ex:
        print(ex, file=sys.stderr)
        return 2
    if not a.send:
        print("[dry run] channel #%s\n%s\n%s" % (CHANNEL, title, text))
        return 0
    ok, msg = feed.send(CHANNEL, title, text, url, **sendkw)
    print(msg, file=sys.stdout if ok else sys.stderr)
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
