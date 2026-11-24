# Verifier report: "a monster tries a player's newest summon first, its merc last" (fidbatch)

> status: verifier report 2026-11-23 (independent Opus verifier, docs only, nothing pushed or posted). Subject: [pr-fidbatch-pet-order-draft.md](pr-fidbatch-pet-order-draft.md); fork branches `didle/fidbatch-pet-search-order-up` @ `21df4aa6` (one commit on upstream main `ff6d47a66`) and `didle/fidbatch-20261124` @ `678ef868`. Tags: V = run or read by this verifier (how), S = read in a doc, U = unverified.

## Verdict

PENDING (filled below once the builder jobs finish).

## 1. The claim re-derived from game.exe (V: disassembly read with capstone on the builder's game.exe, `dis.py` / `xref.py` of the lab)

| Link | What the code does | Tag |
|---|---|---|
| FUN_005b1880 | makes the player's list head (a pool node at `game+0x10f8+4*i`, first free of 8, the unit's `+0xd0` = i) | V read |
| FUN_005b1900 (0x5b1956..0x5b197c) | new node: `new.prev(+0xc) = head`, `new.next(+8) = head.next`, `head.next.prev = new`, `head.next = new`, unit `+0xd0` = list. Insert right after the head: the list runs head, newest, ..., oldest | V read |
| FUN_005dd7f0 walk (0x5dd8e0..0x5dd99e) | from each head along `+8`; the player is skipped with its whole list when another act, no room, a town room or `dist >= 0x37`; a dead player scores 0x7fffffff; then each node: `cmp edi,[best]; jae skip` (0x5dd960): taken only when strictly nearer (unsigned), so the first tried wins a tie | V read |
| Callers of FUN_005b1900 | 15 sites (xref): FUN_00573270 (0x5732cb, monster pet maker incl. the merc, guarded by its 5th arg), FUN_00579aa0 (0x579b43), Raise Skeleton FUN_005c4b00, Golem FUN_005c5100, Revive FUN_005c56c0, Druid summons FUN_005c7390, Conversion FUN_005d0350 (0x5d0501), Decoy FUN_005dc000, Valkyrie FUN_005dc1e0, and FUN_005c07a0 / 005c5250 / 005c6a80 / 005c7520 / 005d6e70 / 005d7430 (not named here) | V xref; names S (repo docs) |
| Merc re-insert on revive | the revive handler FUN_00579c00 (S: merc-revive docs) calls FUN_00579aa0 (0x579cf0), which calls FUN_005b1900(merc, player's list) at 0x579b43: a revived merc goes to the FRONT (newest) | V read |
| Merc leaves the list on death | FUN_005a6520 -> FUN_005b1a90 (0x5a6586) | V xref; S (trace) |
| Merc hire | FUN_00573270 is called from FUN_005770e0, FUN_005774f0 and FUN_00579180 (the free Rogue, S); which of the first two is hire and which load/join is not named | V xref, U names |
| Converted monster | Conversion (FUN_005d0350) inserts the converted monster with FUN_005b1900 at 0x5d0501: it is the newest pet of the caster | V xref |

So the claim "newest summon first" is exact (V). "Its merc last" is true only while every living summon was made after the merc's last insertion (hire, join, revive). After a revive (or a hire) with summons alive, game.exe tries the merc FIRST among those older summons.

## 2-5. Runs, checks, limits, presubmit

PENDING.
