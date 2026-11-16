# Survival A/B rescreen, Paladin, bot.zoneModel 0 vs 1 (3 seeds), 2026-11-16

> status: current (TASK-563). Rerun of [the first screen](survival-ab-screen-2026-11-16.md) after the clock fix of the merged survfix card ([analysis](../survival/ab-screen-analysis-2026-11-16.md)). Harness: [ab-harness](../survival/ab-harness.md). Tags: V verified (run logs, `abreport` output, ledger files of this run), I inferred. Lockstep d2d only (no game.exe); the heroes are CHEATED for provenance (SETUP: clvl 12, 500 gold, free Rogue), so this is a relative A/B, not evidence about game behaviour.

## Setup (V)
- Binary: `d2bot` built from main at the survfix merge (one clock, `be.Clock.Now`), one binary for both arms; the arms differ only in `bot.zoneModel` 0 vs 1, `bot.portalGuard` 0 in both. Same seeds (101, 103, 107), class Paladin (`AB_CLASS=2`), 2 virtual hours, `AB_MAX=4`, as the first screen. The harness hero still has no attack skill attached (the proposal of the analysis is not part of this run).
- `AB_CLASS=2 AB_MAX=4 tools/survival/ab.sh 2 101 103 107` under `tools/agentrun.sh --ttl`; about 38 minutes wall for the six runs on four slots (orion load 12 to 22 on 16 cores during the run, idle plus nice 68 percent at the start; the user's sandbox container and volume untouched). The legacy run of seed 103 hit its wall limit at 111 minutes (host load), so it was rerun alone (`AB_SPEED=5`, longer wall cap, same seed): lockstep made the rerun identical in outcome (0 deaths, 0 chickens, 1 merc death, as in the first screen).
- Every log ends `writes after start: 0` (6 of 6); none CHEATED, none excluded.
- Report: `abreport -base legacy -arm margin -death-slack 1 -chicken-slack 1` (XP is now read at the last minute line of the pair, labelled `xp@N`).

## Result (V)
| seed | arm | deaths | chickens | merc deaths | XP at 30 / 60 / 90 / last minute |
|---|---|---|---|---|---|
| 101 | legacy | 0 | 0 | 1 | 6581 / 12655 / 16728 / 19508 (@119) |
| 101 | margin | 0 | 0 | 1 | 6581 / 12655 / 16025 / 19945 (@118) |
| 103 | legacy | 0 | 0 | 1 | 3460 / 19026 / 23347 / 30471 (@118) |
| 103 | margin | 0 | 1 | 5 | 10347 / 24360 / 32733 / 44668 (@118) |
| 107 | legacy | 0 | 1 | 5 | 9568 / 18288 / 24245 / 32853 (@118) |
| 107 | margin | 0 | 1 | 5 | identical to legacy (same trajectory) |

Pooled: deaths 0 vs 0; chickens legacy 1, margin 2 (limit 1 + slack 1 = 2: ok); XP at the horizon test-base mean +4878, sd 8073, 95 percent interval [-15179, 24935], ratio 1.177 (margin higher). `abreport` verdict: **PASS** (exit 0). The first screen's failure (4 chickens against 1) is gone: seed 101 no longer chickens or dies at minutes 106 to 120 and its two arms stay close.

## Does k move now? (V: ledger.json of each margin instance and the `zone choice:` lines)
- The clock fix works at the ledger: the chicken events land in the ledger with `hard` 1.5 and a consistent time stamp (seed 107: area 5, one chicken; seed 103: area 10, one chicken). Before the fix the `hard` multiplier never took effect.
- `k` on the printed `zone choice:` lines is still `1.00` in every line (101: 9 lines, 103: 37 lines, 107: none). That is not a contradiction: `k` is printed only for zones that lose on margin (Black Marsh here), and the zones that had an event (area 10, area 5) are not among the printed alternatives after the event, or no choice is made at all (seed 107: the margin arm never consulted the controller, so both arms are bit-identical again). So the visible effect of the learned `k` on a decision is still not shown in any run (U for "k changes a pick").
- Two open suspects found in the ledgers (V that the numbers are 0; the cause is I, not traced): (1) `mercDeaths` is 0 in every zone row although 1 to 5 merc deaths happened per run and `ZoneEvMercDied` is wired in `cmd/d2bot/survivalwire.go`; (2) the fight rows have `kills` 0 and `seconds` 0 everywhere, so k_zone from taken/predicted can never leave 1. Both mean k only learns from chickens and deaths in this harness.
- `zone choice:` differences: the margin arm logs them (101: 9, 103: 37) and the legacy arm almost none (101: 2). The picks seen are Dark Wood (best pay among the safe zones, margin 3.2 to 3.8) and Stony Field when the best zone is cleared (`depleted_next`); Black Marsh is refused all run (`margin_low`, margin 1.0 to 1.5, need 1.5). Seed 103 diverged at minute 13 and the margin arm earned 47 percent more XP (44668 vs 30471) at the cost of 4 more merc deaths and one chicken.

## Verdict
**The screen PASSES the agreed rule** (Paladin, 3 seeds, slack 1). It is still a screen: one seed (107) is bit-identical, one (101) is within 2 percent, only 103 differs materially; the XP interval spans zero by a wide margin; merc deaths are 11 against 7 pooled (not in the gate).

## Recommendation: 8 seeds
Run the 8-seed confirmation before the default flips (the user's gate). Cost, scaled from this run (2.2 slot-hours for 3 seeds, about 38 minutes wall on four slots under load): 8 seeds x 2 arms x 2 virtual hours = 32 virtual hours, about 6 slot-hours, about 1.7 hours wall on four slots (give the wall cap `AB_SPEED=5` so loaded runs do not end short). Two things worth doing first, both small: attach an attack skill to the harness hero (the analysis proposal) and look at the two ledger suspects above, because a controller that cannot learn merc deaths or fights may pass 8 seeds for the wrong reason. Sorceress was not part of this brief (first screen: XP 24 percent lower, inconclusive) and still needs its own rescreen.

## Provenance
Heroes `SaHero-101`, `SaHero-103`, `SaHero-107` (Paladin, 6 runs plus one rerun): [test-characters](test-characters.md). Logs: the worker host survab run folder (`ab`), previous results moved to `old1`.
