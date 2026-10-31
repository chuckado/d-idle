# Primitives review (belt, pickup, consumables, NPC, input/state)

> status: read-only review of 2026-10-31 (card primreview). Code at origin/main 571224d; logs on the worker host:
> navsoak, navsoak2, navsoak3 (still running, read by streaming), m2early (8 seeds), goldab (16 runs). Nothing changed in code.

Counts are from `grep`/python streaming over the JSON logs; "navsoak3" numbers are a snapshot of a growing file.
V = verified by reading the code and the logs; I = inferred (code read, no log line proves it).

## Verdict on the hypothesis: PARTIAL

- **Refuted: "the pack-to-belt move fails" and "fill runs only in town".** `cmd belt put` works on d2d. All belt-fill
  lines across the logs: navsoak* 361 `belt: moved` lines, 335 with failed=0. m2early 104/107, goldab 132/141 and
  soakfull 75/76 are clean too. `dragged` is always 0: the d2d `BeltPut` path is used, never the drag. The fill also runs in
  the field (`internal/bot/resource.go:179` beltFillStep; navsoak* `belt fill` 254 in the field vs 31 in town).
- **Confirmed, with a different mechanism: potions sit in the pack while the belt is empty.**
  1. Every failure (67 of 67, all logs) is `hand: item not found` within 1 s of a save-and-exit re-entry.
  2. A failure of *any* kind sets the sticky `beltFullAt` (`internal/town/beltfill.go:111-114`). `beltRoom` then returns 0
     until the belt's (slots,total) changes (`beltfill.go:257`). That blocks both the field fill and the town chore.
  3. Town visits whose first shop plan shows beltHP=0 with packHP>0 skipped the "belt" chore 57 of 57 times. Every one
     matches a preceding failed fill with the same belt total. Example from navsoak3: the 09:19:57 failure left the belt at
     0/8. At 09:21:53 and 09:22:06 two trips followed with `pack: hp1 x4, mp1 x4`, belt 0, "belt" skipped and `wantHP 0`. The
     second was an emergency trip `[0 healing potions]` that did nothing.
  4. No belt step runs after the shop (`internal/town/chores.go:267` "belt" comes before "shop" at :270). d2d buys land in the
     pack, so purchases wait for the field fill (90 s backoff, no hostile near: `resource.go:175,181`). Visits that left the
     shop with belt HP 0 and HP in the pack: navsoak3 75/187, navsoak 50/472, m2early 29/145, goldab 41/201.
- **Refuted: the pickit counts the pack but not the belt.** `needsOf.Want` (`internal/loot/loot.go:479-487`) first asks
  the belt plan (belt counts and free slots). It falls back to `PackCount < PotionBuffer` only when the belt is not short.
  95 percent of the "keep [C] potion already in the pack" skips came while the belt held a healing potion (navsoak3
  725/759; navsoak 619/669; m2early 363/382). The skip is mostly correct. Its reason text is misleading (C3).
- **Share of the 31 percent empty-belt minutes** (navsoak3: 664 of 2148 minute lines):
  - At least 137 (21 percent) provably had HP in the pack: the next belt event moved HP from the pack before any buy
    or pickup.
  - 197 fall inside sticky windows, from a failed fill to the next good one. These overlap the 137.
  - 500 (75 percent) ended with a purchase or a pickup, so the belt was really empty. That part is supply and poverty
    (F2, F4, F5), not the move.

## Findings (ranked by impact x proof)

| id | primitive | bug | proof (code + logs) | failing scenario | fix size / where | conf |
|---|---|---|---|---|---|---|
| F1 | belt fill | sticky `beltFullAt` is set on every failure, including `ErrNoItem`, context errors and anything transient, and is never cleared on re-entry | `town/beltfill.go:107-118,257`; 57/57 skipped belt chores with belt 0 and pack HP explained; navsoak3: 12 windows, 197 empty-belt minutes, 30 trips | after a re-entry the fill fails once with belt 0/8. The belt stays empty with 4 hp1 in the pack and futile emergency trips follow | S: set it only for "no free slot" or a drag that did not land; reset it on game entry; give it a TTL. `town/beltfill.go` | V |
| F2 | loot seam | `looterAdapter` does not forward `SetRecovering` or `RecoveryTotals`, so the recovery pickup rules never switch on (every gold pile, sellables over the PoorMinSell bar) | `cmd/d2bot/adapters.go:196-210` vs `bot/broke.go:43-50,99`; `loot/pickup.go:77,324,357`, `loot.go:123`; 0 non-zero `recovery: loot` lines out of ~20k (navsoak 4238, navsoak2 13563, navsoak3 717, m2early 697, goldab 999) | a broke character farms gold with the normal pickit. The recovery summary always prints +0 | S: two forwarding methods on looterAdapter, plus a compile-time assert `_ recoveryLooter = looterAdapter{}`. `cmd/d2bot/adapters.go` | V |
| F3 | town chores | no belt fill after the shop: bought potions stay in the pack until a field tick that is 90 s clear | `town/chores.go:267-272`; 75/187 navsoak3 visits left with belt HP 0 and pack HP; 86 town minutes with belt 0 and pack HP | buy 4 hp1 with an empty belt, walk into the field, the fight starts before the backoff passes and nothing gets drunk | S: run "belt" again after "shop" (the shop already counts pack potions). `town/chores.go` | V |
| F4 | portal (consumable) | `d2dPortal` picks the first Portal item with no charge check, so an empty Tome of Town Portal is clicked 3 times | `cmd/d2bot/d2dwire.go:210-221` vs `combat/fighter.go:536` (`portalQty <= 0` skipped); 939 `no portal after the click` and 246 fallbacks in navsoak3. Last shop visit before a fallback showed tp=0: 81/246 (navsoak3), 206/329 (navsoak), 23/23 (goldab) | tome empty plus a loose scroll later in the item list: 3 failed clicks, walk home | S: use `portalQty`, prefer a scroll over an empty tome. `cmd/d2bot/d2dwire.go` | V (tp=0 cases); fallbacks with tp>0 not explained, I |
| F5 | kit check | `resInput.HasTP` is true for any "Town Portal" item, so an empty tome counts as a portal | `bot/resource.go:57-58`; feeds `kitgate.go:28` and the trip trigger | empty tome: the kit gate and trip logic believe a portal exists and the field run starts without one | S: count tome charges (the town's `scrollCount`). `internal/bot/resource.go` | V code, I impact |
| F6 | kit gate | `kitMissing` reads belt HP+RV only and ignores pack HP waiting to be belted | `bot/kitgate.go:22-29`; 90 `kit gate on` lines in navsoak3 | belt 0 and pack 4 hp1 during an F1 window: the gate narrows the farm list. The fix is to belt the potions, not to treat the kit as missing | S: once F1 and F3 are fixed, either belt first or count PackHP as "kit present, fill pending". `internal/bot/kitgate.go` | V code, I share |
| F7 | belt read (combat) | `BeltOf` takes `GX` as the column and drops `GX >= 4`. On d2d, belt items carry `x` = box 0..15 and `y` = 0 (fork `put_in_belt`/`pick_up` set `column = box`), so rows 1-3 are invisible to combat | `combat/safety.go:33-36`; `d2dsrc/items.go:85-90,122`; potion-skip lines `belt=hp=0` while the minute lines on both sides showed potions>0 with the same drunk counter: navsoak3 30/1534, navsoak 37, navsoak2 8, m2early 8 | an HP potion in box 5 above an MP in box 1 cannot be reached. Combat says "no healing potion" while the minute line counts 1 | M: model the belt as box = GX (d2d) vs (GX,GY) (Game.exe U), column = box%4, front = lowest row. `combat/safety.go`, `town/beltfill.go` (`beltRoom`, `beltCellPoint`) | V |
| F8 | belt fill | `freeBeltBox` takes the lowest free box whatever its column's plan kind or the bottom potion's kind, so HP can end up above MP (feeds F7) | `town/beltfill.go:141-154` (box = GY*4+GX happens to equal d2d's box) | bottom row hp,mp,mp,mp and 2 more hp: they go to boxes 4 and 5, and box 5 is behind an MP | S: choose the box in a column whose bottom matches the kind (the d2d pickup rule), else an empty column. `town/beltfill.go` | V code, I frequency |
| F9 | belt fill | `ErrNoItem` right after re-entry: the candidate id from the fresh snapshot is not found on the next read | 67/67 failures within 1 s of `re-entered the game`; `beltfill.go:170-173` | first tick in town after a save point | M after logging (see below); the fill could also skip the first second after entry. `town/beltfill.go`, `internal/bot/resource.go` | V timing, I cause |
| F10 | equip | the cursor-state check fails on d2d swaps | 7 `hand: unexpected cursor state: cursor never held item N` (navsoak3 leveler equip) | gloves or cap swap retried and failed | M: needs the d2d cursor trace. `internal/town/hand` | V count, I cause |
| F11 | bank | partial gold deposit is impossible on d2d (no typing), so the bank chore fails every time it is due | 13 `town chore failed bank` in navsoak3 | gold over the keep line is never banked | S: deposit all, or skip the chore on d2d with a log line. `internal/town` | V |

## Not bugs (cleared)

- C1 `belt: moved 0 ... dragged 0 failed 2`: the drag never runs on d2d. failed=2 means two `ErrNoItem` in one call (F1/F9), not a full belt.
- C2 `beltCellPoint` and the `beltRoom` grid check use (GX,GY) and are wrong for d2d boxes 4..15. They only matter for the drag fallback, which d2d does not reach. `beltRoom` falls back to `Free()`, which is correct there.
- C3 "keep [C] potion already in the pack" is the catch-all `spare supplies` rule (`loot/pickit_starter.json:25`; also stamina, antidote and thawing potions, whose `PotionKind` is None). Behaviour is fine, the reason text is not: rename it "supply not needed".
- C4 `loot: pack full, item left on the ground` (navsoak3 124): the items are white or low armor and chipped gems at invPct about 85. That is the sell or stash cadence, not a primitive. The pack after town is under 80 percent in every `inventory contents` line read.
- C5 `recovery: loot +0 ... 0 left for lack of room` (716 in navsoak3) is not a loot failure. It is F2's missing forwarding.
- C6 The pickit can over-pick potions while unbelted pack potions wait (Want ignores them while the belt is short). That costs pack cells but is harmless once F1 and F3 are fixed.

## What to log to settle the open questions

1. Belt fill per candidate: `belt: put id=<id> box=<b> kind=<k> try=<n> -> inBelt=<bool> cursor=<id|-> beltBoxes=[...] packIDs=[...] frame=<f>`. This settles F9: a stale id, an id change on move, or the item left on the cursor.
2. `beltFullAt` edges: `belt: marked full sig=<slots,total> reason=<err>` and `belt: full mark cleared (<why>)`. This shows the F1 windows directly.
3. Portal use: `portal: use item=<code> id=<id> qty=<charges> objects<25=<n> -> result=<object|gone|none> panel=<flag>`. This splits F4 (empty tome) from a d2d portal-object read gap.
4. Minute line: add `packHP`, `packMP` and `beltBoxes` (box:code). The 31 percent then splits into "nothing anywhere" and "in the pack or behind another kind" without inference.
5. Pickup: `loot: potion skip code=<c> beltWant=<n> packHave=<n> buffer=<n> free=<n>` (rate-limited). This proves or clears the remaining 5 percent of skips that happened with belt HP 0.
