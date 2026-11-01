## Why

The user asked (2026-11-01) to post the entire dispatcher tick report, not the truncated one. `tick` capped the post at 900 chars (and each field at 300).

## What Changes

- `tools/discord/feed.py tick`: cap raised to 3800 chars (embed description limit is 4096); the per-field 300 cut is gone; over the cap the text is cut at a line boundary and ends with an explicit `(cut: N chars)` line. Sanitizer, idle refusal and empty values unchanged; still one post.
- Tests for the cap boundary and the cut marker; the two docs name the new cap.

## Impact

Owns tools/discord/feed.py, tools/discord/tests/**, docs/process/feed-after-events.md, docs/community/discord/feeds.md. Nothing is posted by this card.

## Shared-file lines for the merger

- CHANGELOG: `- feed.py tick: the whole tick report is posted (cap 3800 chars, was 900); over the cap it is cut at a line with a final "(cut: N chars)" line.`
- docs/INDEX.md: no new file; the feeds.md entry stays as is.
- docs/BACKLOG.md: none (request done).
