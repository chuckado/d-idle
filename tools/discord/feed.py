#!/usr/bin/env python3
"""Build and post short d.idle feed summaries to Discord. Stdlib only; reuses notify.py and common.py.

  feed.py soak     [--file docs/probes/postfix-soak1.md] [--verdict V --duration D --deaths N --restarts N --xph N --kph N]
  feed.py drill    [--file docs/probes/drills/README.md] [--pass N --fail N --blocked N --finding TEXT ...]
  feed.py upstream --pr N --event opened|merged|reviewed|ci-green|changes-requested --title T --url U
  feed.py pr-status [--pr 'N|title|CI|reviews|state' ...]   (or the same lines on stdin)
  feed.py release  --tag T --notes TEXT|--notes-file F --url U

Dry run is the default: the summary is printed as plain text and nothing is sent. `--send` posts it (one embed,
mentions disabled) through the webhook loaded by common.load_webhook_url; the URL is never printed. Everything goes
through the notify sanitizer (home paths, lab names, IPs, tokens, webhook URLs, emails are refused). Only facts
found in our docs or given as flags are used; a soak summary without a verdict line is refused.
"""
import argparse
import os
import re
import subprocess
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import common as dc
import notify

REPO_URL = "https://github.com/shatteredeagle/d-idle"
UPSTREAM_PR_RE = re.compile(r"https://github\.com/psi29a/d2decomped/pull/(\d+)")
RELEASE_URL_RE = re.compile(re.escape(REPO_URL) + r"/releases/tag/([A-Za-z0-9._-]+)")
EVENTS = {"opened": "opened", "merged": "merged", "reviewed": "reviewed", "ci-green": "CI is green",
          "changes-requested": "changes requested"}
DEFAULT_CHANNEL = {"soak": "progress", "drill": "progress", "pr-status": "progress",
                   "release": "announcements", "upstream": "d2d-upstream"}
CHANNELS = ("progress", "announcements", "d2d-upstream", "stream")
NO_HOOK = "no webhook file for %s yet (needs the user go to create one)"
ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


class FeedError(Exception):
    pass


def read(path):
    try:
        with open(path, encoding="utf-8") as f:
            return f.read()
    except OSError as e:
        raise FeedError("cannot read %s: %s" % (path, e.strerror))


def repo_link(path):
    rel = os.path.relpath(os.path.abspath(path), ROOT) if os.path.isabs(path) or os.path.exists(path) else path
    if rel.startswith(".."):
        raise FeedError("the report must be inside the repo to link it: %s" % path)
    return "%s/blob/main/%s" % (REPO_URL, rel.replace(os.sep, "/"))


def first(rx, text, flags=re.I | re.M):
    m = re.search(rx, text, flags)
    return m if m else None


def build_soak(a):
    text = read(a.file) if a.file else ""
    verdict = a.verdict
    if not verdict:
        m = first(r"^\W*verdict\W*[:=]\s*(.+?)\s*$", text) or first(r"\bOVERALL\s+(PASS|FAIL|INCONCLUSIVE)\b", text, re.M)
        verdict = m.group(1).strip() if m else ""
    if not verdict:
        raise FeedError("refused: the soak report has no verdict line (and no --verdict); nothing built")
    dur = a.duration
    if not dur:
        w = first(r"\b(\d+)\s*wall min", text)
        v = first(r"=\s*(\d+(?:\.\d+)?)\s*virtual h", text)
        dur = ", ".join(x for x in (w and "%s wall min" % w.group(1), v and "%s virtual h" % v.group(1)) if x)
    deaths = a.deaths
    if deaths is None:
        m = first(r"\b(\d+)\s+deaths?\b", text)
        deaths = m.group(1) if m else None
    restarts = a.restarts
    if restarts is None:
        m = first(r"restarts\W+counter\s+(\d+)", text)
        restarts = m.group(1) if m else None
    xph, kph = a.xph, a.kph
    if xph is None or kph is None:
        m = first(r"XP per vh, kills per vh\s*\|[^|]*\|\s*([\d,]+)\s*,\s*([\d,]+)\s*\|", text)
        if m:
            xph = xph if xph is not None else m.group(1).replace(",", "")
            kph = kph if kph is not None else m.group(2).replace(",", "")
    lines = ["Verdict: %s" % verdict]
    if dur:
        lines.append("Duration: %s" % dur)
    stats = []
    if deaths is not None:
        stats.append("deaths %s" % deaths)
    if restarts is not None:
        stats.append("restarts %s" % restarts)
    if xph is not None:
        stats.append("XP/h %s" % xph)
    if kph is not None:
        stats.append("kills/h %s" % kph)
    if stats:
        lines.append(", ".join(stats))
    link = a.link or (repo_link(a.file) if a.file else "")
    if link:
        lines.append("Report: " + link)
    return "Soak summary", "\n".join(lines), link


def newest_section(text):
    parts = re.split(r"(?m)^(?=## )", text)
    secs = [p for p in parts if re.match(r"## DRILLS\d+", p)]
    if not secs:
        raise FeedError("refused: no '## DRILLS<n>' section in the drill README")
    best = max(secs, key=lambda s: int(re.match(r"## DRILLS(\d+)", s).group(1)))
    return best


def build_drill(a):
    counts = {"PASS": a.pass_, "FAIL": a.fail, "BLOCKED": a.blocked}
    findings = list(a.finding or [])
    head = ""
    if a.file and (any(v is None for v in counts.values())):
        sec = newest_section(read(a.file))
        head = re.match(r"## (DRILLS\d+(?: \([^)]*\))?)", sec).group(1)
        got = {"PASS": 0, "FAIL": 0, "BLOCKED": 0}
        rows = 0
        for ln in sec.splitlines():
            cells = [c.strip() for c in ln.strip().strip("|").split("|")]
            if not ln.lstrip().startswith("|") or len(cells) < 4 or set(cells[0]) <= set("-: "):
                continue
            v = cells[2]
            w = re.match(r"[A-Za-z]+", v)
            word = w.group(0).upper() if w else ""
            if word in got:
                got[word] += 1
                rows += 1
            elif word in ("FIXED", "OPEN", "PARTIAL"):
                rows += 1
            if re.search(r"\bopen\b|\bFAIL\b", v, re.I) and word != "PASS":
                findings.append("%s %s: %s" % (cells[0], cells[1], v))
        if not rows and all(v is None for v in counts.values()):
            raise FeedError("refused: the newest drill section has no verdict table rows (give --pass/--fail/--blocked)")
        for k in counts:
            if counts[k] is None:
                counts[k] = got[k]
    for k in counts:
        if counts[k] is None:
            raise FeedError("refused: no drill counts (give --file or --pass/--fail/--blocked)")
    link = a.link or (repo_link(a.file) if a.file else "")
    lines = ["%s: PASS %s, FAIL %s, BLOCKED %s" % (head or "Drills", counts["PASS"], counts["FAIL"], counts["BLOCKED"])]
    if findings:
        lines.append("Open findings: " + "; ".join(findings[:3]) + (" (+%d more)" % (len(findings) - 3) if len(findings) > 3 else ""))
    else:
        lines.append("Open findings: none listed")
    if link:
        lines.append("Report: " + link)
    return "Drill summary", "\n".join(lines), link


def build_upstream(a):
    m = UPSTREAM_PR_RE.fullmatch(a.url or "")
    if not m:
        raise FeedError("refused: --url must be exactly https://github.com/psi29a/d2decomped/pull/<digits>")
    if str(a.pr) != m.group(1):
        raise FeedError("refused: --url does not match --pr %s" % a.pr)
    if a.event not in EVENTS:
        raise FeedError("refused: --event must be one of " + ", ".join(EVENTS))
    if not (a.title or "").strip():
        raise FeedError("refused: --title is required")
    title = notify.trunc(" ".join(a.title.split()), 120)
    return "Upstream PR #%s %s" % (a.pr, EVENTS[a.event]), title, a.url


def build_pr_status(a):
    rows = list(a.pr or [])
    if not rows and not sys.stdin.isatty() and a.stdin_ok:
        rows = [ln for ln in sys.stdin.read().splitlines() if ln.strip()]
    if not rows:
        raise FeedError("refused: no PR rows (use --pr 'N|title|CI|reviews|state' or stdin lines)")
    out = ["PR | title | CI | CR reviews | maintainer"]
    for r in rows:
        f = [c.strip() for c in r.split("|")]
        if len(f) != 5 or not f[0].lstrip("#").isdigit():
            raise FeedError("refused: bad PR row (want N|title|CI|reviews|state): %s" % notify.trunc(r, 60))
        out.append("#%s | %s | %s | %s | %s" % (f[0].lstrip("#"), notify.trunc(f[1], 40), f[2], f[3], f[4]))
    return "Open upstream PRs", "```\n" + "\n".join(out) + "\n```", ""


def git_has_tag(tag, repo):
    try:
        r = subprocess.run(["git", "-C", repo, "rev-parse", "-q", "--verify", "refs/tags/" + tag],
                           capture_output=True, text=True, timeout=15)
    except (OSError, subprocess.SubprocessError):
        return False
    return r.returncode == 0


def build_release(a):
    if not a.tag or not re.fullmatch(r"[A-Za-z0-9._-]+", a.tag):
        raise FeedError("refused: --tag is required (letters, digits, . _ -)")
    if not git_has_tag(a.tag, a.repo or ROOT):
        raise FeedError("refused: no release exists: tag %s is not in the repo" % a.tag)
    m = RELEASE_URL_RE.fullmatch(a.url or "")
    if not m or m.group(1) != a.tag:
        raise FeedError("refused: --url must be %s/releases/tag/<tag> for this tag" % REPO_URL)
    notes = read(a.notes_file) if a.notes_file else (a.notes or "")
    notes = " ".join(notes.split())
    if not notes:
        raise FeedError("refused: no release notes (--notes or --notes-file)")
    return "Release %s" % a.tag, notify.trunc(notes, 500) + "\n" + a.url, a.url


def neutralize(text):
    """Break @mentions in the visible text (allowed_mentions is empty as well)."""
    return text.replace("@", "@​")


def check_channel(ch):
    if ch not in CHANNELS:
        raise FeedError("--channel must be one of " + ", ".join(CHANNELS))
    return ch


def send(channel, title, text, url, *, store=None, env=None, **kw):
    """One real post. Fails closed when the channel has no webhook. Returns (ok, message)."""
    try:
        hook = dc.load_webhook_url(channel, env=env, store=store, prefixes=tuple(kw.get("base_ok", dc.WEBHOOK_PREFIXES)))
    except dc.DiscordError as e:
        return False, str(e)
    if not hook:
        return False, NO_HOOK % channel
    r = notify.post(channel, title, text, url, store=store, env=env, **kw)
    return (True, "posted to #%s" % channel) if r.ok else (False, r.error)


def main(argv=None, **sendkw):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    sub = ap.add_subparsers(dest="cmd", required=True)

    def common(p):
        p.add_argument("--channel", help="default per subcommand: soak/drill/pr-status progress, release announcements, upstream d2d-upstream")
        p.add_argument("--send", action="store_true", help="post for real (default is a dry run)")
        p.add_argument("--dry-run", action="store_true", help="the default; kept for symmetry with notify.py")
        p.add_argument("--link", help="override the report link")

    s = sub.add_parser("soak")
    common(s)
    s.add_argument("--file")
    s.add_argument("--verdict")
    s.add_argument("--duration")
    s.add_argument("--deaths")
    s.add_argument("--restarts")
    s.add_argument("--xph")
    s.add_argument("--kph")
    d = sub.add_parser("drill")
    common(d)
    d.add_argument("--file")
    d.add_argument("--pass", dest="pass_", type=int)
    d.add_argument("--fail", type=int)
    d.add_argument("--blocked", type=int)
    d.add_argument("--finding", action="append")
    u = sub.add_parser("upstream")
    common(u)
    u.add_argument("--pr", required=True)
    u.add_argument("--event", required=True)
    u.add_argument("--title", default="")
    u.add_argument("--url", default="")
    p = sub.add_parser("pr-status")
    common(p)
    p.add_argument("--pr", action="append")
    r = sub.add_parser("release")
    common(r)
    r.add_argument("--tag")
    r.add_argument("--notes")
    r.add_argument("--notes-file")
    r.add_argument("--url", default="")
    r.add_argument("--repo", help="git checkout to look for the tag (default: this repo)")
    a = ap.parse_args(argv)
    a.stdin_ok = True
    try:
        channel = check_channel(a.channel or DEFAULT_CHANNEL[a.cmd])
        builders = {"soak": build_soak, "drill": build_drill, "upstream": build_upstream,
                    "pr-status": build_pr_status, "release": build_release}
        title, text, url = builders[a.cmd](a)
        hits = notify.check_text(title, text, "")
        if hits:
            raise FeedError("refused: text looks private (%s); nothing was sent" % ", ".join(hits))
        title, text = neutralize(title), neutralize(text)
    except FeedError as e:
        print(e, file=sys.stderr)
        return 2
    if not a.send:
        print("[dry run] channel #%s\n%s\n%s" % (channel, title, text) + ("\nLink: " + url if url and url not in text else ""))
        return 0
    ok, msg = send(channel, title, text, url, **sendkw)
    print(msg, file=sys.stdout if ok else sys.stderr)
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
