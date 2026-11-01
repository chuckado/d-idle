## Why

The Discord feeds had only the generic `notify.py`; soak and drill summaries and upstream PR news were typed by hand. The user said 'wire the feeds' (2026-11-01).

## What Changes

- tools/discord/feed.py: subcommands soak, drill, upstream, pr-status, release; dry run by default, `--send` posts; reuses notify.py and common.py (sanitizer, webhook loader, rate limit); fail-closed for a channel without a webhook file; no secret is created or printed.
- tools/discord/tests/test_feed.py: offline tests (mock server, fake home).
- docs/community/discord/feeds.md and server.md: subcommands, channel mapping, what not to post, the 'Not wired yet' list; docs/process/feed-after-events.md; INDEX, CHANGELOG, BACKLOG lines.
- Not done (needs the user go): CI-failure-only feed (Actions secret and workflow edit), `#d2d-upstream` webhook.
- No numeric gate changes (no delta).

Overlap (declared): docs/INDEX.md, CHANGELOG.md and docs/BACKLOG.md are standing `main` docs, one added line each; no other open card owns the files above.
