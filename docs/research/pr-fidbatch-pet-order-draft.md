# Draft (local, NOT posted): "fight: a monster tries a player's newest summon first and its merc last (FUN_005dd7f0)"

> status: draft 2026-11-23, agent fidbatch (bundle brief: turn the open fidelity cards into proven upstream candidates). One claim. Nothing is pushed to GitHub or posted; the private fork holds the branches. Source trace: [pets-merc-trace.md](pets-merc-trace.md) (D1, 2026-11-01); this doc re-proves it on upstream main `ff6d47a66` and builds the fix.

## What the open cards turned out to be (checked 2026-11-23 against upstream main and `gh`)

| Card / item | Verdict | Evidence |
|---|---|---|
| TASK-101 fid-load-life-mana | already upstream, MERGED | psi29a/d2decomped #242 (2026-11-08) |
| TASK-125 fid-revive-refill | already upstream, MERGED | #245 (2026-11-08) |
| TASK-203 fid-load-gold-clamp | already upstream, MERGED | #539 "a saved purse or stash over its cap is 0 on entering a game" (2026-11-15) |
| TASK-102 fid-load-maxmana | NO DEVIATION in d2d (bot reader defect) | task result of 2026-11-08, `fid-loadmaxmana-trace.md`; nothing to propose |
| TASK-179 ptrclick | UNVERIFIED, no PR candidate | needs a live probe (docs/probes/ptrclick.md); a click verb artifact of the lab fork, not an upstream d2d behaviour |
| monster missile gates (can-block, MissA ToHit) | already upstream | #553 (can-block); ToHit gate: `stats.no_hit_roll = !info.to_hit` on every monster missile path (ai.cpp:3480, 3510, deaths.cpp:357, objects.cpp:1035) |
| monster mods D1 D2 G1 | already upstream | #157 (boss stats), ghostly speed skipped (`components/rules/uniques.cpp:198`), Cold Enchanted length (`uniques.hpp` elem_len) |
| pet tie order D1 | OPEN on upstream main; this draft | below |

Open PR #708 (merc RogueMissile) is not touched.

## Title and body (PR template)

Title: `fight: a monster tries a player's newest summon first and its merc last (FUN_005dd7f0)`

> **What changes.** A hostile monster's target search (FUN_005dd7f0) walks the player's list in order and takes a candidate only when it is strictly nearer than the best so far, so on a tie the first one tried wins. game.exe builds that list with FUN_005b1900, which inserts each pet right after its player (0x5b1966..0x5b1979): the newest summon is tried first and the merc, hired before any summon, last. d2d tried player, merc, summons oldest first. `rules::search_order` now gives the order and `search_target` uses it.
> **Evidence.** `tools/emu/pets_merc.py` builds the list with game.exe's own FUN_005b1880 / FUN_005b1900 and runs FUN_005dd7f0 natively on 100,000 cases (seeds 21 to 24, one player, the merc 85 % of cases, 0 to 3 summons, distances biased to the ties and to the 0x37 / aidist / 6 / 0x13 / 0x31 edges); `pets_merc_harness.cpp` runs d2d's real `rules::search_order` and `rules::search_pick` on the same cases: 1,254 of 100,000 differ with the old order, 0 with the new. `tests/game/test_pet_search_order.cpp` drives the real `Fight::world`: it fails on the parent (3 of 5 checks) and passes with the change.

Claims table (one commit):

| Claim | Evidence | Check |
|---|---|---|
| FUN_005b1900 inserts a pet right after its player, so the list runs newest summon .. oldest summon, merc (hired first) | oracle builds the list by calling FUN_005b1880 then FUN_005b1900 per pet (V run) | `pets_merc.py`, README row |
| FUN_005dd7f0 takes a candidate only when strictly nearer (first tried wins a tie) | same oracle; the breakdown of the 1,254 (all equal-distance ties between pets, 780 where d2d picked the merc and game.exe a newer summon, none involving the player) is from the 2026-11-01 run on the same case file (identical md5), not re-tabulated today (I) | `./pm_h before` vs `./pm_h after` |
| d2d now does the same | `search_order` + `test_pet_search_order` | `test_pet_search_order` (needs the game data) |

Limits (to be in the PR body):
- One player. With two players in the foes vector (co-op) only the entries after the last player are reordered (d2d already lists every pet after the last player, a separate quirk); not exercised.
- The merc is treated as the oldest pet. game.exe re-inserts it when it is hired, loaded or revived, so after a merc revive (or a hire after a summon) it is the newest in game.exe; d2d keeps no insertion order for the merc. U: not decoded further, outside the claim.
- A converted monster (Conversion) is a list pet in game.exe too (newest first by its conversion time); d2d keeps it where it was. Outside the claim.
- The harness restates `search_target`'s input mapping per case line (as `search.py` does); the in-game test covers the real call site only for the cases it builds (ties between summons, merc against a summon, a nearer merc, a nearer older summon).
- Dead pets are not generated (game.exe removes a dying monster from its list, FUN_005a6520 -> FUN_005b1a90).
- Only the monster's search is changed; the merc's own target search (FUN_005ddc30, 0 of 100,000 differ) and the pets' thinks are not.
- Hooked in the oracle (not the claim): the room, outdoors, town, sight, in-melee, alignment, size, area, FUN_005dd510's pathfinder, the pool allocator (list in the oracle docstring).

## Reproduce

The worker's builder, upstream checkout with this commit (fork branch `didle/fidbatch-pet-search-order-up` on the private remote, commit `21df4aa6`, one commit on upstream main `ff6d47a66`), game.exe as `$D2_MPQ_DIR/bin/game.exe`:

```sh
cd tools/emu
for s in 21 22 23 24; do uv run python pets_merc.py 25000 $s c$s.txt > g$s.txt; done
cat c2?.txt > c.txt; cat g2?.txt > g.txt
g++ -O2 -std=c++23 -I../../components $(find ../../components -name '*.hpp' -printf '-I%h\n' | sort -u) \
    pets_merc_harness.cpp ../../build/components/rules/librules.a -o pm_h
./pm_h before < c.txt | diff g.txt - | grep -c '^>'    # 1254
./pm_h after  < c.txt | diff g.txt -                   # nothing
# md5 prefixes: cases b4a2be90, game 168c5cf7
build/tests/test_pet_search_order                       # parent: 3 FAIL lines; this commit: OK
```

## Checks run (V run unless said)

- Oracle: 200,000 case lines (S and M, seeds 21..24): `before` 1,254 differ, `after` 0. md5 of the cases and the game side equal those of 2026-11-01 (b4a2be90, 168c5cf7): the game side is stable.
- `test_pet_search_order`: parent (upstream main + the test file only) prints three FAIL lines (took foe 2 / 2 / 1, game.exe's order takes 3 / 4 / 2); the commit prints OK.
- `python3 tools/pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN. `python3 tools/presubmit.py --quick --base upstream/main`: the only FAILs left are `tools/emu/test_structs.py` (a `structs.tsv` row cites `ai.hpp:509`, the file has 371 lines; the diff does not touch ai.hpp; pre-existing on main, V read) and a short-name hit fixed before the commit. WARN: siblings (co-op partner: not exercised, stated in Limits), doc-claims on the README row (the row says "no output", not "every").
- Tests of the one-commit tree (upstream main + the claim, GCC Release, game data): `test_pet_search_order`, `test_combat`, `test_monster_ai`, `test_rules`, `test_monsters`, `test_game_rules`, `test_game` all rc 0 (V run). Not run: the other test binaries (the whole set is 1,250 targets, about two hours on the builder lane), `ctest`, `-Wshadow -Werror`, clang-tidy, the det state hashes (tools/d2ddet) for the claimed change: the pick only differs with two or more summons or merc plus summon at equal distance, so the det scenes without a Necromancer or Druid should be identical (I, not run).

## Not verified / independent verification

An independent verifier (d2-reviewer, own harness, own case generator) has NOT looked at it. Checklist for the verifier: (1) decode FUN_005b1900 (0x5b1966..0x5b1979) and FUN_005b1880 yourself and confirm insert-after-head; (2) confirm FUN_005dd7f0's compare is strict `<` (first tried wins); (3) run your own S cases with a merc re-inserted after the summons (the revive case) and say how d2d would need to track it; (4) look for another reader of the foes order (`foes[side.at + 1]` is the merc: unchanged; only `picks` is reordered); (5) co-op with two players.
