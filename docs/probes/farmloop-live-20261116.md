# Farm loop live check in the sandbox, 2026-11-16

> status: record. Task farmlive (TASK-553). Tags: V = verified in this run (container logs, read only), U = not verified.

Request: turn `bot.farmReset` on again after the guards of the thrash fix (docs/farmloop.md section 6; image devp9, sandbox on main 87b285515) and watch it for 45 minutes. No watch token, no POST.

## Preconditions (V)

Container healthy, 0 restarts, hero clvl 12 alive, 0 deaths, 0 ERROR lines. Baseline window 14:47 to 16:30Z (102 minute lines): 8.47 kills/min, 49 xp/min, 0 deaths; 25 `Save and Exit` lines (about one per 4 minutes), all with the new `new game requested from=farmreset.go:291` caller where the line exists (only since the 16:09Z deploy: 3 lines). Caller note: `farmreset.go:291` is `startFreshGameDiff`, the one shared start, so the leveling slice's "everything cleared" fresh game and a farm reset show the same `from=`; a farm reset is told apart only by its own `farm reset:` line. Log reading caveat: `docker logs` without `--tail` stops at 14:19Z on this container (a read stops at one bad entry); `--tail 30000` and filtering by time works.

## Enable (V)

`tuning.json` (instance state dir) backed up as `tuning.json.bak-farm2`; `"bot.farmReset": 1` merged next to the existing overrides (invdesk.verbs, invdesk.instant and the rest kept); `docker restart d2play-20261108` once at 16:31:20Z; the hero resumed from its save.

## Observation 16:31:20 to 17:16:27Z (45 minutes, 9 checks, V)

| | before (102 min) | farm loop on (43 minute lines) |
|---|---|---|
| kills/min | 8.47 | 9.14 |
| xp/min | 49 | 50 |
| deaths | 0 | 0 |
| ERROR lines | 0 | 0 |
| `farm reset:` lines | 0 | 0 |
| `farm reset disabled: thrash` | 0 | 0 |
| `new game requested` | n/a | 8, all `farmreset.go:291` |
| container | healthy | healthy, 0 restarts |

The 8 fresh games (16:39, 16:46, 16:49, 16:53, 16:59, 17:04, 17:10, 17:14) each followed `sweep: coverage 124/124 points ... cleared=true` in area 17 (Burial Grounds, one 79/79): every one at 100 percent coverage, the normal leveling fresh game, the same path as before the switch. Gaps 3 to 6 minutes, never a cluster inside a minute, no burst of leaves. The loop itself never fired: the zone is cleared in 2 to 5 minutes, so the hero is never at 85 percent with the 4 minute and 3 minute rate guards met before the area is done. Kills/min and xp/min differ from before within noise (one hero, one zone, 43 minutes: U that the difference means anything).

## Verdict

Keep ON. Clean for the whole window by every abort rule (no death, no ERROR, no thrash line, 0 farm resets). One reading of the "more than 1 leave per 5 min that is not a farm reset" rule: leaves 16:46/16:49 and 16:49/16:53 are 3 to 4 minutes apart, but they are the cleared-zone fresh games that ran at the same rate with the loop off (about one per 4 minutes in the baseline), so they were not counted as a burst. The thrash fix holds (no repeat of the 5-in-20-s cycle). What this does NOT show: that a farm reset works live (`farm reset:` never appeared, so the value rule, the walk back and the cooldown are still V offline only). To exercise it, run a hero in a slow-clearing leveling zone (a larger area than Burial Grounds) with `bot.exploredPct` below 100.

Rollback (not needed): copy `tuning.json.bak-farm2` over `tuning.json`, `docker restart d2play-20261108`.
