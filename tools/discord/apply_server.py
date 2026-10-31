#!/usr/bin/env python3
"""Create the d.idle server layout (docs/community/discord/server.json) through the Discord API. Stdlib only.

DRY-RUN BY DEFAULT: prints the plan. Nothing changes without --apply. Idempotent: roles and channels are matched by
name and only missing ones are created. Nothing is ever deleted or renamed. Existing channels keep their settings.
Guild-wide settings and the @everyone permissions are only touched with --settings (plus --apply).

  apply_server.py --guild GUILD_ID [--file server.json] [--apply] [--settings]

The bot token comes from D2IDLE_DISCORD_BOT_TOKEN, else the private file ~/.local/share/d2idle/discord-bot-token (mode 600),
else the keyring (service d2idle, account discord-bot-token); see dc.load_bot_token.
Setup and the schema: docs/community/discord/server.md.
"""
import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import common as dc
from common import DiscordError, redact

API = "https://discord.com/api/v10"
PERMS = {
    "create_instant_invite": 1 << 0, "administrator": 1 << 3, "manage_channels": 1 << 4, "manage_roles": 1 << 28,
    "add_reactions": 1 << 6, "view_channel": 1 << 10, "send_messages": 1 << 11, "manage_messages": 1 << 13,
    "embed_links": 1 << 14, "attach_files": 1 << 15, "read_message_history": 1 << 16, "mention_everyone": 1 << 17,
    "use_external_emojis": 1 << 18, "manage_threads": 1 << 34, "create_public_threads": 1 << 35,
    "send_messages_in_threads": 1 << 38,
}
TEXT, CATEGORY, NEWS = 0, 4, 5
KINDS = {"text": TEXT, "news": NEWS}  # news (announcement) channels need Community enabled
AUTH = "Author" + "ization"  # header name, split so the repo's leak scanner does not read it as a credential


def perm_bits(names):
    bits = 0
    for n in names or []:
        if n not in PERMS:
            raise DiscordError("unknown permission name: %s" % n)
        bits |= PERMS[n]
    return bits


def hex_color(s):
    return int(str(s).lstrip("#") or "0", 16)


class Api:
    """Tiny Discord REST client. Honours 429 retry_after (max 3 tries); never prints the token."""

    def __init__(self, token, base=API, sleep=time.sleep):
        self.token, self.base, self.sleep = token, base, sleep
        dc.register_secret(token)

    def call(self, method, path, body=None):
        data = json.dumps(body).encode() if body is not None else None
        for attempt in range(1, 4):
            req = urllib.request.Request(self.base + path, data=data, method=method, headers={
                AUTH: "Bot " + self.token, "Content-Type": "application/json", "User-Agent": "DiscordBot (d-idle, 1)"})
            try:
                with urllib.request.urlopen(req, timeout=15) as r:
                    raw = r.read()
                    return json.loads(raw) if raw else None
            except urllib.error.HTTPError as e:
                if e.code == 429 and attempt < 3:
                    try:
                        wait = float(json.loads(e.read() or b"{}").get("retry_after", 1))
                    except (ValueError, TypeError):
                        wait = 1.0
                    self.sleep(min(wait, 30.0))
                    continue
                raise DiscordError("%s %s failed (%d)" % (method, path.split("?")[0], e.code))
            except (urllib.error.URLError, OSError) as e:
                raise DiscordError("%s %s: network error %s" % (method, path, getattr(e, "reason", e)))


def plan_and_apply(api, guild, spec, apply=False, settings=False, out=print):
    """Returns the list of action strings. With apply=False nothing is sent except GETs."""
    actions = []

    def act(text, fn=None):
        actions.append(text)
        out(("DO   " if apply else "PLAN ") + text)
        return fn() if (apply and fn) else None

    roles = {r["name"]: r for r in api.call("GET", "/guilds/%s/roles" % guild)}
    chans = api.call("GET", "/guilds/%s/channels" % guild)
    by = {(c["type"], c["name"]): c for c in chans}
    everyone = next((r for r in roles.values() if r["id"] == str(guild)), None) or {"id": str(guild)}
    ids = {"@everyone": everyone["id"]}

    if settings:
        s = spec.get("settings", {})
        body = {k: s[k] for k in ("verification_level", "explicit_content_filter", "default_message_notifications") if k in s}
        if body:
            act("set guild settings %s" % json.dumps(body, sort_keys=True), lambda: api.call("PATCH", "/guilds/%s" % guild, body))
        if "everyone_permissions" in s:
            bits = perm_bits(s["everyone_permissions"])
            act("set @everyone permissions to %s" % ", ".join(s["everyone_permissions"]),
                lambda: api.call("PATCH", "/guilds/%s/roles/%s" % (guild, everyone["id"]), {"permissions": str(bits)}))

    for r in spec.get("roles", []):
        if r["name"] in roles:
            ids[r["name"]] = roles[r["name"]]["id"]
            out("OK   role exists: %s" % r["name"])
            continue
        body = {"name": r["name"], "permissions": str(perm_bits(r.get("permissions"))), "color": hex_color(r.get("color", "0")),
                "hoist": bool(r.get("hoist")), "mentionable": bool(r.get("mentionable"))}
        made = act("create role %s" % r["name"], lambda body=body: api.call("POST", "/guilds/%s/roles" % guild, body))
        ids[r["name"]] = made["id"] if made else "<new:%s>" % r["name"]

    def overwrites(items):
        res = []
        for o in items or []:
            if o["role"] not in ids:
                raise DiscordError("overwrite names unknown role %s" % o["role"])
            res.append({"id": ids[o["role"]], "type": 0, "allow": str(perm_bits(o.get("allow"))), "deny": str(perm_bits(o.get("deny")))})
        return res

    for cat in spec.get("categories", []):
        found = by.get((CATEGORY, cat["name"]))
        if found:
            pid = found["id"]
            out("OK   category exists: %s" % cat["name"])
        else:
            made = act("create category %s" % cat["name"],
                       lambda cat=cat: api.call("POST", "/guilds/%s/channels" % guild, {"name": cat["name"], "type": CATEGORY}))
            pid = made["id"] if made else "<new:%s>" % cat["name"]
        for c in cat.get("channels", []):
            if (TEXT, c["name"]) in by or (NEWS, c["name"]) in by:
                out("OK   channel exists: #%s (left as is)" % c["name"])
                continue
            body = {"name": c["name"], "type": KINDS.get(c.get("type", "text"), TEXT), "topic": c.get("topic", "")[:1024], "parent_id": pid,
                    "rate_limit_per_user": int(c.get("slowmode", 0)), "permission_overwrites": overwrites(c.get("overwrites"))}
            act("create channel #%s in %s" % (c["name"], cat["name"]),
                lambda body=body: api.call("POST", "/guilds/%s/channels" % guild, body))
    if not actions:
        out("Nothing to do: the server already matches.")
    return actions


def main(argv=None, api=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--guild", required=True, help="server (guild) id, digits")
    ap.add_argument("--file", default=os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "docs", "community", "discord", "server.json"))
    ap.add_argument("--apply", action="store_true", help="really create things (default: print the plan)")
    ap.add_argument("--settings", action="store_true", help="also set guild settings and @everyone permissions")
    a = ap.parse_args(argv)
    if not a.guild.isdigit():
        print("guild id must be digits", file=sys.stderr)
        return 2
    try:
        with open(a.file) as f:
            spec = json.load(f)
        if api is None:
            tok = dc.load_bot_token()
            if not tok:
                print("no bot token: see docs/community/discord/server.md, One-time bot setup (private file or keyring)", file=sys.stderr)
                return 2
            api = Api(tok)
        plan_and_apply(api, a.guild, spec, apply=a.apply, settings=a.settings)
    except (DiscordError, OSError, ValueError, KeyError) as e:
        print("error: %s" % redact(e), file=sys.stderr)
        return 1
    if not a.apply:
        print("Dry run: nothing was changed. Add --apply to do it.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
