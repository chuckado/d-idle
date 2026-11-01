## Why

The user asked (2026-11-01) to post the dispatcher's 15-minute tick report to the d.idle Discord progress feed as well.

## What Changes

- `tools/discord/feed.py tick`: one short `#progress` post (running, merged, launched, upstream, blocked, waiting), through the existing sanitizer, fail-closed webhook loader and dry-run default.
- An idle tick (nothing merged, launched, blocked, upstream or waiting) is refused with exit 2.

## Impact

- Owns tools/discord/** and docs/process/feed-after-events.md, docs/community/discord/feeds.md. Nothing is posted by this card.
