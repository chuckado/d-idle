# KITGATE: recovery without the kit stays near town

> status: shipped on branch agent/kitgate; knob `recoveryKitGate` (default 1). d2d-only in effect, shared code.

Finding (SOAKWATCH, livestream soak seed 5, d2d): one death at clvl 7 in Burial Grounds in "Recovery: farming gold", belt 0 potions, 0 portal scrolls, 35 gold; 49 x "no town portal available"; near-deaths in the Crypt and Mausoleum; a Cold Plains / Cave 1 border loop. Death rule not violated: a policy gap.

## Why the bot was there (code path)
- VERIFIED by reading: `Runner.target0` (goal.go) for a ReachLevel goal calls `recoveryArea` while recovery runs. When that returned no area (both safe fields cleared or banned after deaths, or `starveStep` found its first field cleared) the code FELL THROUGH to the normal planner, which picks from the quest ladder at the character's level: Burial Grounds, then the Crypt and Mausoleum. The recovery slice (`recoveryStep`, 40 min) then walked there with no scroll to get home. This is the probable route of the death and the near-deaths.
- VERIFIED by reading: `safeTier` and `recoveryXPTier` (broke.go, recoveryxp.go) include the Den, Cave 1 and Cave 2 (and the XP tier Dark Wood and Black Marsh); the only potion rule is "Den needs 2 potions". An empty belt with a portal scroll therefore still farmed the caves.
- VERIFIED by reading: nav `Route` priced every level the same, and `NearMouth` (mouth guard) is off by default, so a walk Blood Moor to Cold Plains or Cold Plains to Stony Field could pass a mouth; a seen crossing through Cave 1 beats the direct hint.
- UNVERIFIED (inferred from the logs, not replayed): that the 91 "stepping away from a cave mouth" and the 20 "left level 2 for 3 but ended in 8" lines are the same mouths; and that the first Burial Grounds entry was the fall-through rather than the quest step. The fix covers both causes.

## Fix
- `kitMissing`: no healing or rejuvenation potion in the belt (`Belt.HP+RV`, as `updateRecovery`) or no town portal scroll (`resInput.HasTP`); unknown belt is not missing. Same parts `resource.KitShortfall` prices.
- Gate on (knob, recovery running, Normal act 1, kit missing): `recoveryArea` keeps to `kitSafeTier` (Blood Moor, Cold Plains), `target0` returns no area instead of falling through (the leveling slice waits at home), and `syncKitGate` gives the navigator `dungeonMouths` (Den, Caves 1 and 2, Burial Grounds, Crypt, Mausoleum): `Route` adds a 30-edge cost to entering them and `NearMouth` answers while set.
- With the kit, or knob 0, nothing changes. Tests: TestKitGate*, TestRouteAvoidLevels.
- Not done (follow-up): the exit-gate choice inside one level (`findGate`) is unchanged; `penalizeExits` already prices mouth tiles. The border-loop breaker still fires for a mouth the walk cannot leave.

## KITGATELEVEL: the gate by character level (card kitgatelevel; user decision 2026-11-02 "adopt the kit gate rule")

Why (V, docs/soaks/gate1r2-20261101.md Review 2 section 6): at clvl 11 and up 366 of 925 virtual minutes were in the Blood Moor and Cold Plains because the gate allowed no other field while the kit was missing; those fields pay about 0.5 XP and 0.4 gold per kill at that level. The gate was written for clvl 1 to 7. Rule set is the d2-sme veteran advisor's.

Rules (code: internal/bot/kitgate.go; knobs `resource.kitGateLevel` 8, `kitMinHP` 4, `kitMinMP` 3, `kitTripCap` 2, generated in docs/knobs.md; `kitGateLevel` 0 restores the old gate at every level):
1. Below clvl 8: unchanged. No healing potion or no portal source holds a recovering character in Blood Moor and Cold Plains.
2. From clvl 8, Normal act 1, an incomplete kit (`kit.missing`: portal source absent, healing potions under `kitMinHP` counting belt, pack and rejuvenation, mana potions under `kitMinMP` in the belt) is NOT a hold: `kitTrip` adds the chore reason `kit: ...` (`choresDue`, not floorable by the trip gold floor) and the existing shop buys the kit; the planner's next area goes on, never back to the two fields because of the kit. The buy order (portal, healing, mana) is the existing belt plan and BuyKit (S: not re-verified here that the shop keeps a scroll reserve when it already holds a tome; with a tome the reserve is moot).
3. Gold short (gold under the price of the cheapest missing item): no kit trip, no hold; the existing sell logic (poverty sell, low-gold sell in `choresDue`, `povertySellMin`, internal/town sell) brings gold and the character prefers whatever area the planner picks (it no longer restricts the fields). The bot has a sell routine, so no follow-up card is needed for it. The trip also needs a way home: a portal source, or a walkable field (Blood Moor, Den, Cold Plains).
4. Hold only with no portal source and no healing potion (cannot flee or heal): `kitHolds`.
5. Cap: after `kitTripCap` kit trips (counted in `logTrip` by the `kit: ` reason) with the kit still incomplete the bot stops asking and logs one INFO line `kitgate-skip` with `trips`, `missing`, `gold`, `level`. A complete kit resets count and flag.
6. The minimum kit lives in `kit` (kitgate.go) and the knobs. Rejuvenation and identify are not required. Mana is asked only where `resource.ManaNeed(class) >= 2` (no class named in the gate; multi-build safe). Belt default.
The tome drop-out (stat 70 absent for a moment) is covered by the debounce in `hasPortalSource` (tpdebounce.go): `kitTrip` reads the debounced verdict, so a false empty tome makes no trip (TestKitGateLevelTomeDropout).

Tests (offline, mem fakes): `go test -p 1 -run 'KitGate|KitMissing|TestKit' ./internal/bot`: TestKitGateLevelHold (table: clvl 7 incomplete holds; clvl 8 and 14 do not unless no portal and no potion), TestKitGateLevelTripThenProgress (clvl 8, 11, 14 trip, then a non-gated area), TestKitGateLevelGoldShort, TestKitGateLevelCapLogsSkip, TestKitGateLevelCompleteKit, TestKitGateLevelTomeDropout. NOT verified live (U).

Live drill checklist (rerun of gate 1 after merge; read the bot log):
- `grep -c "kit gate on"` only at clvl 7 and below, or at clvl 8+ with no portal and no healing potion (rare).
- Each `kit: ` chore reason at clvl 8+ is followed by `chores done` and a `trip budget: trip done` line, then the next area is not 2 or 3 unless the planner picks it.
- Share of minutes in areas 2 and 3 at clvl 11+ below 20 percent (was 40 percent; `d2dsoak bands`).
- `grep kitgate-skip` lines are rare (a few per run at most); each shows trips=2 and what was missing. Many of them mean the shop could not buy the kit (price, vendor stock, mana potion not in the catalog): look at `shop:` lines of the trips before it.
- No death in the first 5 minutes after a trip-less proceed.
