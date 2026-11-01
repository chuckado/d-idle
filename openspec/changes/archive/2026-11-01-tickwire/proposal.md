## Why

The owner approved (2026-11-01) wiring the optional Haiku tick helper into the dispatcher's Discord tick post. It saves the coordinator session the output tokens of writing the tick text (about 600 to 900 per tick); decisions still need a Claude Code session.

## What Changes

- `tools/discord/tick_post.py`: facts JSON on stdin, Haiku writes the field text, the same sanitizer and privacy guard as `feed.py tick` validate it (one Haiku reword of the offending phrase, then a deterministic plain rendering, no API); any API failure falls back and says so on stderr; dry run by default, `--send` posts one embed to #progress.
- `tools/discord/tests/test_tick_post.py` (fake HTTP and sender), added to `tools/ci/scripts.sh`.
- `docs/process/api-credits.md`: what is wired, fallback, measured token use.

## Impact

One real API call and one real post to #progress were made as the user-approved wiring test; numbers are in docs/process/api-credits.md. Overlap (declared): `tools/ci/scripts.sh` and the `tools/discord` and `tools/ci` areas belong to earlier cards; this card adds files there and one line to scripts.sh, and edits nothing in feed.py or haiku_tick.py.

## Shared-file lines for the merger
- CHANGELOG: `The dispatcher tick can be written by Haiku (tools/discord/tick_post.py, plain fallback); see docs/process/api-credits.md.`
