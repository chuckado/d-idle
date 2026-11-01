#!/usr/bin/env python3
"""Post the dispatcher tick to #progress, with the field text written by Haiku. Stdlib only.

  python3 tools/discord/tick_post.py [--send] [--no-api] [--channel progress] < facts.json

stdin: JSON with the keys running, merged, launched, upstream, blocked, waiting (short strings or lists).
Dry run is the default: the text and its size are printed, nothing is posted. --send posts ONE embed through
feed.send. The text goes through the same checks as `feed.py tick` (idle refusal, 3800 cap, notify sanitizer incl. the
privacy guard). If the guard would refuse the Haiku text, Haiku is asked ONCE to reword the offending phrase; if that
fails or any API step fails (no key, HTTP error, timeout, text that drops a fact) the plain deterministic rendering of
the facts is used instead and stderr says so. It never raises and never prints the key (haiku_tick.read_key is the only
reader; the key only goes into the x-api-key header). Thinking is disabled for the call (it used the token budget). Needs tools/ci/haiku_tick.py and tools/ci/anthropic_http.py.
"""
import argparse
import json
import os
import re
import sys
import types

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
sys.path.insert(0, os.path.join(os.path.dirname(HERE), "ci"))
import anthropic_http  # noqa: E402
import common as dc  # noqa: E402
import feed  # noqa: E402
import haiku_tick  # noqa: E402
import notify  # noqa: E402

FIELDS = feed.TICK_FIELDS
LISTY = ("running", "merged", "launched")
RULES_TEXT = (" Keep every fact, every number and every name exactly as given, and add nothing. Never write host names, file paths, emails, tokens or addresses; "
              "say 'the worker host' instead. Avoid any run of five or more short words (4 to 6 letters each) in a "
              "row: break such runs with a longer word or punctuation. Each field is one line.")


def facts_to_vals(facts):
    vals = {}
    for k in FIELDS:
        v = facts.get(k)
        if isinstance(v, (list, tuple)):
            v = ", ".join(str(i) for i in v if str(i).strip()) if k in LISTY else "; ".join(str(i) for i in v if str(i).strip())
        vals[k] = feed.tick_clean(str(v) if v is not None else "")
    return vals


def render(vals, now=None):
    """Deterministic rendering through feed.build_tick: (title, text). Raises feed.FeedError when idle."""
    ns = types.SimpleNamespace(file=None, now=now, **vals)
    title, text, _ = feed.build_tick(ns)
    return title, text


def parse_haiku(text):
    """'<field>: text' lines to vals; None when a line is not in that form (fall back)."""
    vals = {k: "" for k in FIELDS}
    for ln in text.splitlines():
        if not ln.strip():
            continue
        m = re.match(r"^\W*([A-Za-z ]+?)\s*:\s*(.*)$", ln)
        key = m and next((k for k in FIELDS if m.group(1).strip().lower().startswith(k)), None)
        if not key:
            return None
        vals[key] = (vals[key] + "; " + m.group(2)).strip("; ")
    return vals


NUM = re.compile(r"\d+(?:\.\d+)?")
WORD = re.compile(r"[A-Za-z0-9][A-Za-z0-9_./-]*")


def idlike(w):
    """Salient outside the list fields: has a digit, '_' or '-', or is capitalised (PR, Navi)."""
    return bool(re.search(r"[\d_-]", w)) or w[0].isupper()


def salient(vals):
    """Lowercase tokens that must survive rewording: numbers, every word of list items (card names), ids."""
    toks = set()
    for k in FIELDS:
        v = vals[k]
        toks.update(NUM.findall(v))
        for w in WORD.findall(v):
            w = w.strip("./-")
            if w and (k in LISTY or idlike(w)):
                toks.add(w.lower())
    return toks


def fact_problem(src, got):
    """None when the text keeps every fact and invents none; else a short reason. Set based, so rewording passes."""
    for k in FIELDS:
        if bool(src[k]) != bool(got[k]):
            return "field %s %s" % (k, "dropped" if src[k] else "invented")
    have = set(w.strip("./-").lower() for k in FIELDS for w in WORD.findall(got[k])) | \
        set(n for k in FIELDS for n in NUM.findall(got[k]))
    missing = sorted(salient(src) - have)
    if missing:
        return "dropped %s" % ", ".join(missing[:6])
    extra = sorted(set(n for k in FIELDS for n in NUM.findall(got[k])) - set(n for k in FIELDS for n in NUM.findall(src[k])))
    if extra:
        return "invented number %s" % ", ".join(extra[:3])
    return None


NOTHING = re.compile(r"(?i)^\W*(?:nothing|none|n/a|no change|unchanged|no one)\b")


def drop_nothing(src, got):
    """A field the facts left empty that the model filled with a 'Nothing ...' sentence is no invented fact: empty it."""
    for k in FIELDS:
        if not src[k] and got[k] and NOTHING.match(got[k]):
            got[k] = ""


def facts_kept(src, got):
    return fact_problem(src, got) is None


def offending(text):
    """Matched phrases of the sanitizer rules (the part to reword) and the rule names."""
    out = []
    for name, rx in dc.RULES:
        for m in rx.finditer(text):
            out.append(m.group(0))
    return out, notify.check_text("", text, "")


CDKEY = dict(dc.RULES)["cdkey"]  # the guard's own pattern: five 4-6 char tokens joined by one space or hyphen
REPAIR_TRIES = 5


def repair_runs(text, tries=REPAIR_TRIES):
    """Deterministic guard repair on the TEXT: returns (text, runs_found). Each cdkey-pattern run found by the guard's
    own regex gets its separator after the 3rd word (a space when it has one) replaced by ', ', which the pattern
    cannot join across. Only a separator changes: no word, number, id or name is touched. Gives up after `tries`."""
    runs = []
    for _ in range(tries):
        m = CDKEY.search(text)
        if not m:
            break
        runs.append(m.group(0))
        seps = list(re.finditer(r"[- ]", m.group(0)))
        spaces = [x for x in seps if x.group() == " "]
        pick = (seps[2] if seps[2].group() == " " else spaces[min(2, len(spaces) - 1)]) if spaces else seps[2]
        at = m.start() + pick.start()
        if pick.group() == " ":
            text = text[:at] + "," + text[at:]
        else:
            text = text[:at] + ", " + text[at + 1:]
    return text, runs


def guard_fix(title, text, log):
    """(text, hits): the text unchanged when the guard passes; else the separator repair when that makes it pass."""
    hits = notify.check_text(title, text, "")
    if not hits:
        return text, hits
    if "cdkey" in hits:
        fixed, runs = repair_runs(text)
        if runs:
            log("tick_post: guard run found: %s" % "; ".join(repr(r[:80]) for r in runs))
        if fixed != text:
            after = notify.check_text(title, fixed, "")
            if not after:
                log("tick_post: repaired %d separator(s); the guard passes now" % len(runs))
                return fixed, after
    return text, hits


def usage_http(http, acc):
    """Wrap an HTTP layer so the API's usage block is summed into acc."""
    def wrapped(url, headers, body, *a, **kw):
        status, data = http(url, headers, body, *a, **kw)
        u = (data or {}).get("usage") if isinstance(data, dict) else None
        if isinstance(u, dict):
            acc["input"] += int(u.get("input_tokens") or 0)
            acc["output"] += int(u.get("output_tokens") or 0)
            acc["calls"] += 1
        return status, data
    return wrapped


def call(key, system, user, http, acc):
    """One Haiku call, thinking off; a cut-off answer (stop_reason max_tokens) is an error, not text to parse."""
    meta = {}
    text = anthropic_http.message(key, haiku_tick.MODEL, system, user, haiku_tick.MAX_TOKENS,
                                  http=usage_http(http, acc), thinking=haiku_tick.THINKING, meta=meta)
    acc["stop"] = meta.get("stop_reason")
    if meta.get("stop_reason") == "max_tokens":
        raise RuntimeError("haiku answer truncated (stop_reason max_tokens, output_tokens=%s)"
                           % (meta.get("usage") or {}).get("output_tokens"))
    return text


def haiku_text(vals, key, http, acc):
    facts = {k: v for k, v in vals.items() if v}
    user = haiku_tick.build_user(facts)
    return call(key, haiku_tick.SYSTEM + RULES_TEXT, user, http, acc)


def reword(text, phrases, key, http, acc):
    system = ("You fix one problem in a status text. Reword only the listed phrases so they no longer match; keep every "
              "other word and all facts. Output the full corrected '<field>: <text>' lines only." + RULES_TEXT)
    user = "Phrases to reword: %s\n\nText:\n%s" % (json.dumps(phrases), text)
    return call(key, system, user, http, acc)


def compose(vals, no_api, http, key_reader, acc, log, now=None):
    """Returns (title, text, source). Raises feed.FeedError for an idle tick or text no rendering can post."""
    det = render(vals, now)  # idle tick refuses here, before any API call
    det = (det[0], guard_fix(det[0], det[1], log)[0])  # plain rendering: separator repair when the guard would refuse
    if no_api:
        return det + ("deterministic",)
    try:
        key = key_reader()
        raw = haiku_text(vals, key, http, acc)
        for attempt in (0, 1):
            got = parse_haiku(raw)
            if got is None:
                bad = next((ln for ln in raw.splitlines() if ln.strip() and not parse_haiku(ln)), "")
                raise RuntimeError("not in the field format (stop_reason %s; first bad line: %r)" % (acc.get("stop"), bad[:80]))
            drop_nothing(vals, got)
            why = fact_problem(vals, got)
            if why:
                raise RuntimeError("fact check failed: %s (stop_reason %s)" % (why, acc.get("stop")))
            title, text = render(got, now)
            phrases, hits = offending(text)
            if not hits:
                return title, text, "haiku"
            if attempt:
                fixed, after = guard_fix(title, text, log)
                if not after:
                    return title, fixed, "haiku"
                raise RuntimeError("guard still trips after one reword and the separator repair (%s)" % ", ".join(after))
            log("tick_post: privacy guard would refuse the Haiku text (%s); one reword" % ", ".join(hits))
            raw = reword(raw, sorted(set(phrases)), key, http, acc)
    except Exception as e:  # never block the tick; only our own key-free messages are shown, else the type name
        why = str(e) if isinstance(e, RuntimeError) else type(e).__name__
        log("tick_post: API path failed (%s); using the plain rendering" % why)
    return det + ("deterministic",)


def main(argv=None, stdin=None, http=anthropic_http.post_json, sender=None, key_reader=haiku_tick.read_key, now=None,
          out=None, err=None):
    out, err = out or sys.stdout, err or sys.stderr
    log = lambda m: print(m, file=err)
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--send", action="store_true", help="post to Discord for real (default: dry run)")
    ap.add_argument("--no-api", action="store_true", help="deterministic rendering only, no API call")
    ap.add_argument("--channel", default="progress")
    a = ap.parse_args(argv)
    try:
        facts = json.load(stdin or sys.stdin)
        if not isinstance(facts, dict):
            raise ValueError("facts must be a JSON object")
        feed.check_channel(a.channel)
        acc = {"input": 0, "output": 0, "calls": 0}
        title, text, src = compose(facts_to_vals(facts), a.no_api, http, key_reader, acc, log, now)
        hits = notify.check_text(title, text, "")
        if hits:
            raise feed.FeedError("refused: text looks private (%s); nothing was sent" % ", ".join(hits))
    except (feed.FeedError, ValueError) as e:
        log(str(e))
        return 2
    title, text = feed.neutralize(title), feed.neutralize(text)
    print("[%s] source=%s chars=%d api_calls=%d input_tokens=%d output_tokens=%d"
          % ("send" if a.send else "dry run", src, len(text), acc["calls"], acc["input"], acc["output"]), file=out)
    if not a.send:
        print("channel #%s\n%s\n%s" % (a.channel, title, text), file=out)
        return 0
    ok, msg = (sender or feed.send)(a.channel, title, text, "")
    print(msg, file=out if ok else err)
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
