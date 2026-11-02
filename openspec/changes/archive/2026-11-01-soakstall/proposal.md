## Why

Soak run gate1-20261101-1704 stalled at clvl 10: no kill and no XP for 58+ virtual minutes while the bot ping-ponged between Blood Moor and Cold Plains.

## What Changes

- docs/soakstall.md: cause, log evidence, fix, live drill checklist.
- internal/bot/goal.go, runner.go, emptyarea_test.go: a failed explore in a quiet area, repeated with no kill, marks the area cleared (per-area count).

## Impact

No spec delta. Overlap (declared): internal/bot/goal.go is also touched by cards in flight on the planner (m2-roam task 1.3 turns bot.emptyAreaSec on; this fix does not depend on it and does not change that knob).
