## Why

SKTELE finding B6: game.exe's Teleport handler refuses only when the level's Levels.txt Teleport value is 0, and with value 2 only when the line to the clicked point is blocked. Duriel's Lair (73) is the only level with value 2, so Teleport is allowed there along a clear line. Our code and docs read the lair as "no Teleport".

## What Changes

- data.Level gets TeleportRule (refused, allowed, clear line) and data.CanTeleportIn (towns refuse); the raw column comment is corrected.
- Comments and docs/act2.md say "allowed on a clear line" for the lair.
- No behaviour change: the Duriel script keeps walking (Anchor.Teleport stays false), a test pins that.

## Impact

Offline; no live proof possible early (the lair needs the Act 2 quest chain, see docs/fixlair.md). No spec names the lair teleport flag (grep openspec/specs), so no spec delta.
