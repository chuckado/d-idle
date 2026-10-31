# fix-restart (drill D10 findings)

> status: current. Card openspec/changes/fix-restart; evidence docs/probes/drills/2026-10-31-d10.txt.

## What was wrong
- `restarts` on the minute line did not move after a kill -9 of the slot's d2d. A d2d never shows the menus, and a game-lost recovery (`gameLost` in internal/bot/loop.go) goes through `Deps.Restart`, not `newGameRun`/`Recycle`, so neither the menu count nor `recycles` saw it. Gate 1 "restarts +1" could not pass.
- `xpRisk` equalled the whole XP on d2d: `runD2D` passed no `SavedExp`, and the fallback (exp at game entry) is only taken on a menu-observed entry.

## Restarts semantics (shared)
`restarts = max(menuEntries - 1 - plannedSavePoints, recycles + lostRestarts)` (internal/bot/observe.go `restartCount`).
- `lostRestarts`: successful `gameLost` restarts, a counter of its own.
- `recycles`: bot-started unplanned new games (`newGameRun`, `Recycle`). The two sets are disjoint (gameLost calls Restart directly), so nothing is counted twice.
- Game.exe: the menu observer sees the recovery's new game, so the first term already includes it and wins or ties; behaviour unchanged.
- Planned save points are never restarts (they raise `plannedGames`).

## SavedExp on d2d (d2d only)
`d2dSavedExp` (cmd/d2bot/savedexp.go) finds `<char>.d2s` under the slot's managed d2d home (cached, a miss is retried every 10 s) and re-reads the exp when the file's mtime or size changes; d2d rewrites it on every Save and Exit, so a save point or restart refreshes it. Between saves the value is the last saved exp, so `xpRisk` is the exp gained since then. UNVERIFIED live: that the file appears in the home exactly as the drill harness finds it (the drill reads it that way). Attached d2d (no managed home): SavedExp stays nil; the fallback is re-armed at each restart/recycle (`savedNew`).

## Drill to rerun
D10: `d2drill -only D10`. Pass: `restarts` delta exactly 1 after the kill; `xpRisk` on the first lines after the kill equals the unsaved XP only (0 right after re-entry), not the whole `xp`.
