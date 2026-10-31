## Why

The user wants a project Discord named d.idle for contributors, followers and people who run the bot, as a separate small server that links to the d2d maintainer's community. This card writes the drafts and the tooling only: layout, rules, welcome and FAQ text, feed instructions, a webhook notifier and an optional server-setup script. Nothing is created, posted or linked; every outward step stays with the user.

## What Changes

- docs/community/discord/: plan, server layout (+ server.json), welcome, rules, faq, templates, maintainer message draft, feeds, public-page proposal.
- tools/discord/: notify.py (webhook posts with a fail-closed sanitizer, keyring or env secrets, mock-tested), apply_server.py (dry-run by default, idempotent, mock-tested).
- docs/INDEX.md and CHANGELOG.md lines. README, ADRs, GLOSSARY and decisions are not touched; the README line is proposed in plan.md.

## Dependencies

None.
