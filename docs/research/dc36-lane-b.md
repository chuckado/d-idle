# decomp36 (lane B): death penalty, gamble price, MonUMod eligibility: game.exe 1.14d vs d2d vs the bot

> status: current (2026-11-16, card decomp36, agent d2-decomp on Sonnet). Offline: no game, no d2d slot, nothing posted, no Go changed. The game.exe side ran under the emulator and the d2d side ran as fork builds, both on the builder container (nothing on Orion, nothing heavy on herdr). Tags: **V run** (game.exe's own code under the emulator, compared byte for byte with a d2d build of the real code; the numbers are counts of cases), **V read** (disassembly read, `FUN_` addresses of 1.14d `Game.exe`), **D** (patched 1.14d tables as game.exe loaded them), **S** (an earlier proof of ours or upstream's, named), **I** inferred, **U** unverified. Scratch tooling here: `tools/emu/dc36_death.py`, `dc36_gamble.py`, `dc36_umod.py`, `dc36_cmp.py` (the fork branches carry the upstream-style copies, `death_penalty.py`, `gamble_price.py`, `umod_can_take.py`).

## Answer in eight lines

1. **Picked (lane B scope, impact x tractability):** the death penalty (gold toll and experience loss, `FUN_00580ec0` -> `FUN_00535ab0`), the gamble price (`FUN_0062efb0` mode 2 -> `FUN_00629370`), and which MonUMod rows a champion or unique may roll (`FUN_005a03e0`). Each had no game.exe run behind it (upstream: no `tools/emu` hit for `5357d0`, `5359f0`, `629370`, `5a03e0`).
2. **Death penalty:** the purse toll, its floor, the pile left on the ground, the experience after, the stash and the goldlost stat equal game.exe in 240,000 cases (two seeds). One deviation: the experience the corpse is told it lost is 1 too small when the share reaches the level's start (parent 31,792 and 31,750 of 120,000 differ, claim 0). Effect: at most 1 point of corpse experience. Branch `fid/death-corpse-exp-r1`.
3. **Gamble price:** d2d prices an upgraded (exceptional or elite) gamble base by its own record; game.exe prices it by its normal version (`FUN_006287d0`, the `normcode`). 48,310 of 73,062 (code x clvl 1..99 x reduced 0/10) differ on the parent, 0 on the claim; at clvl 45 an elite Hand Axe costs 51,186 in d2d and 12,933 in game.exe. Branch `fid/gamble-price-r1`. This is the one with an economy effect.
4. **MonUMod eligibility:** 11 of 31,562 (monster row x mod row) pairs differ, all the Multiple Shots mod on the 11 `nomultishot` monsters (the vampires 1..9, slinger5, slinger6); d2d never reads that column. 0 after. Branch `fid/umod-can-take-r1`. Effect: Acts 3 to 5 uniques only (U), none in Acts 1 and 2.
5. **Clean:** gold toll rule, floor, single-player no-stash-tax, XP loss percentages (0/5/10), the XP floor `low + 1`, the exp after, the pile; every other (monster, mod) pair of the eligibility rule (31,551).
6. **Not run:** the corpse's 75 percent step (`FUN_0057f700`, V read only: d2d differs above 1,048,576 points of experience, clvl 60 and up), the gamble stock roll (`FUN_00578790`, no oracle), the non-single-player branches of the death gold rule (not modelled by d2d).
7. **PR state:** three local branches on the private fork remote, none posted. Verifier-ready in this order of value: gamble price, MonUMod, death corpse (trivial effect; bundle it into another death or corpse PR or leave it).
8. Topics I considered and dropped are under "Dead ends and why".

## 1. Method

- game.exe side: the emulator of `docs/emu-image.md` on the builder (`skill_tips.boot()` loads the patched tables, then the real functions are called with a fake game record and unit; hooks only where named). d2d side: a new test file per topic in the fork (`tests/game/test_death_penalty.cpp`, `test_gamble_price.cpp`, `test_umod_can_take.cpp`, in the upstream style of `test_load_gold.cpp`: an oracle mode that prints one line per case), built on the builder (`-j3`, `-j6` was killed by memory).
- Branch shape (upstream pipeline S2): commit 1 = oracle (script + test mode + README row, builds alone and prints what the parent does), commit 2 = claim. Identity is the scratch identity for now; the author line must be changed to the user's own noreply identity before any push to the GitHub fork copy (not done, not allowed here).
- Compare with `python3 -I tools/emu/dc36_cmp.py GAME D2D [column names]`.

## 2. Q1: the death penalty (`FUN_00580ec0` -> `FUN_00535ab0`)

### Rule (V read, then V run)

- `FUN_00580ec0` (player death mode, `0x580f51`) calls `FUN_00535ab0(game, player, killer)`. That calls `FUN_005357d0` (gold), then `FUN_005359f0` (experience) unless the killer is a player or a monster owned by a player (`0x535ac4..0x535ae5`). A monster without an owner or no killer: both run.
- **Gold** (`FUN_005357d0`): L = stat 12, purse g = stat 14, stash b = stat 15. toll = `min(L, 20) x (g + b) / 100` (`0x5357fb..0x53582f`, the `0x51eb851f` divide). Game type byte `game + 0x6a == 3` (single player or TCP host, `net-join.md`): if `(g + b) - toll < L x 500` the loss is `(g + b) - L x 500`, never below 0 (`0x535887..0x535893`); the loss is then capped at g (`0x535895..0x53589e`). `g - loss` falls as one pile (`FUN_00535510`, nothing when 0) and the purse is set to 0; stat 175 (goldlost) = loss. The stash is untouched in this branch. Other game types also tax the stash (`0x535962..0x535994`) and a player killer takes a different path: d2d models neither (single player only, so not a deviation).
- **Experience** (`FUN_005359f0`): row = `FUN_00611d30(game + 0x6d)` (DifficultyLevels, +4 = 0 / 5 / 10 %); L must exceed 1; low = `FUN_00611800(class, L - 1)`, high = `FUN_00611800(class, L)`; loss = `pct x (high - low) / 100`; zero loss stops. new = exp - loss. If new (unsigned) is above low: exp := new and the corpse is told `loss`. Else (`0x535a63..0x535a6f`): exp := low + 1 and the corpse is told `exp - low` (not `exp - low - 1`). The told value goes to `FUN_005391e0` on the record `FUN_005531c0` returns.
- **Corpse** (`FUN_0057f700`, `0x57f80d..0x57f86c`): the corpse's experience stat is the told value x 75 / 100 for values up to `0x100000`, `(told / 100) x 75` above (V read only).

### What ran (V run)

`tools/emu/dc36_death.py game` calls `FUN_00535ab0` for real; hooked: `FUN_00535510` (records the pile), `FUN_0058f0d0` (no owner), `FUN_005531c0` / `FUN_005391e0` (the corpse record). 120,000 random cases per seed (class 0..6, difficulty 0..2, level 1..99, experience from the level start to just under the next, with edge picks at start, start + 1, start + loss -1/0/1/2; purse and stash at 0, 1, 99..101, floor +-1, cap/2, cap, random) on seeds 1 and 7; killer none and killer a monster without owner agree on 2,000 cases. d2d: `World::death_penalty` (`components/game/world/travel.cpp`) after `World::enter`, game type is single player by construction.

| column | parent (seed 1 / seed 7) | claim |
|---|---|---|
| purse after, stash after, experience after, goldlost, pile | 0 / 0 differ | 0 |
| experience the corpse is told | 31,792 / 31,750 differ (every one is the clamped case; game = d2d + 1) | 0 |

An exp below the level's start is not a reachable state (game wraps it; excluded from the grid).

## 3. Q2: the gamble price (`FUN_0062efb0` mode 2)

### Rule (V read, V run)

- `FUN_0062efb0` mode 2 (`0x62f062..0x62f09a`): clvl = stat 12 and reduced prices = stat 0x57 (capped 99) of the player; when the item's data word `+0x30` is at least 1 (`FUN_0062a670`) the price is `FUN_00629370(item, clvl)` minus the reduced percentage (`FUN_00483360`); with 0 it is the base's `+0xd4` column (U what `+0x30` is; the probe shows no dependence on quality, identified, ethereal, NPC or difficulty).
- `FUN_00629370`: the base is `FUN_006287d0(item)` = the Items record's **normcode** when non-zero, else its code (`0x629379..0x62938b`); rings (`rin`) and amulets (`amu`) cost their `+0xd4`; every other base mixes its own cost, the exceptional and elite upgrade costs by `pb = max(0, (clvl - exc level) x 100 / 2 + 1)`, `pu = max(0, (clvl - elite level) x 100 / 4 + 1)` and scales with clvl (formula of `docs/research/re/store.md`, equal in d2d).
- d2d: `rules::gamble_price` (`components/rules/rules.hpp`) looked the code up as written. A gambled **upgraded** base is stock (`gamble_upgrade_code`, `FUN_005786a0`); its code is the exceptional or elite one, so d2d priced it from that record's own cost and level.

### What ran (V run)

`tools/emu/dc36_gamble.py game` calls `FUN_0062efb0(mode 2)` on the price oracle's rig with the player's clvl and reduced stats set. Cases: every code a gamble tab can hold (gamble.txt's 125 plus the exceptional and elite codes their `+0x88` / `+0x8c` name = 369 codes) x clvl 1..99 x reduced 0 and 10 = 73,062. Excluded: `elx` (not a gamble base; game prices it negative, a table oddity, 18 lines in the full 130,482-line grid).

| | cases | differ |
|---|---|---|
| parent (`gamble_price` by the stock code) | 73,062 | 48,310 |
| claim (by the normal code) | 73,062 | 0 |

Examples at clvl 45, reduced 0 (game / parent d2d): `7ha` (elite version of `hax`) 12,933 / 51,186; `7gi` 16,920 / 73,380; `9ha` (exceptional version of `hax`) 12,933 / 11,033. At clvl 30: `9ha` 6,896 / 5,493. d2d is wrong in both directions (up to 4x too dear, ratios of 0.0013 at clvl 1: `9ha` 771 / 1).

Not run: the stock roll itself (`FUN_00578790`: 14 items, level `clvl - 5 + seed % 10`, the gamble table order, upgrade weights, quality `rand(100000)`, grid fit). The earlier docs treat it as read (S, store.md); an oracle needs the vendor record and the item maker `FUN_00559ce0` (drops.py has a harness for the maker). Left as follow-up F1.

## 4. Q3: which MonUMod rows may a monster roll (`FUN_005a03e0`)

### Rule (V read, V run)

`FUN_005a03e0(row, expansion)` with the monster unit in EBX (`FUN_005a0500` champion pick and `FUN_005a0600` unique pick call it for each weighted row): refuse a disabled row (`+6`), a row `>= 100` in a classic game (`+4`), a monster of an excluded type (`+0xa` / `+0xc` through `FUN_005a0070`, the type matrix), and by fPick (`+9`): 1 `FUN_0046c140` flag 4 (swing), 3 `FUN_0046c140` flag 2 (walk), 2 (Multiple Shots, `0x5a046e..0x5a04c5`) a monster whose MonStats flag `+0xc & 2` (isMelee) **or** `+0xe & 2` (nomultishot) is set.

### What ran (V run)

`tools/emu/dc36_umod.py game`: all 734 MonStats rows x 43 MonUMod rows = 31,562 calls of the real `FUN_005a03e0` (expansion game). d2d: `rules::unique_detail::can_take` on `data->umods.rows` x `data->monsters.types`.

| | pairs | differ |
|---|---|---|
| parent | 31,562 | 11 (mod 29 Multiple Shots on vampire1..9, slinger5, slinger6; game refuses, d2d allows) |
| claim (`nomultishot` column read) | 31,562 | 0 |

Per D2 the ids come from d2d's own row names (`--monster-ids`); that `+0xe & 2` is the `nomultishot` column is I from the equal count (11 of 11 names), not a separate read. Weight: upick 6 of the unique total, so a unique vampire's mod rolls change: the weights of the other mods shift (the total is smaller) and the draw uses a different range (RNG-visible).

## 5. Deviations d2d vs game.exe (D1..)

| id | game.exe | d2d | tag | differential |
|---|---|---|---|---|
| D1 | the corpse is told `exp - low` when the experience share reaches the level's start (`0x535a63..0x535a6f`); exp := low + 1 | told `exp - (low + 1)`: one less | V run, both sides (120,000 x 2 seeds) | yes, exists: `death_penalty.py` + `test_death_penalty`, parent 31,792 / 31,750, claim 0. Value: at most 1 point of corpse experience (the corpse gives 75 percent), band = within 5 / 10 percent of a level span after the start, Nightmare and Hell only (Normal loses nothing) |
| D2 | corpse experience is `(told / 100) x 75` above `0x100000` (`0x57f81c..0x57f86c`) | `told x 75 / 100` always | V read (game), V read (d2d) | not built: `FUN_0057f700` is a whole-corpse function; an oracle would run it with the corpse inventory hooked. Reachable only from clvl 60 (Hell, share above 1,048,576 at level 60 for 10 percent; Nightmare level 68), far past the bot's horizon |
| D3 | gamble price uses the normal version of an upgraded base (normcode) | uses the upgraded record | V run (73,062 cases) | yes, exists: `gamble_price.py` + `test_gamble_price`, parent 48,310 differ, claim 0 |
| D4 | Multiple Shots mod refused on `nomultishot` monsters (vampire1..9, slinger5/6) | allowed | V run (31,562 pairs) | yes, exists: `umod_can_take.py` + `test_umod_can_take`, parent 11 differ, claim 0 |
| D5 | non-single-player death also taxes the stash overflow and a killing player takes the toll (`0x535962..0x535994`, `535ab0` player branch) | single player only | V read | not needed (d2d runs single-player and TCP/IP-host games only, game +0x6a 3, as `components/game/world/quests/act4.cpp` notes) |
| D6 | gamble stock roll (`FUN_00578790`, `FUN_005786a0`) | `open_gamble`, `gamble_upgrade_code` | S (store.md, read), not run | buildable: run `FUN_00578790` with the vendor record and the maker; compare the 14 codes, qualities and levels; F1 |

## 6. Bot findings (B1..), described, NOT changed

| id | file:line | finding |
|---|---|---|
| B1 | `internal/bot/death.go:397` | `CostGold = max(goldBefore - gold, 0)` is the purse before minus the purse after the respawn. The game empties the purse and drops `purse - loss` as a pile at the corpse owned by the player (`FUN_00535510`); the real cost of a death is the toll `min(clvl, 20) percent of (purse + stash)`, limited by the floor `clvl x 500` (V run, section 2). So `CostGold` overstates the cost while the pile is recoverable, and the death log (`internal/bot/deathlog.go:41`, `cmd/d2soak/deathrule.go:590` "gold lost ... carried at deaths") mixes the two. Whether the bot walks to the pile is U (no grep hit for a pile pickup after respawn). |
| B2 | `internal/town/npc.go:339` | the town table has a Gamble entry, but nothing prices a gamble (grep: no gamble price model in `internal/decide`, `internal/loot`). Any gold plan for gambling must use the game.exe rule (section 3), not d2d's display (D3) or the base cost. |
| B3 | same entry | if the bot ever reads a gamble price from d2d in a soak, items bought above the exceptional threshold are mispriced by the D3 amount until the fork carries the claim. |

No change made.

## 7. PR candidates

| branch (private fork remote) | head | commits | proof | verifier |
|---|---|---|---|---|
| `fid/gamble-price-r1` | `eba73ede` on upstream `b6e6940d` | oracle `373a2cf4`, claim `eba73ede` | V run, 73,062 cases: parent 48,310 differ, claim 0; the test builds on both commits; claim test passes | not yet; body [gamble-price-body.md](../upstream/bodies/gamble-price-body.md) |
| `fid/umod-can-take-r1` | `6d1fc304` | oracle `21d235d0`, claim `6d1fc304` | V run, 31,562 pairs: parent 11 differ, claim 0 | not yet; body [umod-multishot-body.md](../upstream/bodies/umod-multishot-body.md) |
| `fid/death-corpse-exp-r1` | `c7746e98` | oracle `787628ff`, claim `c7746e98` | V run, 120,000 cases x 2 seeds: parent 31,792 / 31,750 differ, claim 0 | not yet; body [death-corpse-exp-body.md](../upstream/bodies/death-corpse-exp-body.md). Value: trivial; hold it for a PR that also fixes the corpse's 75 percent step (D2) |

Not run for any of the three: `tools/presubmit.py --quick`, `pr_guard.py`, the full `ctest` (the verifier does these on the worker host); "the test fails on the parent" was shown by the oracle differential, not by the claim's own assert run on the parent code. The author and committer identity on the branches is the scratch one (see Method) and must be replaced before the user's go; no GitHub fork copy has been pushed. Posted: nothing.

## 8. Drill recipe (fork verbs, run later in the lab; nothing here ran a game)

Verbs of `docs/forkdbg.md`, `docs/forkcheats.md`, lab binary `docs/d2d-fork.md`. A game is needed only for the U items.

1. **Death toll in the lab binary (confirms D1 and the pile):** a throwaway Nightmare character; `debug stat 12 <clvl>`, `debug stat 13 <exp>`, `debug stat 14 <purse>`, `debug stat 15 <stash>` as declared setup writes before the run start (`docs/cheats.md`), then `debug die`; read `dump stats` (gold, stash, goldlost 175), `debug corpses` (the corpse's experience; `internal/d2dctl/verbs.go` `Die` / `Corpses`) and the gold pile on the ground (the ground listing of `docs/d2d-fork.md`; verb name U). Expect: purse 0, stash unchanged, goldlost = loss, pile = purse - loss; for an experience within 5 percent of the level start, the corpse experience 75 percent of `exp - start`. Results of setup-written characters test the mechanism only (provenance rule).
2. **Gamble price (confirms D3 with the lab binary):** at Gheed (the town verbs of `internal/d2dctl/town_verbs.go`), open the gamble tab, read the offered codes and prices from the store dump (verb name U) and compare an exceptional or elite code with the V run table of section 3; the lab binary shows the parent's number unless it carries the claim. The gamble drill is still a bot gap (`docs/forkcheats.md`, D8.g).
3. **MonUMod (D4):** `debug spawn vampire1 1 unique` (`docs/forkcheats.md`: `debug spawn <monster> [n] [near] [champion|unique [mods]]`) many times without a mod list over several `--seed` values; read the mods in `monsters`. game.exe never rolls Multiple Shots (mod 29) on these rows; the parent shows it on about one unique in a few (weight 6 of the pick total; U the rate). The unit test and the oracle are the proof; this only shows it in a live d2d.
4. **F1 (gamble stock):** the emulator side: call `FUN_00578790` with a vendor record (gamble table sorted by `FUN_00638ae0`'s order), the proxy seed and `FUN_00559ce0` real; compare the 14 stock lines with `rules::open_gamble` over 100,000 seeds and clvls.

## 9. Matrix cell updates (proposed text; the owners apply them)

- **PR08** (death penalty): "UPDATE 2026-11-16 (decomp36, dc36-lane-b.md): V run, game.exe's `FUN_00535ab0` against d2d `death_penalty` in 240,000 cases: purse toll, floor, pile, stash, goldlost and the experience after equal; the experience the corpse is told is 1 low when the share reaches the level's start (D1, trivial). Corpse step 75 percent over 1,048,576 points is V read only (D2). Branch `fid/death-corpse-exp-r1`."
- **VN04** (stock lists): add "gamble price: 48,310 of 73,062 differ on an upgraded base (D3, V run, `fid/gamble-price-r1`); the stock roll itself is not run (F1)".
- **IT05 / MA06**: "MonUMod eligibility FUN_005a03e0 compared on 31,562 pairs: only Multiple Shots on the 11 nomultishot rows differs (D4, V run, `fid/umod-can-take-r1`)."
- Pipeline (docs/upstream/pipeline-2026-11-14.md section 1b): three new rows P20 (gamble price, rank above P05), P21 (MonUMod nomultishot), P22 (death corpse experience, trivial).

## 10. Dead ends and why (topics considered and dropped)

- **Equip legality and requirements (IT03):** already run: `covitemtypes-trace.md`, 553,647 cases including `FUN_0062ed50` slot fit and `FUN_0063dbc0` two hands, 0 differ.
- **Drops, quality, affixes (IT05):** upstream's `drops.py` runs 100,000 / 100,000 jobs, 39,980 of 39,980 made items, containers 20,000, stands 20,000; the "Not ours yet" list is the gap, not a new trace.
- **Vendor stock, prices, repair, identify:** `fid-stock-oracle.md` (40,800 lines, 0 differ), `price-trace.md`, `fid-vendor-price-trace.md`, `fid-identify.md`; gamble price was the one mode of `FUN_0062efb0` left.
- **Quests, waypoints, town portal, shrines, warp arrival:** upstream's `tools/emu/quest_*.py` family, `waypoint-trace.md`, `town-portal-trace.md`, `shrine-trace.md` (the game side there is S and the image is rebuilt now: a rerun of shrines is the best follow-up in the town and world lane), `fid-warp-arrival.md`.
- **Monster AI beyond Fallen flee:** `monsterai2-trace.md`, `covmonai2-trace.md`, `decomp30-mercai.md`; the open items there (Teleportation AI 40 / 15 percent, Lightning Enchanted bolt cadence) were not rerun.
- **U kept:** the meaning of item data `+0x30` in mode 2 (section 3), the Acts where the 11 `nomultishot` rows spawn, whether the bot walks to the death pile (B1), the share of gamble stock items that are upgraded at a given clvl.

## Shared-file lines for the merger

BACKLOG / CHANGELOG / INDEX lines are generated from the task card (TASK-558); `docs/upstream/pipeline-2026-11-14.md` rows P20..P22 and the matrix cells above are for the owners.
