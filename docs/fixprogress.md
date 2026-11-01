# FIXPROGRESS: Gate 1 progress criterion

> status: current. Fix for the NOKILL violation of the Gate 1 rerun (docs/soaks/gate1-rerun.md on branch agent/gate1). Tags: V = counted from the logs named here, I = inferred from code, U = unverified. Not verified live.

## Evidence (V)
Run: bot f4ca264, fork binary lab2-freezeresume, fresh Sorceress, 22.6 virtual hours. Sources: the per-minute log docs/probes/soak-logs/gate1-rerun-minutes.jsonl.gz and the full bot log on the worker host.

Timeline (worker clock):
- 14:25 to 14:27: clvl 13 to 14, normal trips, shop visits show `tp 4..5, tome true` (the pack holds one tome, no loose scrolls).
- 14:27:46: death in Black Marsh (belt 0 healing, gold 82). Recovery: `d2d town portal: no portal source (empty)`, the waypoint is not on that level, then recovery succeeds. 14:27:47 shop visit: `Scroll of Town Portal` x1 bought, gold 0, `tp 0`.
- 14:28 to 14:56: normal play, XP 141k to 161k. Many trips, each costs a scroll; scrolls are bought again in most visits (tp 1 to 5).
- 14:56:58 and 14:57:10: ToTown in Burial Grounds logs `no portal source (empty)` and walks (neighbour Cold Plains, then the waypoint). The shop visit at 14:57:00 buys 1 portal scroll for 100 gold and logs `tp 2, tome true`; back to the field `no portal in the unit table`, on foot.
- 14:57:10: field check `Going to town ([no town portal scroll])` again. 14:57:12 chores in town: `chores done ... skipped` lists every chore including `shop`, gold 433, no `shop: visit` line. Back on foot to Burial Grounds.
- 14:57:10 to 15:03: this loop repeats every 12 to 14 wall seconds (about 14 minute lines, gold stays 603): minute log shows step alternating `Going to town ([no town portal scroll])` and `Returning to the field`, areas 17, 3, 2, 1, 1, 2, 3, 17. Kills 6335 to 6349 over the window; XP flat 161398 to 161412. The loop ended at the 15:07 visit (`shop: visit` buys 2, then 5 scrolls).

What the log proves (V): the 43-minute stretch is a trip loop (walk to town, no purchase, walk back), not an unbounded `nav stuck` loop. Each leg finished in seconds and every call ended inside the town timeout. The 103 `nav stuck` lines are the normal rung 0/1 recoveries of the town to Burial Grounds walk (`nav no exit known yet, exploring`, `roam: done` 345 points in town), at most two per leg. The death itself is not the cause: scrolls were rebought after it (14:29, 14:33).

## Cause
- V: two parts of the bot disagree about the same pack. The field check (`NeedsTrip`, `combat.PortalSource`, d2d) counts 0 portal charges; the town visit (`town.scrollCount`) read 2 and so did not see anything to buy (`needScrolls` false, belt full, shop skipped).
- I: `town.qty` and `combat.PortalQty` apply the same rule (stat 70 of the tome, 0 when missing) but read different snapshots. On d2d `UnitStats` answers from the last snapshot's item extras (internal/d2denv/reader.go), so a stale or shifted extras entry reads as charges. This was not reproduced offline.
- U: the exact source of the 2 (stale extras, or a scroll counted in the tome that d2d does not treat as a source). The log has no line for the tome's raw stat.
- Because the visit judged its own reason fixed, `noteFutile` never marked it, so nothing stopped the loop.
- Not the cause: the reserve policy. KitBuy, tpAllowed and the shop target already buy a scroll first (V: 14:57:00 purchase). A "reserve scroll" rule would not have helped because the bot believed it had scrolls.

## Fix (internal/town)
- `ToTown` records `NotePortalEmpty()` when the portal cast finds no source (ErrNoPortal). The next `Chores` treats a tome as holding 0 charges (scrollCount), so the shop need check is true and the shop buys up to the existing target (TPScrollTarget, the gold-floor caps stay). The game's answer outranks the stat read; no class or level is hard-coded.
- Watchdog: two town visits in a row that began with that fact (`maxEmptyVisits`) silence the `no town portal scroll` trip for `FutileMinutes` (existing knob, default 10) and log one warning. A portal that opens again resets the count. Escalation order is therefore: restock against the game's count, restock once more, then fight on with the existing walk-home path as the only way out, then retry after the quiet time. The walk itself already runs under the trip's town timeout (TownTimeoutSec), so no second timer was added.
- Unchanged: all other trips, the walk-to-town path (townNeighbour, hopToNeighbour, waypoint), the resource policy.

Tests (`go test -p 1 -count=1 ./internal/town ./internal/bot` run on the worker host, pass; internal/town skips without the gear tables, so run it on the worker host): `TestEmptySourceRestocksAgainstTheGame`, `TestTwoEmptyVisitsQuietTheScrollTrip`, `TestOneEmptyVisitDoesNotQuietIt`.

## Live recipe (not run)
Slot with a throwaway character at clvl 10 to 14 in Cold Plains, gold about 600, belt full, one empty tome in the pack (no scroll), life below 50 percent. Expect: `no portal source (empty)`, walk or waypoint to town, `shop: visit` buys scrolls (`tp` rises and the next ToTown opens a portal with `back to field: portal in town`), arrival in town within TownTimeoutSec. Short soak leg: 30 virtual minutes from that start; pass = no `chores done` with `shop` skipped twice in a row while `no portal source` is logged, longest no kill and no XP stretch under 10 virtual minutes, and add a log line of the tome's raw stat 70 at the first `[no town portal scroll]` trip to settle the U above.

## Shared-file lines for the merger
- docs/BACKLOG.md: `- FIXPROGRESS: Gate 1 NOKILL loop (empty portal source seen by the field, not by the town shop) fixed in internal/town, watchdog after 2 visits; awaits a live proof (docs/fixprogress.md)`
- CHANGELOG.md (unreleased, fixes): `- fix: a town visit after "no portal source" restocks scrolls and a repeating scroll trip goes quiet after two visits (docs/fixprogress.md).`
- docs/INDEX.md: `- [fixprogress.md](fixprogress.md): Gate 1 progress loop, cause and fix`
- docs/decomp-queue.md, docs/research/README.md: none.
