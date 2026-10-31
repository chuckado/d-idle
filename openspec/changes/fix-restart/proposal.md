## Why

Live drill D10 (docs/probes/drills/2026-10-31-d10.txt): a kill -9 of the slot's d2d recovered fine but (1) the minute line's `restarts` did not rise, because a game-lost recovery never went through `noteRecycle` and d2d shows no menus, so the Gate 1 criterion "restarts +1" could never pass; (2) `xpRisk` always equalled the whole XP on d2d because `runD2D` built the bot dependencies without `SavedExp`.

## What Changes

- Shared: `gameLost` counts a successful restart in `lostRestarts`; `restartCount` takes max(menu entries - 1 - planned, recycles + lost), so Game.exe is unchanged (the menu count already includes it) and d2d gains the missing +1.
- d2d only: `SavedExp` reads the character's save file under the slot's managed d2d home (re-read when the file changes, on every Save and Exit); nil for an attached d2d (fallback: exp at the last entry or restart).
- Shared: a restart or recycle re-arms the fallback saved exp.
- Delta: soak-gates ADDED requirements.
