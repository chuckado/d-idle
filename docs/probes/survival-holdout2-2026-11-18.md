# Survival A/B holdout 2: gold gap, recovery without potions, return portal (2026-11-18)

> status: current (card survhold2, user decision 2026-11-17: `bot.zoneModel` stays 0 until the holdout passes; nothing was flipped, no default changed). Follows [holdout 1](survival-holdout-2026-11-17.md). Tags: V verified (run logs of the builder, `abreport -death-slack 1 -chicken-slack 1 -xp-floor 0.95`), I inferred, U unverified. Lockstep d2d on the builder (new long lane for the Sorceress wave, the 45 minute main lane for the Paladin waves), heroes CHEATED for provenance (SETUP clvl 12, 500 gold, free Rogue, Paladin weapon `hax`; Sorceress none): a relative A/B, not evidence about game behaviour. Writes after start: 0 for every run of this card (V: the last line of each of the 24 logs).

## 1. What was wrong
1. **The gold gap was mostly a symptom (V, I).** In holdout 1 the margin arm ended on 635-1067 gold against 2179-2886. Per zone (minute lines, `tools/survival` logs): the legacy arm spent 15-20 of 90 minutes in Black Marsh (about 40 gold/min on seed 149) and Dark Wood; the margin arm never entered Black Marsh on seeds 149, 163, 167 and farmed Stony Field, Dark Wood and Cave Level 1 (zone income 1.5-45 gold/min, lumpy: sales dominate pickups). Potions used were not higher (38-57 against 43-70). The model's gold term is `lambda * Net` and `lambda` is 0 once the purse covers the kit (`internal/survival/controller.go lambda`), so a zone that is safer and pays the same XP is chosen without regard to its gold: a design property, not a bug. With the two fixes below and the same seeds the gap closed (margin 3.5-4.8k gold on 149, 163, 167 against legacy 0.6-3.1k on the same new binary); I could not isolate which of the fixes or of the other merges since holdout 1 did it (U). No gold weight was added: it would be tuning on the holdout.
2. **Recovery without potions (V, code and test).** `kitHolds` from clvl 8 (`resource.kitGateLevel`) held the Blood Moor / Cold Plains gate only for a hero with no portal AND no potion. Seed 149 had a portal scroll and an empty belt, so the gate was off, the recovery picked the XP-best area (Cave Level 1) with 5 gold and chickened three times in 20 s. Fix (`internal/bot/kitgate.go`): a recovering hero from clvl 8 with no healing potion and gold under `resource.kitPotionGold` (30) is held on the first fields; when the purse reaches the price the hold lifts and `kitTrip` buys the potion first. Must-fail test `TestKitGateHoldsBrokeHeroWithoutPotions` (Sorceress and Paladin; it fails on the old code, V).
3. **Return portal "wanted 6, in 9" (V, code and test).** The town portal leads to where the last scroll was read (`ob.portalFrom`), not to the zone the trip returns to. When the model (`zoneRecoverReturn`) or a gate redirected the return to another zone, `BackToField` entered the portal into the zone that had just hurt the hero and only then walked on. Fix (`internal/bot/zonemodel.go returnToField`, used by the trip's last stage): when the landing zone is burned or barred and differs from the target, walk from town (`travelTo`, waypoints and exits) and fall back to the portal only if the walk fails. Tests `TestReturnToFieldWalksPastBurnedPortal` / `...PortalWhenLandingIsClean` (both builds). The line `back to field: portal lands in a burned zone, walking instead` did not appear in any of the 24 runs (the case did not recur in these seeds), so the fix is unit-tested but not seen live (U).
4. **Residual (V, not fixed).** Paladin seed 167 margin: two chickens in 40 s in Underground Passage Level 1 (area 10) with 4 and 2 potions in the belt after a chicken (leave game) and the re-planned walk; the model had just chosen Cave Level 1. A plan/model mismatch after a chicken, not a potion problem. Sorceress seed 103 margin: see section 3.

## 2. Paladin (class 2, `AB_SPEED=5`, horizon 90 virtual minutes, 6 seeds, BOTH arms on the new binary: main as of this branch plus the two fixes)
Seeds 149, 163, 167 (holdout) and 171, 173, 177 (fresh). The legacy arm was rerun on the new binary because main moved since holdout 1 (other merges), so the arms differ only in `bot.zoneModel`.

| seed | legacy xp@90 / chick / gold | margin xp@90 / chick / gold |
|---|---|---|
| 149 | 32565 / 0 / 1616 | 38908 / 0 / 4786 |
| 163 | 24391 / 1 / 567 | 34791 / 1 / 3521 |
| 167 | 31702 / 1 / 3146 | 36651 / 3 / 3684 |
| 171 | 31810 / 0 / 1480 | 28033 / 1 / 1518 |
| 173 | 30411 / 4 / 234 | 32941 / 0 / 2753 |
| 177 | 28461 / 0 / 2557 | 33245 / 1 / 3244 |

Pooled: deaths 0 / 0, chickens 6 / 6, XP ratio 1.141, per-seed difference mean +4205, 95 percent interval [-726, 9136]. **Verdict (abreport, slack 1/1, floor 0.95, 6 pairs): PASS.** Holdout 1 on the old binary read FAIL (chickens 4 against 0); the legacy arm of the new binary itself chickens 6 times, so the chicken comparison moved with the baseline. Not the 8 seeds the flip needs (docs/ab-design.md): 3 more fresh seeds are missing.

## 3. Sorceress (class 4, speed 5, 6 runs at once on the long lane, seeds 101 103 107, both arms on the new binary)
The 48 minute wall limit of the run reached 70 to 120 virtual minutes (6 concurrent runs are slower), so the horizon is 70 (the shortest pair).

| seed | legacy xp / chick | margin xp / chick |
|---|---|---|
| 101 | 15823 (minute 95) / 0 | 14173 (minute 80) / 0 |
| 103 | 23602 (94) / 0 | 2513 (69) / 1 |
| 107 | 12121 (119) / 9 | 17514 (118) / 0 |

Deaths 0 / 0, chickens 9 / 1, XP ratio at minute 70 0.792, interval [-25728, 19722]. **Verdict: INCONCLUSIVE (3 pairs; leaning FAIL because of seed 103).** Seed 103 margin farmed 2.5k XP in 69 minutes: Cold Plains (54 xp/min), Burial Grounds flagged `mana-starved` with realized 6 xp/min against a model 152, a recovery cycle at gold 40, and a chicken in Underground Passage after a `pays_for_risk` pick (Dark Wood, margin 0.7, k 3.00 = the clamp). The model priced a mana-starved caster at its fictional kill speed and `k` sat at the clamp for an hour; this is a model defect for casters not touched by this card (U: the starvation multiplier `starveValueMul` applies only after `starveCasts` lines). Seeds 101 and 107 are at or above legacy (107: legacy 9 chickens).

## 4. Recommendation (the user decides; nothing flipped)
1. **Paladin: the model now passes a 6-seed A/B (XP +14 percent, 0 deaths, chickens equal).** Two more fresh seeds complete the 8 of the ab-design before a flip; run them on the long lane at horizon 120 (4 at once, speed 5).
2. **Sorceress: no flip.** One catastrophic seed (103) from a mana-starved zone choice. Next card: make the zone value use the realized yield sooner for a starved caster (lower `yieldMinMinutes` or `starveCasts` for the caster archetype) and stop the `pays_for_risk` pick when `k` is at `kMax`; then 8 fresh Sorceress seeds.
3. Keep both fixes merged: they cost nothing with the knob at 0 (the kit gate fix also helps the legacy arm in recovery).
4. A per-class default is possible: Paladin first, Sorceress later.

## Provenance
Heroes `SaHero-<seed>`, SETUP in `prep.abmerc` (declared): [test-characters](test-characters.md). d2d lab binary of holdout 1 (`d2d.survtune-base`), d2bot built from agent/survhold2 (kit gate hold and return-walk fixes, main merged 2026-11-18). Builder work dir `d2d-lab/survhold2` (waves: Paladin on the 45 minute lane twice plus once, Sorceress once on the long lane; `tools/survival/wave.sh`). Heavy runs only on the builder, never on the user desktop. Processes left: none (each wave ended on its own).
