## Why

SKILLINV finding F1: a fresh Sorceress never learns Static Field (BUILDV2 puts it at clvl 86) while the Andariel card and `staticOpener` assumed it. User decision 2026-11-01: drop it for fresh characters. SME: Static Field never kills and its radius (5 subtiles at level 1) is far below the Andariel kite range (10+).

## What Changes

- The opener, cold-immune fallback and knobs stay but only act for a character that already learned Static Field (they already look the skill up in `KnownSkills`; V by reading and a new test); the legacy table (`SelectBuild`) is unchanged. No code is deleted, no knob changes.
- The bot `minute` line carries `table=<skill table>`; `d2dsoakstat` parses it and `d2dsoak` prints one `table=` line.
- Prose that said Andariel opens with Static Field is corrected; docs/fixstatic.md, BACKLOG F1 resolved, INDEX, CHANGELOG.
- No spec names Static Field (grep openspec/specs); the sentence in the open card m3-andariel delta is stale and left to that card.
