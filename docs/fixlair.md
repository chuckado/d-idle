# fixlair: Duriel's Lair Teleport value 2

## What changed
- `internal/data/levels.go`: `Level.TeleportRule()` models the Levels.txt Teleport column (0 refused, 1 allowed, 2 allowed on a clear line, mask 0x804) and `CanTeleportIn` adds the town ban (V read FUN_0056f640: refused in any town room regardless of the column). The value still comes from the real levels.txt the repo loads (`loadLevels`), nothing is hand-written.
- Comments in `internal/data/levels.go`, `internal/boss/act2.go` and `docs/act2.md` no longer say "forbidden".
- Tests: `TestTeleportRule` (three values, unknown, towns), `TestRealLairTeleport` (with a real install: only area 73 has value 2; skipped without one), `TestDurielDoesNotTeleport`.

## Safe-change decision
Data and comments only. The Duriel script keeps its movement (walking and kiting, `Anchor.Teleport` false): a clear-line hop needs the line test (doors and missile barriers) that the bot does not have, a refused hop costs mana on d2d (SKTELE B2), and nothing in the repo consumes `Level.Teleport` yet, so no behaviour moves. Using Teleport in the lair is a later, separate card.

## Tags
- Value 2 = clear line, only the lair: V read of FUN_005ca360 (docs/research/sk-teleport-trace.md, branch agent/sktele, unmerged at writing) and D (levels.txt dump).
- Town refusal: V read FUN_0056f640.
- The real-data test is UNVERIFIED here (no install on the build box); run it on the worker host.

## Live proof idea
Drill T6 of the trace: `debug goto 73`, hop across a doorway (refused) and along a clear corridor (lands). It needs the lair, which sits behind the Act 2 quest chain and is usually unreachable early; until then the offline table test is the proof.

## Shared-file lines for the merger
- docs/BACKLOG.md: update the SKTELE line: "bot B6 fixed offline by FIXLAIR (docs/fixlair.md): data and comments only, the Duriel script still walks".
- docs/INDEX.md: `docs/fixlair.md`: Duriel's Lair Teleport value 2 (clear line), data model and the safe-change decision.
- CHANGELOG.md: "Level data models Teleport 0/1/2 (lair = allowed on a clear line); comments corrected; no behaviour change."
