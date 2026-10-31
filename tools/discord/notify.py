#!/usr/bin/env python3
"""Post a short message to a d.idle Discord channel through a webhook. Stdlib only.

  notify.py init                                   print the one-time setup steps
  notify.py post --channel progress --title T --text TEXT [--url U] [--dry-run]

The webhook URL comes from the env var D2IDLE_DISCORD_WEBHOOK_<CHANNEL> or the desktop keyring
(service d2idle, account discord-webhook-<channel>). It is never printed, logged or put in an error.
Text that looks private (home paths, lab hostnames, IPs, tokens) is refused before anything is sent.
post() never raises: it returns a Result.
"""
import argparse
import json
import os
import re
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import common as dc
from common import DiscordError, redact

LIMITS = {"title": 256, "description": 4096, "content": 2000}
MAX_RETRIES = 3
MAX_RETRY_AFTER = 30.0
COLOR = 0x8B1A1A
WEBHOOK_PREFIXES = ("https://discord.com/api/webhooks/", "https://discordapp.com/api/webhooks/")
CHANNEL_RE = re.compile(r"[a-z][a-z0-9-]{0,31}")


@dataclass
class Result:
    ok: bool
    error: str = ""
    refused: bool = False
    status: int = 0
    attempts: int = 0


def account_for(channel):
    return "discord-webhook-" + channel


def env_name(channel):
    return "D2IDLE_DISCORD_WEBHOOK_" + channel.upper().replace("-", "_")


def get_webhook(channel, store=None, env=None):
    env = os.environ if env is None else env
    url = env.get(env_name(channel)) or (store or dc.SecretToolStore()).get(account_for(channel))
    if url:
        dc.register_secret(url)
        dc.register_secret(urllib.parse.urlsplit(url).path.rstrip("/").rsplit("/", 1)[-1])
    return url or None


def trunc(text, limit):
    text = str(text or "")
    return text if len(text) <= limit else text[:limit - 3].rstrip() + "..."


def build_payload(title="", text="", url=""):
    """Embed payload; mentions are disabled so a post can never ping anyone."""
    emb = {"color": COLOR}
    if title:
        emb["title"] = trunc(title, LIMITS["title"])
    if text:
        emb["description"] = trunc(text, LIMITS["description"])
    if url:
        emb["url"] = url
    payload = {"username": "d.idle", "allowed_mentions": {"parse": []}}
    if len(emb) > 1:
        payload["embeds"] = [emb]
    else:
        payload["content"] = trunc("d.idle", LIMITS["content"])
    return payload


def check_text(title, text, url):
    """Sanitizer verdict for everything that would be sent: list of rule names (empty = fine)."""
    hits = dc.scan(" \n".join([title or "", text or ""]))
    if url:
        if not re.fullmatch(r"https://[^\s]+", url):
            hits.append("url-not-https")
        hits += dc.scan(url)
    return sorted(set(hits))


def _sleep(s):
    time.sleep(s)


def post(channel, title="", text="", url="", *, store=None, env=None, sleep=_sleep, base_ok=WEBHOOK_PREFIXES, timeout=10):
    """Send one embed. Never raises. `base_ok` lists accepted webhook URL prefixes (tests pass a local one)."""
    try:
        if not CHANNEL_RE.fullmatch(channel or ""):
            return Result(False, "bad channel name", refused=True)
        if not (title or text):
            return Result(False, "nothing to post (need --title or --text)", refused=True)
        hits = check_text(title, text, url)
        if hits:
            return Result(False, "refused: text looks private (%s); nothing was sent" % ", ".join(hits), refused=True)
        hook = get_webhook(channel, store, env)
        if not hook:
            return Result(False, "no webhook for channel %s: run `notify.py init`" % channel)
        if not hook.startswith(tuple(base_ok)):
            return Result(False, "stored webhook is not a Discord webhook URL")
        body = json.dumps(build_payload(title, text, url)).encode()
        last = ""
        for attempt in range(1, MAX_RETRIES + 1):
            req = urllib.request.Request(hook, data=body, method="POST",
                                         headers={"Content-Type": "application/json", "User-Agent": "d-idle-notify/1"})
            try:
                with urllib.request.urlopen(req, timeout=timeout) as r:
                    return Result(True, status=r.status, attempts=attempt)
            except urllib.error.HTTPError as e:
                code = e.code
                if code == 429:
                    try:
                        wait = float(json.loads(e.read() or b"{}").get("retry_after", 1))
                    except (ValueError, TypeError, AttributeError):
                        wait = 1.0
                    last = "rate limited (429)"
                    if attempt < MAX_RETRIES:
                        sleep(min(max(wait, 0.0), MAX_RETRY_AFTER))
                    continue
                if code >= 500:
                    last = "discord error %d" % code
                    if attempt < MAX_RETRIES:
                        sleep(min(2.0 * attempt, MAX_RETRY_AFTER))
                    continue
                return Result(False, "discord rejected the post (%d)" % code, status=code, attempts=attempt)
            except (urllib.error.URLError, OSError, ValueError) as e:
                last = "network error: %s" % redact(getattr(e, "reason", e))
                if attempt < MAX_RETRIES:
                    sleep(1.0)
        return Result(False, "gave up after %d tries: %s" % (MAX_RETRIES, last), attempts=MAX_RETRIES)
    except Exception as e:  # post() must never raise into a soak or release script
        return Result(False, "unexpected error: %s" % redact(e))


INIT_TEXT = """One-time setup (never paste a webhook URL into chat, an issue or the repo):

1. In Discord: channel settings -> Integrations -> Webhooks -> New Webhook, pick the channel, Copy Webhook URL.
2. Store it in the desktop keyring (paste at the prompt; the value is not echoed to a file):
     secret-tool store --label="d.idle Discord {ch} webhook" service d2idle account discord-webhook-{ch}
   Or, for one shell only:  export {env}=...   (do not put it in a file in the repo)
3. Dry run (prints the payload, sends nothing):
     python3 tools/discord/notify.py post --channel {ch} --title "test" --text "hello" --dry-run
4. Real test, once the server is private-ready:
     python3 tools/discord/notify.py post --channel {ch} --title "test" --text "hello"
Rotate: delete the webhook in Discord, create a new one, store it again (secret-tool store overwrites).
Remove: secret-tool clear service d2idle account discord-webhook-{ch}
"""


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    sub = ap.add_subparsers(dest="cmd", required=True)
    pi = sub.add_parser("init", help="print the one-time setup steps")
    pi.add_argument("--channel", default="progress")
    pp = sub.add_parser("post", help="post an embed")
    pp.add_argument("--channel", required=True)
    pp.add_argument("--title", default="")
    pp.add_argument("--text", default="", help="message body; '-' reads stdin")
    pp.add_argument("--url", default="")
    pp.add_argument("--dry-run", action="store_true")
    a = ap.parse_args(argv)
    if a.cmd == "init":
        if not CHANNEL_RE.fullmatch(a.channel):
            print("bad channel name", file=sys.stderr)
            return 2
        print(INIT_TEXT.format(ch=a.channel, env=env_name(a.channel)))
        return 0
    text = sys.stdin.read() if a.text == "-" else a.text
    if a.dry_run:
        hits = check_text(a.title, text, a.url)
        if hits:
            print("refused: text looks private (%s)" % ", ".join(hits), file=sys.stderr)
            return 2
        print(json.dumps(build_payload(a.title, text, a.url), indent=2))
        return 0
    r = post(a.channel, a.title, text, a.url)
    if r.ok:
        print("posted to #%s" % a.channel)
        return 0
    print(r.error, file=sys.stderr)
    return 2 if r.refused else 1


if __name__ == "__main__":
    sys.exit(main())
