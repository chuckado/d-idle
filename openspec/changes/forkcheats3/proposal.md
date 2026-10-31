## Why

The bot's missile dodge never fires on d2d because the state source has no missile listing: d2d exposes only `state missiles=<count>` (card fixmissile asked for a `debug missiles` verb). The user's standing approval of 2026-10-31 allows adding functionality to the private fork when documented and vetted, and keeps it out of upstream pull requests. no observable behaviour change: the bot and the specs are untouched; this card records the new lab verb and its live proof (docs only).

## What Changes

- The lab d2d binary gets `debug missiles` (read only: one row per missile in flight with position, direction, speed, owner type and unit, ttl). The patch lives in the private fork, never in this repo.
- docs/cheats.md: the verb's row (columns, units, read-only, negative cases). docs/forkcheats.md: live results, the harness regression and the note for fix-missile (docs/fixmissile.md is that card's).
