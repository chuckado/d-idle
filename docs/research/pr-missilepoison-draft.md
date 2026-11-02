# PR draft: a monster missile's poison is per frame, not a whole-point El total (skill-sourced rows, trap missiles)

> status: DRAFT held for the user's go (nothing pushed, opened or posted upstream). Local candidate on the private fork remote only: `local/missilepoison-skillrows` (head `71f98145`, one more commit than the first head `70fa479d`, plus a wording commit naming the right trap row; on upstream main 6935cf64). Verdict of the independent re-verifier on `70fa479d`: READY WITH CHANGES (`docs/research/pr-missilerecord-reverify.md`, branch agent/reverifymissile, section 7 and 9); the changes are done here; third check on `71f98145`: READY WITH CHANGES, text only (`docs/research/pr-missilerecord-reverify.md`, "Third check"), done here. A sibling PR, the launch record: [pr-missilerecord-draft.md](pr-missilerecord-draft.md).

Title (first commit subject): `A skill-sourced monster missile poison goes in MonStats::poison_*, per frame (missile_blow)`

## What changes

`missile_blow` (`components/game/ai.hpp`; `andariel_missile` and `missile_from` call it) stored a poison element as `P x ELen >> 8` in `MonStats::elements[0]`. `monster_blow` reads an `elements` entry in the MonStats El convention (roll x 10 a frame for 2 x dur frames), so the stored total was applied about 20 x ELen / 256 times over. game.exe's record (FUN_0064b860 slots +0x34 / +0x38 / +0x3c) is poison per frame in 256ths for ELen frames, which `MonStats::poison_min / poison_max / poison_len` hold and `monster_blow` applies as they are.

The change puts a type-3 element in `poison_*` as it is and leaves every other element as it was (`>> 8` in `elements[0]`, byte-identical to before). The element placement is one helper, `put_missile_element`, which `missile_blow` and the map's trap missiles (`world/objects.cpp`, the same wrong store) now share.

## The x 62 is real on upstream main

V run (the re-verifier's probe on upstream main 6935cf64 against game.exe's record, defender with no resistance; 13,272 poison comparisons on the skill-sourced set, all 13,272 differ):

| row, level | game.exe record | upstream main's hit | ratio to the record max |
|---|---|---|---|
| andypoisonbolt, 1 | 32-64 x 800 (100-200 points) | 1,130-1,520 x 1,600 (7,063-9,500 points) | x 35-48 |
| andypoisonbolt, 8 | | | x 62-66 |
| andarielspray | | | x 18-44 |
| undeadmissile1 | | | x 26-27 |

After the first commit (`70fa479d`, same set): poison compared 13,272, 0 differ; no poison left as an `elements` entry; other elements 3,018 compared, 0 differ; through the real `monster_blow` 12,586 poison hits, 0 differ from the record (misses at the 95 % to-hit cap skipped).

## Siblings

The same unit mistake lived in three other writers; the re-verifier traced them (section 6 of his doc):

- **Trap missiles** (`world/objects.cpp`, the `kTrapMissile` path): `pts()` stored `P x ELen >> 8`, the same store; the trap d2d fires is `kTrapMissile[3]` = primepoisoncloud (`rules/shrines.hpp`; no trap entry names trappoisonjavcloud), ELen 300 / 360 / 440 at trap levels 1 / 4 / 8, so the old store was x 23 / 28 / 34 too strong. Folded in here through `put_missile_element`. `test_missile_blow_poison` covers the helper (a trap-shaped element), not `spring_trap` itself, which no test runs (its call is V read). V run (the re-verifier's third check): game.exe's row-path record (FUN_0064b860, zero owner) against `row_damage` + the helper for primepoisoncloud, trappoisonjavcloud, trapfirebolt and trapnova at levels 1..40 / 50 / 60 / 99: poison 86 compared, 0 differ; other elements 86, 0 differ; hits through `monster_blow` 80, 0 differ. primepoisoncloud is 10 x 300 at level 1, 25 x 360 at level 4, 45 x 440 at level 8. U: that game.exe's object-trap launcher reaches that row path with nothing added (the oracle uses a zero-stat monster owner, not a trap object); U whether a cloud missile hits more than once.
- **Death clouds** (`fight/deaths.cpp:135`, `boss_missile`, the mummies' `corpsepoisoncloud`): `elo / 10` with `dur = elen`: the rate is right to within 9 / 256 but the length doubles (2 x ELen frames), x 2 in all. Not fixed here (a different file and shape, the damage is only x 2): named so the review can ask for it as a follow-up.
- A monster melee skill's added element (`ai.cpp:434`, `elo >> 8` for any type): a poison skill there would be off by the same convention. U whether any monster attack skill has EType poison.

## Proof

`tests/game/test_missile_blow_poison.cpp` (no game data; the row is typed in from Missiles.txt 1.14d): Andariel's poison bolt at level 1 through the real `monster_blow` gives 32-64 a frame for 800 frames (100-200 points, not 8,812); fire stays an El row; a trap-shaped poison element lands in `poison_*` (the helper, not `spring_trap`).

Checks (V, run on the worker host): `pr_guard` 0 FAIL / 0 WARN against 6935cf64; `presubmit --quick` 0 FAIL, 3 WARN (answered below), short names ok; ctest `test_game`, `test_combat`, `test_game_rules`, `test_rules`, `test_monster_ai`, `test_monsters` and `test_missile_blow_poison` 7 of 7 pass on `71f98145`. The re-verifier's differentials on the head (his harness unchanged, the author's runs): the skill set (18,312 cases; 13,272 poison compared, 3,018 other elements) 0 differences, hits through the real `monster_blow` 0 differences, the dot set (5,926 in-domain) 0.

Presubmit WARNs, answered: *siblings, hit paths*: the monster skill-sourced path (`andariel_missile` / `missile_from`) and the trap path are fixed here; the death clouds and the monster melee skill element are named above; the player's own skill missiles use `rules::missile_blow` (combat.hpp), a different function with its own El convention, not touched; *docs-sync*: no doc describes this poison (`docs/research/re/skills.md` line 671 names only the player's `rules::missile_blow`), so no `.md` changes; *build*: answered by the ctest runs.

## Limits

- Not run: an MSVC build, a live game, the trap launcher in game.exe (the trap record arithmetic is V run against game.exe's row path, see Siblings; the launcher is U).
- The poison roll in `monster_blow` is `rng.range(min, max)`, inclusive; game.exe's FUN_0057a880 rolls `min + rand(max - min)` and never reaches the max (low, pre-existing, not this PR).

## Merge order with the launch-record PR

Both PRs edit `components/game/ai.hpp` a few lines apart (`missile_blow` here, `monster_shot_stats` there); `git merge-tree` of the two heads is clean. Whichever merges second should re-run `test_monster_missile` and `test_missile_blow_poison`. Done: `git merge` of `1824870b` (the launch-record head) and `71f98145` is clean, and `test_monster_missile` and `test_missile_blow_poison` both pass on the merged tree (a throwaway worktree on the worker host); each passes on its own head.

## Files

`components/game/ai.hpp`, `components/game/world/objects.cpp`, `tests/game/test_missile_blow_poison.cpp`.
