## Why

The user asked for a sourced guideline for roles in a public Discord server, keeping the recommendation that `@everyone` gets chat and `member` is a hand-assigned trusted tier. This card writes the guideline and aligns the Discord drafts with it. Docs only; nothing is created on Discord.

## What Changes

- docs/community/discord/roles.md (new).
- Minimal edits to server.md, server.json and plan.md where they conflicted (optional `mod`, attach files, member permissions).
- INDEX and CHANGELOG lines.

## Dependencies

Builds on agent/discord (card `discord`, commit 4a4a0d6): the discord card still owns these paths until it is archived; this card claims roles.md and takes follow-up edits to the other three files only. It releases nothing of the discord card's; merge discord first.
