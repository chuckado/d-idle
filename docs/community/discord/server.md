# d.idle Discord: server layout

> status: DRAFT. The same layout as data is [server.json](server.json); `tools/discord/apply_server.py` can create it. Doing it by hand
> in the Discord UI is equally fine and takes about 20 minutes; the tool is only a convenience.

## Channels
Topic text is the Discord channel topic (each at most 100 characters). Read-only means members cannot post; the owner and webhooks can.

| category | channel | topic | slowmode (s) | posting |
|---|---|---|---|---|
| WELCOME | `#welcome` | What d.idle is and where to start. Read-only. | off | read-only |
| WELCOME | `#rules` | Server rules. Read before posting. | off | read-only |
| WELCOME | `#announcements` | Releases and project news. Read-only. | off | read-only |
| PROJECT | `#progress` | Automated feed: releases, merged PRs, CI, soak summaries. Read-only. | off | read-only |
| PROJECT | `#dev` | Code, specs, design and contributor chat. Tag claims verified or unverified. | off | open |
| PROJECT | `#d2d-upstream` | Our work around d2d, the open-source engine we run on. Be kind to its maintainer. | 10 | open |
| HELP | `#help` | Running the bot yourself? Ask here. Pinned template, scrub paths and names. | 10 | open |
| HELP | `#bug-reports` | One bug per thread. Use the pinned template. | 60 | open |
| COMMUNITY | `#general` | Chat about d.idle and idle or bot-played Diablo II. Offline single-player only. | 5 | open |
| COMMUNITY | `#show-your-runs` | Share a run: class, build, hours, deaths, how measured. | 60 | open |
| STREAM | `#stream` | Stream notices and chat about the d.idle stream. | 5 | open |

Descriptions: `#welcome` start page; `#rules` the rules; `#announcements` releases and news (after enabling Community, convert it to an
Announcement channel by hand); `#progress` the automated feed ([feeds.md](feeds.md)); `#dev` contributor talk; `#d2d-upstream` our
work around the engine we run on; `#help` and `#bug-reports` support with pinned templates; `#general` chat; `#show-your-runs` shared
runs; `#stream` stream notices and chat (the stream is not live yet).

## Roles
- `owner`: Administrator, hoisted. The user. Keep 2FA on.
- `contributor`: people with merged work or real review help; hoisted; may create and manage threads. Assigned by hand.
- `member`: the trusted tier, assigned by hand to people the moderator has seen behave; may create public threads and use external emoji.
- `mod` (OPTIONAL, disabled by default: listed under `optional_roles` in server.json, not created by the tool): helper with Manage Messages, Timeout, Kick, View Audit Log; below `owner`. See [roles.md](roles.md).
- `@everyone`: view, read history, send, embed links, react, send in threads; attach files only in `#show-your-runs` and `#bug-reports`. No mentions of everyone, no thread creation, no Create Invite.
- Bots and webhooks: only webhooks at the start (GitHub feed, `notify.py`). They post regardless of channel permissions. No role needed.
Raid switch: remove send permissions from `@everyone` so only `member` and above can talk (`member` already holds them). Full guideline, permission table and runbook: [roles.md](roles.md).

## Permission matrix
| channel | @everyone | member | contributor | owner |
|---|---|---|---|---|
| `#welcome`, `#rules`, `#announcements` | read | read | read | post |
| `#progress` | read | read | read | post (feed via webhooks) |
| `#dev`, `#d2d-upstream`, `#help`, `#bug-reports`, `#general`, `#show-your-runs`, `#stream` | read, post | read, post | read, post | post |

Implemented as one overwrite per read-only channel: `@everyone` denied Send Messages, Send in Threads and Create Public Threads; `#show-your-runs` and `#bug-reports` additionally allow Attach Files for `@everyone`.

## Server settings
- Verification level: Medium (registered over 5 minutes); High if arrivals spike. Explicit media content filter: scan all members.
- Default notification setting: only mentions. Enable Community (needed for Rules Screening, Announcement channels, Onboarding).
- Rules Screening: on, with the text of `rules.md`. Onboarding: default channels `#welcome`, `#rules`, `#help`, `#general`.
- AutoMod: block mass mentions and invite links; flag new-account links. Safety alerts channel: none at first (use the audit log).
- Invites: no permanent unlimited invite. Public invite: no expiry but a member cap (start 100), temporary membership off; revoke
  and recreate to rotate. Only the owner has Create Invite.
- Slowmode: values in the table; raise them during a raid.

## One-time bot setup for apply_server.py (optional)
1. Developer portal: New Application, Bot tab, Reset Token; store it in the keyring (paste at the prompt, never in chat):
   `secret-tool store --label="d.idle Discord bot token" service d2idle account discord-bot-token`
2. Invite it to the EMPTY server with only Manage Channels and Manage Roles (OAuth2 URL generator, scope `bot`; add Manage Server
   only if you use `--settings`). Do not enable privileged intents.
3. Get the server id (Developer Mode, right-click the server, Copy Server ID), then dry run, read, apply:
   `python3 tools/discord/apply_server.py --guild <id>`  then `... --apply` (and `--settings` once, to set verification, the filter and the @everyone permissions)
4. Afterwards kick the bot or strip its permissions, and Reset Token in the portal; `secret-tool clear service d2idle account discord-bot-token`.
It matches roles and channels by name and only creates missing ones; it never deletes or renames; existing channels are left as is.
New channels are created at the bottom. The `owner` role must be assigned to the user by hand. Do the Community, Rules Screening,
Onboarding and invite steps in the UI: the tool does not touch them.

## server.json schema (schema 1)
```
{ "schema": 1, "name": "d.idle",
  "settings": { "verification_level": 0-4, "explicit_content_filter": 0-2, "default_message_notifications": 0-1,
                "everyone_permissions": [permission names] },          (applied only with --settings)
  "optional_roles": [ same shape as roles, plus "enabled": false and "note" ]   (documentation only, ignored by the tool)
  "roles": [ { "name", "color": "#RRGGBB", "hoist": bool, "mentionable": bool, "permissions": [names] } ],
  "categories": [ { "name", "channels": [ { "name", "type": "text", "topic", "slowmode": seconds,
        "overwrites": [ { "role": "@everyone" | role name, "allow": [names], "deny": [names] } ] } ] } ] }
```
Permission names: administrator, view_channel, read_message_history, send_messages, send_messages_in_threads, embed_links, attach_files,
add_reactions, create_public_threads, manage_threads, manage_messages, mention_everyone, use_external_emojis, manage_channels,
manage_roles, create_instant_invite. Verification 2 is Medium; filter 2 is all members; notifications 1 is only mentions.
