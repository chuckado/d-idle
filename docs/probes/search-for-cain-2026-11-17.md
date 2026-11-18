# The Search for Cain on the sandbox: diagnosis and fix plan (2026-11-17)

> status: diagnosis, read-only and offline. No game was started and nothing was sent to the sandbox; the only live evidence is `docker logs --tail 30000` of the sandbox container (21:57Z to 23:56Z, deploy 10, Paladin clvl 16, normal). Tags: V verified (code read or log line quoted), S secondary (Kolbot / game knowledge from memory), U unverified.

## 1. What is modelled (V)

`data/questboard/act1.json` `search_for_cain` has FOUR chunks, ten steps: `scroll` (travel_dark_wood, take_scroll; nodes wp-dark-wood, to-wood, tree), `read` (return_town, talk_akara; hand-in Akara), `free_cain` (travel_stony_field, touch_stones, enter_tristram, free_cain; nodes to-stony, stones, to-tristram, clear, cage, home) and `report` (Akara again, handler `search_for_cain_board`). The real chain (tree, scroll, Akara, five stones, Tristram, cage, Cain) is fully present as data. "Step 2 of 2" on the board is the position INSIDE chunk 1 (`scroll` has two steps), not the whole quest. So the chain is not under-modelled; the model is fine. The execution is what fails.

Executors (V): `internal/task/act1.go` `cainNodes` (the step graph with done-evidence), `cmd/d2bot/objects.go` (`objSpecs`: tree class 30, cage class 26, `Tristram portal` class 60, `find`/`findNear`/`click`), `cmd/d2bot/cain.go` (`solveCairn`, `liveStones`), `internal/bot/task.go` `taskActs.Interact` (click tree then `PickUp("Scroll of Inifuss")`, radius 30) and `internal/talk/tables.go` (Akara "bark" and "cain report" rows). Object clicks go through `cmd interact <NPCIDBase+preset index>` (V drill H.obj: chest 139, shrine 81 and the waypoint open this way, `internal/d2dsrc/items.go` `fieldObjects`).

## 2. What the sandbox log shows (V)

Only the first step ever ran. Quoted lines: `task step quest.search_for_cain/tree outcome retry kind not_found err "object class 30 not found"` at 21:58:42, 22:00:23, 22:02:05 (three tries about 100 s apart), then `quest board: parked search_for_cain/scroll: gave up after 3 tries: object class 30 not found` at 22:02:07 (`MaxFailures` 3, `internal/questboard/exec.go`). `minute` rows show `explored` 42, 50, 60, 67, 75, 79 percent of Dark Wood across those tries: the hero was still exploring when each 100 s budget (`objectExploreBudget`, `cmd/d2bot/objects.go`) ended ("context deadline exceeded" on the last nav leg). No tree click, no pickup, no Akara talk, no stones appear anywhere in the window. After the park the hero farms Dark Wood (goal g4) as the board's farm mode; the park is in memory only and needs the player's "Try again". The earlier "no scroll in the pack" is consistent: the scroll was never taken.

## 3. Root causes (ranked)

1. RC1, tree never found (V log): `find` explores for 100 s and gives up as a retryable failure; three failures park the chunk. Dark Wood is large and d2d `debug presets` lists only rooms that are built (`docs/d2d-backend.md` World section, S/V probes doc), so the tree preset does not exist until the hero is near; `fieldObjects` also lists only objects within 40 cells. Exploration progress (42 to 79 percent) is lost to the failure counter: an attempt that still gains coverage counts like a dead end. Also U: whether preset `ObjectID` of the tree is really 30 (S Kolbot `InifussTree`; `docs/quest.md` check 7 still lists it open) and whether `debug presets` names it. The log has no preset dump, so a wrong class cannot be ruled out.
2. RC2, stones cannot be solved on d2d (V code, U live): `fieldObjects` sets `Mode: -1` for every preset row ("open/closed state not in presets", `rp.unknown("object.mode")`). `liveStones.Lit` compares Mode with its first reading, so no stone ever reads lit, `solveCairn` re-clicks the prefix every time and ends with "no stone stays lit" (this is the case `TestSolveCairnGivesUpWhenLitIsUnreadable` pins as an error). The puzzle needs a lit-state source (`debug objects` carries mode under another index, `docs/d2d-backend.md`) or a no-state search.
3. RC3, quest item names disagree (V code, S real): `internal/task/world.go` `ItemCainSc = "Deckard Cain's Scroll"` but `internal/quest/act1.go` calls the code `bkd` "Key to the Cairn Stones", and Kolbot lists the item "Key to the Cairn Stones". Pack scan is by display name (`scanQuestItems`), so after Akara reads the scroll the world sees neither item (`QuestState` Started via `ItemCainSc` never fires). `read` still completes (bark gone) but evidence is thin. `Scroll of Inifuss` is the right bks name (S).
4. RC4, pickup timing (U): `taskActs.Interact` calls `PickUp` immediately after the tree click; the scroll appears after the tree animation. No wait or retry in that path (V code), and it needs a ground row named "Scroll of Inifuss" (`namedItem`, U).
5. RC5, untested downstream (U): `Tristram portal` (class 60 `objLeave`), `Clear Tristram` (12 min), `Cain's Cage` (class 26, `objOnce`), Cain in town (class 245 U) and the report talk were never run live; `docs/probes/drills/README.md` records the cage effect as U ("H.obj.cage effect ... not set by `debug quest set` alone").

Not a cause: the model, Akara's dialogue rows (the "bark" and "cain report" rows exist; Akara rows are V from the other quests), the planner.

## 4. Missing primitives (summary)

| need | exists | gap |
|---|---|---|
| find a far, unbuilt preset | `find` + `nav.Explore(known)` | bounded 100 s, restarts per try, progress not kept, class U |
| click the tree | `cmd interact index` via presets (V for chests) | tree not verified |
| take the scroll | `Loot.PickUp(Names)` | no wait for the drop, name match U |
| scroll in pack evidence | name scan | code/name for `bkd` wrong |
| talk to Akara | `talk` table, hand-in chunk | none known |
| stone order | backtracking `solveCairn` | needs a lit state; d2d gives Mode -1 |
| Tristram entry, cage, Cain | objSpecs rows | never run |

## 5. Fix plan (slices, one PR each)

- CAIN-1 explore-to-find (offline, build-neutral). Owns `cmd/d2bot/objects.go`, `cmd/d2bot/objects_test.go`, `internal/questboard/exec.go` (only the failure rule). Make `find` return a distinct "still exploring" error while coverage grows (the Explore result already knows ErrExplored), and let the board not count it toward `MaxFailures`; keep a long total budget (data: `explore.maxMinutes` per object in `objSpecs`) and park only on ErrExplored (whole level walked) or the budget. Test: fake `nav` and `World` where the preset shows after N legs, plus the park case. Live: CAIN-L1.
- CAIN-2 item codes (offline). Owns `internal/task/world.go`, `internal/bot/task.go` (`scanQuestItems`), `internal/quest/act1.go`. Scan by item code (`bks`, `bkd`) from the catalog instead of display name for these two, fix the `bkd` display name, keep names as aliases. Test with a fake catalog containing both codes; also Paladin and Sorceress packs (build-neutral by construction).
- CAIN-3 tree click and pickup (offline, mem.Fake). Owns `internal/bot/task.go` (`Interact` tree branch), `internal/loot/pickup.go` only if the wait belongs there. After the click wait up to a few seconds for the ground row, retry `PickUp` and verify the pack, else a clear error ("tree clicked, no scroll"). Test with a fake looter that drops the item after 2 polls. Live: CAIN-L1.
- CAIN-4 stones (offline plus fork read). Owns `cmd/d2bot/cain.go`, `cmd/d2bot/cain_test.go`, `internal/d2dsrc/items.go` (mode source). Preferred: read stone mode from `debug objects` for the five stone classes and feed `Lit`; fallback with no mode at all: try permutations in lexicographic order, resetting by waiting, until the portal object appears (at most 120 orders; S rule: a wrong order resets). Test: `stoneOps` fake with a hidden order and (a) readable mode (b) portal-only. Live: CAIN-L2. If the fork needs a mode field in `debug presets` that is a fork change (own task, fork branch, pushed the same day).
- CAIN-5 Tristram, cage, Cain (live-led). Owns `cmd/d2bot/objects.go` (cage, portal rows), `internal/task/act1.go` (`cainNodes` evidence), `internal/talk/tables.go`. Written after CAIN-L2 shows what the portal, the cage and the quest word do. Cain town class 245 is U.
- CAIN-6 board wording (cosmetic). Owns `internal/watch`: show "part 1 of 4: Take scroll" instead of "step 2 of 2", and show a parked chunk as Needs you with a Try again hint. The park already exists; the report said Running.

Build-flexibility: all chain data is data (`data/questboard/act1.json`, `objSpecs`, item codes); no class or skill literal is involved. The only combat-adjacent piece is the Tristram clear (uses the normal fighter profile; check it with one melee and one caster build before it ships).

## 6. Live drills (need a hidden instance)

All on hidden instances, never i0. Suggested: i1 (Navi) for L1 and L2 sequentially, a fresh or genuinely played Act 1 character at clvl 10 or more; any cheated character may only test mechanics (provenance rule). Setup writes only before the start, declared.
- CAIN-L0 (read-only first, 5 minutes): in Dark Wood, dump `debug presets` and `debug objects` while walking toward the west half; record the tree's `ObjectID`, its preset index, how far from the waypoint it appears and the real `explored` percent at which it shows. Settles RC1's class and distance question for good.
- CAIN-L1: find tree, click, scroll in pack (`dump items`), talk Akara, `bkd` name and code in the pack.
- CAIN-L2: stones: record the `Mode` of each stone via `debug objects` before and after clicks, the reset on a wrong order, the portal appearing; then Tristram arrival, cage click, quest word, Cain in the camp.
