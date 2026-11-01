# Boss-stat candidate: verifier pointer, corrected numbers and det state hashes (card modsprep)

> status: current addendum to [pr-mods-proof.md](pr-mods-proof.md) (left untouched); the PR body is [pr-mods-draft-v2.md](pr-mods-draft-v2.md). Tags: V = run by us (how), S = read in a doc, U = unverified. Nothing posted or opened upstream.

## Verifier verdict (pointer)

The independent verifier (card verifymods, branch `agent/verifymods`, section appended to pr-mods-proof.md there, not merged when this was written) re-derived the levels and ran its own 120,000-case differential on head c048a356: D1 narrow form HOLDS WITH LIMITS 143,984 -> 0; D2 HOLDS WITH LIMITS 36,203 -> 0; D3 HOLDS 3,427 -> 0; B HOLDS 7,300 -> 0; each commit removes only its own slice. Verdict: claims hold; NOT READY until the det hashes and four draft edits. The edits are in draft v2; the det hashes are below. Residuals independent of the patch (same before and after): life within 2 points 18,591; rows with no life 4,089; stone skin 327; rows 0x137 / 0x138 313. The -33 % store at 0x5a1062 is dead code.

## Corrected impact numbers

Recomputed from the pack's Act 1 Normal table (pr-mods-proof.md, "Gameplay impact"): the level 10 champion life cell is 1.14, not 1.26 (the unique cell at level 10 stays 1.26). So the champion life range is 1.1-1.7x (not 1.3-1.7x) and "champions and uniques 1.3-2.1x the life" reads 1.1-2.1x. Places that repeat the old range: pr-mods-proof.md ("Read:" paragraph and the level 10 row), pr-mods-draft.md, the boss-stats row of pr-bundle-queue.md, docs/BACKLOG.md (BUILDMODS entry). The draft v2 carries the corrected text; the older files are kept as written (they say the verifier corrects them: this addendum is that correction).

## Det state hashes (V)

Method: the DETHASH process (det-hashes-2026-11-01.md) and the scenario set of the dethash2 branch (`tools/emu/det_scenarios_lab.py`, `det_scenarios_run.py`, read from the worker host copy): a fresh headless d2d per run (seed 3, lockstep, `--start-screen ingame`), scripted `step` blocks, a snapshot after each (monsters, `debug monstate`, player cell, merc, pets); hashes are md5 over the `monsters` replies (mon) and over every snapshot line (state, first 12 hex). One hidden slot (slot 2, `d2dslot.sh run -n 2`) under `agentrun.sh modsprep`; host idle+nice was 37 % at the start. No game.exe.

Builds (V, worker host, scratch `~/d2d-lab/pr/modsprep/`): base = upstream main 72132b99 + the lab commits + the harness `debug monstate` (commit 5e0a51bc of the dethash2 tree); boss = base + `git cherry-pick 72132b99..c048a356` (the five commits, no conflicts; the picked heads are e2e58310 D1, 8e921e49 D2, 406031fc D3, 109235e3 B, 906acd43 tooling). Same two trees as the dethash2 base, so the base hashes cross-check against that run: passive d1ca6c463f45, fight 5f90630302ee, cold bba8362ccde9, static 34a1b9ef9f44, cold2 bba8362ccde9, hit_plain 81b48ba4bdb0, hit_champion 9b7c25f25f6a, hit_unique 04f91fcf36ff, hit_boss d49d11125ccc, merc 1ecd7abc1cd7, pet 99d912dcc2e9, follow 9f51307f5f85, hit_minion 27714ca3de68, hit_immune 2cf37f51eea2 (the mon hash column equals the dethash2 base for all 14). Self-check: base rerun (base2) for hit_boss, hit_unique, hit_champion, hit_minion and passive repeats both hashes.

| scenario | what is in it | mon hash base | mon hash boss | result |
|---|---|---|---|---|
| passive | clvl 8, Blood Moor zombies standing | d1ca6c463f45 | same | EQUAL |
| fight | clvl 30, zombies walk | 5f90630302ee | same | EQUAL |
| cold | Ice Bolt at zombies | bba8362ccde9 | same | EQUAL |
| cold2 | Frozen Orb | bba8362ccde9 | same | EQUAL |
| static | Static Field | 34a1b9ef9f44 | same | EQUAL |
| merc | hired merc, three zombies | 1ecd7abc1cd7 | same | EQUAL |
| pet | druid, wolf and raven, two zombies | 99d912dcc2e9 | same | EQUAL |
| follow | druid, wolf, merc, walking | 9f51307f5f85 | same | EQUAL |
| hit_plain | cold skills on a plain Drowned Carcass (147 life) | 81b48ba4bdb0 | same | EQUAL |
| hit_boss | cold skills on Bloodraven (boss-column monster, not made by make_boss with a level add) | d49d11125ccc | same | EQUAL (state hash equal too) |
| hit_champion | a spawned champion Drowned Carcass | 9b7c25f25f6a | 7fc909f80dd1 | DIFFERS, explained |
| hit_unique | a spawned unique (Flesh Call, others) | 04f91fcf36ff | a0e68a9c87f7 | DIFFERS, explained |
| hit_minion | Hell Blood Moor unique pack: its minions and the uniques | 27714ca3de68 | 0f5423cf7850 | DIFFERS, explained |
| hit_immune | Hell: cold-immune zombie pack (minions of a unique pack) | 2cf37f51eea2 | aeffc4504155 | DIFFERS, explained |

Eight scenarios with no boss, champion or minion in them (the five of the first run, merc, pet, follow) and the two with a plain or boss-column target (hit_plain, hit_boss) are hash-equal on both the mon and the state hash (the hashes in the table are the mon hashes; the state hashes also agree, from results.txt). The four that differ contain a champion, a unique or a unique's minions, as the patch predicts.

Why they differ (V: line-by-line diff of the two logs, `det8_*_base.txt` vs `det8_*_boss.txt`, after masking the level and life fields):

- hit_champion: all 90 differing lines are the champion's life only: 444 -> 369 (spawn print level 24 = base + 3 - 1, so base 22 by arithmetic; life now read at the base level: claim D1). After masking life, 0 lines differ (positions, states, hp of others, the player).
- hit_unique: all 126 differing lines are the unique's life only: Flesh Call 700 -> 492 (spawn print level 25 = base 22 + 3, life from the base level: D1). 0 lines differ after masking life.
- hit_minion: 432 of 972 lines differ: the minions' level 67 -> 70 (D2, +3, 288 lines) and the unique's life (Bane Kill 4,374 -> 3,398, Ash Growler 13,166 -> 10,234: D1) (144 lines, which include the level change). The minions' own life is equal (their life row is read at the base level before and after: D1 with D2 together). 0 lines differ after masking level and life.
- hit_immune: 216 of 450 lines: minion level 67 -> 70 (180) and the unique's life 4,374 -> 3,398 (36); 0 after masking.
- Nothing else moves: no position, state letter, chill or stun deadline, merc, pet or player line differs in any scenario. The monsters never died in the hit scenarios, so the lower life changes no later event inside these runs (U: a scenario that kills a boss would show shorter fights; that is the claim's visible effect, not a regression).

Unexplained differences: none. Not covered (U): ghostly (D3) and berserker (B) champions are not spawned by any of these scenarios (the spawned champion kind is 16 only: the verifier's differential covers them: D3 3,427 -> 0, B 7,300 -> 0); Nightmare and Hell expansion area levels (d2d models none).

Limits: monsters and one seed (3), one run per build and scenario (plus the base reruns above), the lab patch set is not upstream (the same base for both builds, so the only difference is the five commits).

## Re-run

```sh
# worker host, scratch dir with det_scenarios_lab.py, det_scenarios_run.py (dethash2 branch tools/emu), bin/d2d.base and bin/d2d.boss
export D2D_DATA=<game data dir with the MPQs> D2D_AGENT=modsprep
tools/agentrun.sh modsprep --ttl 3400 -- tools/d2dslot.sh run -n 2 -w 60 -- \
  python3 det_scenarios_run.py run bin out passive,fight,cold,static,cold2,merc,pet,follow,hit_plain,hit_boss,hit_champion,hit_unique,hit_minion,hit_immune base,boss
python3 det_scenarios_run.py compare out hit_champion,hit_unique,hit_minion,hit_immune base boss
tools/agentreap.sh modsprep; tools/agentreap.sh --list
```

## Verdict for the queue

READY to open in the next bundle slot (the user's go still needed) with draft v2: verifier HOLDS WITH LIMITS, det hashes equal where no boss is spawned and different only by the claimed life and minion level, 0 unexplained.
