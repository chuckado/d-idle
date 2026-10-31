## Why

The d.idle Discord server was built on 2026-10-31 and its live state differs slightly from the drafts (hidden STAFF category, news channel, AutoMod, feeds, bot keeps Administrator). no observable behaviour change: docs and a layout file only; the apply tool gains an optional channel kind.

## What Changes

- docs/community/discord plan, server, roles, feeds: live state, remaining user steps, the accepted-risk bot decision.
- server.json: STAFF category with hidden `discord-updates`, `announcements` of type `news`; `apply_server.py` accepts `"type": "news"` and recognises an existing news channel.
