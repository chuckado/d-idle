# tphoard: loose town portal scrolls beside a tome

> status: current (card tphoard, 2026-11-20). Tags: V verified (command or source named), S secondary, U unverified.

## Symptom (sandbox hero, Paladin clvl 18, d2d backend)

Minute line `tp 32, invPct 82`; `inv-tp-dump portal_items`: ONE tome (class 518, counted 10) and dozens of single Scroll of Town Portal items (class 529, counted 1 each); three `town portal did not appear ... tool:tsc toolQty:1`. The pack filled with scrolls.

## What the code does (V, read)

- d2d merges a picked scroll into a carried tome with room (V fork source `rules.hpp` `pick_up`: `book.quantity < max_stack`, then `++book->quantity`; a buy does the same, `tome_takes`). A scroll only stays loose when the tome is full or absent. `debug give <code> pack` skips the merge (docs/forkcheats.md), so a given scroll stays loose.
- The pickup rule (`internal/loot/loot.go`, `needsOf.Want("scroll")`) wanted a scroll while `tomeQty < 20`, and `tomeQty` reads a missing stat 70 as 0 (an empty tome has no stat list; docs/tomefix.md records qty dropouts in the field). A full tome whose stat read comes back empty therefore "has room" and every drop is taken: d2d cannot merge it, it lands in a cell. Nothing bounded the loose count. (Cause of the sandbox hoard: S, inferred from the code and the dump; no lab run reproduced the dropout.)
- Nothing sold them: scrolls are in `neverSell` and the `supplies` invariant, and `spareSurplus` only handles tier C.
- `PortalSource` returns the first portal item in pack order, so a loose scroll is cast before the tome and the surplus is burned one per portal, slowly.

## Fix

- Knob `town.looseScrollKeep` (2, range 0..10; was unbounded): the pickup rule takes a scroll beside a tome only while fewer than that many loose scrolls of its kind are in the pack, whatever the tome read says. No tome: unchanged (`tpScrollTarget` / `idScrollTarget`).
- Sell chore (`looseScrollSurplus`, `internal/town/shop.go`): with a tome of the same kind in the pack, loose scrolls beyond the keep are sold (the one exception to never-sell scrolls), at any pack fill, which also raises the sell visit (`needSell`). Without a tome nothing is sold.
- Build-specific part: none (no class or build check).

## The failed casts (V, lab run)

The first Sorceress scenario run was RED on `town portal did not appear` with `area=1` (the town): after two failed clicks on a stale portal (`back to field: portal click failed ... still 1`) `BackToField` fell back to `fallback=newportal` and cast a scroll IN TOWN. game.exe opens no portal in a town (V fork source `components/game/world/travel.cpp:363`, FUN_005be290; d2d returns before the cast), so that branch can never work; it was never verified live (task-11). It is not a loose-scroll problem: toolQty was 1 because the tome had 0 and one scroll was left. Fix: the branch is removed, the fallback goes straight to the waypoint and the walk; `TestBackToFieldStalePortalDroppedNoScrollCast` (Sorceress, Barbarian). U: whether all three sandbox failures were this path (the sandbox lines carry no area; the dump shows a tool of class tsc with qty 1, which fits).

## Not resolved

- No fork finding: d2d merges on pickup (V source). Control run: with `town.looseScrollKeep` 99 the Paladin kept 6 loose scrolls (RED); with the fix the Paladin and Sorceress runs are GREEN.
- A scroll in the pack that could still go into a tome with room is sold (about 50 gold) and the tome refilled by the shop (100): a drag onto the tome would save the difference; no hand action for it exists (U for the cursor click form).

## Checks

- Unit: `TestLooseScrollsCappedBesideTome` (Paladin and Sorceress, TP and ID), `TestLooseScrollsBesideTomeAreSold` (needs the game tables), `TestItemInPackMax`.
- Scenario `town.tp-scroll-hoard` (tome plus 12 loose scrolls given to the pack, CHEATED setup, mechanics only); the control run was the same scenario with the trigger knob `town.looseScrollKeep` 99 (the old behaviour; a temporary copy, not committed). Results (builder long lane, 2026-11-20, CHEATED setup): control RED (Paladin seed 11, 6 loose scrolls left), fix GREEN (Sorceress 3, Paladin 11). Unit runs with the game tables (`-race`, `-count=20` on the town tests) passed on the builder.
