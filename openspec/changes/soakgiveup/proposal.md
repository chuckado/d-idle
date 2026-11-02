## Why

The SOAKSTALL fix (25918ea) skips its give-up for the planner's run area, but the planner sets a run for every leveling step, so in the Gate 1 stall 20 of 22 explore timeouts were uncounted (docs/soaks/gate1-20261101.md, Review). Observable behaviour changes only where an emptied clear-run area was stuck (a log line `area abandoned` appears, an area is marked cleared), no spec number moves: no observable behaviour change to any contract in openspec/specs.

## What Changes

- internal/bot/goal.go `exploreGaveUp`: the quest guard becomes `l.run.Kind != RunClear` (quest and boss runs keep roaming, clear runs may give up).
- internal/bot/emptyarea_test.go: a clear run in the failing area gives up after two failures; a quest run still does not.

## Impact

No spec delta. Overlap (declared): internal/bot/goal.go and emptyarea_test.go are held by the finished card soakstall (merged, awaiting archive); no other card in flight touches them.

Also in this branch (agent/soakfixes): ping-pong guard (noteArea) in goal.go, tpDebounce for povertyflip. Overlap (declared): internal/bot/goal.go and runner.go.
