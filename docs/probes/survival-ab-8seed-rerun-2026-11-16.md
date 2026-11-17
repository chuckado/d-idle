# Survival A/B 8-seed rerun, Paladin (and a Sorceress screen), fixed harness and retuned controller, 2026-11-16

> status: current (TASK-572). Follows [the 8-seed run](survival-ab-8seed-2026-11-16.md) (FAIL on XP). Harness: [ab-harness](../survival/ab-harness.md) sections 6 to 9. Tags: V verified (run logs, `abreport` output, prep ledgers of this run), I inferred, U unverified. Lockstep d2d only (no game.exe); heroes are CHEATED for provenance (SETUP: clvl 12, 500 gold, free Rogue, weapon `hax`), so this is a relative A/B, not evidence about game behaviour.

## What changed since the failed run (V)
1. **Weapon read.** The old `quantity=-1` reply of `debug equip` is the verb's normal answer. The old d2d build (Nov 6 lab3-labverbs) has no `stats.weapons` and no `stats.hero` in `dump stats`, so the bot could not read any weapon; every Paladin line said `src=default dps=30`. `tools/ab/prep` now reads the weapon back after the SETUP verbs and refuses to save a hero whose weapon is not there; the d2d now comes from fork branch `didle/rebase-20261115` (built on the NAS builder with `--jobs 2`; `--jobs 6` is killed for memory on `main.cpp`). Proof: prep ledger `weapon hax slot 4 base 3-6; hero hand slot 4 dmg 3-6`, bot `zone entry dmg 3-6 ar 70`, `zone shadow dps=19.9 src=skill`. The read damage is LOWER than the old default (19.9 against 30), so the weapon was not the cause of the XP gap.
2. **Controller retune** (card survtune, `internal/survival`, data in `params.json`, all U): the pay-for-risk route (a zone the bad pack calls too tight is entered when the typical pack is comfortable and the bad pack's corrected best margin is at least `badFloor`; value times `riskValueMul`), `badFloor` 0.4 balanced (first 0.7), and `freshNeedsMerc` false (a dead merc no longer blocks the fresh game of a pick that is eligible without him). Default `bot.zoneModel` stays 0.
3. **Both arms rerun** on the new d2d: with weapon read, the legacy arm is far safer than in the old build (0 to 1 chicken a run, no deaths), so the earlier safety win (chickens 10 against 36) is not reproducible; it belonged to the old d2d build.

Binary: `d2bot` built from the branch (legacy arm and margin arm differ only in `bot.zoneModel`; the legacy arm ran with the pre-retune binary, the controller code does not act when the knob is 0). Seeds 101 103 107 109 113 127 131 137, Paladin, 2 virtual hours, 16 runs per controller version, three or four at once at nice 19 (the user was gaming on the host; load 5 to 21 on 16 cores). All 16 logs of each arm end `writes after start: 0`; none CHEATED beyond the declared SETUP, none excluded, no run short.

## Result, Paladin, three controller versions (V: `abreport -base legacy -arm margin -death-slack 1 -chicken-slack 1`)
| version | XP ratio margin/legacy | 95 percent interval of the XP difference | deaths legacy / margin | chickens legacy / margin | verdict |
|---|---|---|---|---|---|
| old controller, old d2d, no weapon read (merged doc) | 0.823 | [-9236, -252] | 2 / 1 | 36 / 10 | FAIL |
| route with `badFloor` 0.7, new d2d, weapon read | 0.895 | [-11942, 3262] | 0 / 0 | 4 / 3 | INCONCLUSIVE |
| route `badFloor` 0.4 plus `freshNeedsMerc` false (shipped) | **1.002** | **[-4181, 4387]** | **0 / 0** | **4 / 1** | **PASS** |

Final version per seed (xp at the last minute line, start 90180): 101 40963 / 44743, 103 46836 / 45742, 107 31084 / 30127, 109 44840 / 42210, 113 52265 / 43186, 127 34687 / 42674, 131 41059 / 40497, 137 39923 / 43509 (legacy / margin); mean difference +103, sd 5124. Potions 520 / 505, merc deaths 40 / 40 (the controller still does nothing for the mercenary), clvl end 12 to 14 in both arms. Intake per minute is the same in every zone both arms visit (area 6 Black Marsh higher in margin, 63 against 48, over 32 minutes against 177).

## Where the gap came from (V, `zone choice:` lines and per-area xp of the first rerun)
- Black Marsh pays about 640 xp/min and the legacy arm farmed it for 177 minutes; the margin arm 21: after one hard event the pooled correction `k` reads 2.6 (observed fight seconds per kill 1.76 against 0.51 predicted, so every zone's bad-pack margin is divided by about three) and Black Marsh's best margin fell to 0.4. That is why `badFloor` went to 0.4 (the per-zone step-down rows, two chickens in 30 minutes, are the real brake and are unchanged).
- Seed 113 (27k against 52k xp in the first rerun) sat in depleted zones paying 24 to 300 xp/min for an hour: the controller's fresh game was refused with the merc left dead. After `freshNeedsMerc` false it reads 43k.
- Both changes were chosen after looking at 3 to 5 of these same eight seeds: see the holdout below.

## Sorceress 3-seed screen (V, same binary, `AB_CLASS=4`, no SETUP weapon)
Seeds 101 103 107: XP ratio **0.836**, interval [-20581, 14046], deaths 0 / 1 (seed 101 margin), chickens 7 / 7 (5 4 | 0 1 | 2 2), merc deaths equal: **INCONCLUSIVE** (3 seeds cannot show an XP effect under 30 percent, `docs/ab-design.md`). Per seed margin/legacy 1.24, 0.58, 0.69; seed 107's first margin run hit its wall-time cap at 110 minutes and was rerun with a longer cap (`107.short.log` kept). The Sorceress is not shown equal; the route was not tuned on her.

## Holdout (seeds never looked at)
Seeds 139 149 151 157 (Paladin, same binary and d2d, `AB_SPEED=3`, 8 runs, all `writes after start: 0`, none short): `abreport` **FAIL** (chickens 3 against 1 plus the slack of 1), XP ratio 0.967, interval [-15741, 12971] (spans the floor), deaths 0 / 1 (within the slack).

| seed | arm | deaths | chickens | xp at the last minute (start 90180) |
|---|---|---|---|---|
| 139 | legacy / margin | 0 / 0 | 0 / 0 | 40781 / 52094 |
| 149 | legacy / margin | 0 / 1 | 0 / 2 | 42326 / 30709 |
| 151 | legacy / margin | 0 / 0 | 1 / 1 | 39824 / 39098 |
| 157 | legacy / margin | 0 / 0 | 0 / 0 | 44989 / 39090 |

Seed 149 is the whole safety difference: the margin arm took `pays_for_risk` picks at corrected best margins of 0.6 to 0.9 (`k` 2.9 to 3.4: Dark Wood, Black Marsh) after a hard event, lost its merc three times in 80 seconds and died once (V: `zone choice:` lines of `survhold/ab/margin/149.log`). So the 0.4 floor lets the route enter a zone the model itself calls lost-ish at the current `k`; on the eight tuning seeds that never ended in a death, on four fresh seeds it did once. Pooled over all twelve seeds (sums of the two reports): deaths 0 / 1, chickens 5 / 4, mean XP difference about -390 (ratio about 0.99).

## Recommendation (plain)
1. **Do not flip `bot.zoneModel` yet.** On the eight seeds the controller was tuned on, the Paladin gate passes (XP ratio 1.002, no more deaths, fewer chickens); on four fresh seeds it fails the chicken part by one (3 against 1 plus slack 1) with one death, and XP stays equal within noise (0.967). **Nothing here shows the model is better than the plan tables**: with the weapon read and the new d2d the legacy arm is already safe (0 deaths, 5 chickens in 24 virtual hours), so the margin controller buys no measurable safety; it only matches the legacy XP when allowed to take risks the legacy arm takes anyway. The first run's safety win (chickens 10 against 36) belonged to the old d2d build.
2. **Do not flip for all classes**: the Sorceress screen is inconclusive and leans slower (0.836, one death in the margin arm). The knob is global, so a Paladin-only flip needs a per-class default first (not built).
3. Next step if the user still wants the model on: raise the Paladin `badFloor` between 0.5 and 0.65 (the 0.7 version was safe and 0.895 on XP, 0.4 is 1.0 on XP and one death in twelve fresh-ish seeds) and judge on a NEW set of seeds, not on these twelve (the values were chosen with these seeds in view); `freshNeedsMerc` false is the part with a clear effect (seed 113 27k to 43k xp) and no safety cost seen. Both are data (`params.json`), tunable per archetype without code. The Sorceress needs its own 8 seeds first.
4. Process: do not stop an `agentrun.sh` driver while its runs go (it killed three runs at minute 101 to 108 on 2026-11-16, `ab-harness.md` section 8); the queue driver is `tools/survival/abq.sh`.

## Provenance
Heroes `SaHero-<seed>` (Paladin and Sorceress) and `SaProbe`: [test-characters](test-characters.md). Logs on the worker host: `d2d-lab/survtune` (legacy arm and the killed first batch), `survtune2` (first route), `survtune3` (shipped), `survsorc` (Sorceress), `survhold` (holdout). The d2d binary is `d2d.survtune-base-fabae935-linuxbuild` built from fork branch `didle/rebase-20261115` at `fabae935` (private forge, already pushed).
