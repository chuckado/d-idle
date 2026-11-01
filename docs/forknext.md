# FORKNEXT: the private fork carries the NextHit fix (card forknext, 2026-11-01)

> status: current. Branch `lab-nexthit` of the private d2d fork, lab binary `bin/d2d.lab-nexthit` on the worker host. Patches never live in this repo (GPL, docs/decisions.md). Standing user approval 2026-10-31: functionality in the private fork when documented and vetted; never part of an upstream pull request without the user's go (the upstream candidate is the separate `fix/nexthit`, docs/research/pr-nexthit-proof.md on branch agent/nexthit).

Tags: **V run** measured on a real build in this card; **S** source or earlier doc; **I** inferred.

## Why

Upstream d2d strikes one monster with every Nova or Frost Nova ring missile that touches it, so one cast lands 4.43 (Nova) and 4.16 (Frost Nova) hits per touched monster on average and 64 at point blank on a size-3 monster; game.exe lands 1.03 and 1.04 (V run, NEXTHIT differential, re-run below). Soak and drill numbers for the Frost Nova crowd rule and any Nova build are inflated on a binary without the fix. The fork carries the fix so our runs measure game.exe-like hits.

## What the fork carries

- Branch `lab-nexthit` on the private remote, built from `lab-debug-skills` (all earlier lab verbs plus the skill-drill verbs, docs/forkdbg.md). Pushed tip `989517079642c3f73350c78ea39529a766ea8042` (pushed the same day, before this doc cites it).
- Two commits cherry-picked with no conflict from the upstream candidate: the claim `8abab0ed` (now `27a172fc` on the branch: a NextHit hit marks the unit `justhit` for NextDelay frames, a LastCollide row skips its last unit, a missile hits one unit a frame) and the tooling `7901fda2` (now `98951707`: the differential test mode and oracle). Both stay on the fork: the test binary is needed here to vet the build.
- No new debug verb, no change to `docs/cheats.md`. Every other lab behaviour is unchanged; rows without NextHit are untouched (S).

## Vetting (V run, worker host, same cases and tools as NEXTHIT)

Differential: the fork's real `Fight::fire` and `Fight::world` run one Nova or Frost Nova over 1..6 standing monsters; game.exe's own filter and hit code replay the same contacts (`tools/emu/next_hit.py`). 100,198 cases (198 boundary, 100,000 random, seed 1).

| Build | per-monster hits differ | H lines differ | d2d hits per touched monster per cast (Nova / Frost Nova) | game.exe |
|---|---|---|---|---|
| `lab-debug-skills` d9b96c2b plus the test mode only | **84,779 of 100,198** | 248,513 | 4.43 / 4.16 | 1.03 / 1.04 |
| `lab-nexthit` 989517079642 | **0 of 100,198** | 0 | 1.03 / 1.04 | same |

The fork's extra commits do not change these contacts: the contact lines of the fork run equal the upstream candidate's run line for line, and the game.exe output hashes alike (prefix 53bbb1d6cfc0). Both numbers equal the upstream proof's, so each build agrees with game.exe on its own contacts.

Live smoke (one hidden slot, `--no-save` throwaway Sorceress, `d2drill -probe`, every debug call ledgered; Blood Moor, `debug points 44 20` and `48 20`, `debug mon set` to 100000 life, no damage, all resistances 0; one `debug cast <skill> near` then `step 40`, life read from `monsters`, refilled between casts; monster adjacent to her for casts 3 to 6):

| Cast | `lab-debug-skills` (life points lost) | `lab-nexthit` |
|---|---|---|
| Frost Nova 1, monster about 2.7 cells away | 66 | 66 |
| Frost Nova 2 | 61 | 61 |
| Frost Nova 3, adjacent | **249** | **67** |
| Nova 1, adjacent | **1438** | **174** |
| Nova 2 | 1396 | 151 |
| Nova 3 | 1445 | 136 |

Reading: one hit is about 60 to 70 points of Frost Nova and about 130 to 175 of Nova at level 20 (the nexthit column), so the stock binary lands about 4 and about 9 to 10 hits at point blank where `lab-nexthit` lands one (I: hits = life lost over the one-hit range). The two builds agree where the ring touches a monster once. Same seed, lockstep, so the casts are comparable; the damage roll is random per hit.

## Binary and how runs select it

- File `bin/d2d.lab-nexthit` on the worker host (Release, `D2D_WERROR=OFF` like every lab build). Pass it as `D2D_BIN` or `-bin`: for example `D2D_BIN=<fork>/bin/d2d.lab-nexthit tools/d2dslot.sh run -n N -- d2drill -bin $D2D_BIN ...`. The soak and drill tools take the binary from `D2D_BIN` or their `-bin` flag; the default stays whatever the tool names (`d2d.lab6` for `d2drill`), so nothing changes unless a run asks for it.
- Other agents keep `bin/d2d.lab-debug-skills` and the cheats3 binary: this card did not touch them.
- Rebuild: worktree of the fork branch, `cmake -DCMAKE_BUILD_TYPE=Release -DD2D_WERROR=OFF`, targets `d2d` and `test_next_hit`.

## Rules for results

- Every soak or drill result measured with `lab-nexthit` names the binary in its report (file name, branch, hash). Frost Nova and Nova numbers (kill rates, time to kill, packs cleared, mana per kill) differ from `lab-debug-skills` runs by design: about 4x fewer hits on a ring, up to about 10x at point blank. Do not compare across the two binaries, and do not merge them into one series.
- Stale-binary risk: a slot that still points at `d2d.lab-debug-skills`, `d2d.lab6` or the cheats3 binary silently reports the old, inflated numbers. Check the binary named in the run log before reading any Nova result. A run that never uses Nova or Frost Nova is unaffected (S, rows without NextHit are not touched).
- Not covered by the fix (S, pr-nexthit-proof limits): Chain Lightning's hop search still picks the just-struck monster (claim C2, not built), piercing Lightning (C3, not built), two casts within 4 frames of each other and moving monsters are not in the differential. Monster missiles do not read NextHit.

## Dropping it

When upstream merges the claim (or an equivalent), rebuild from the new upstream plus the lab patch set, run the differential again (expect 0 of 100,198), then retire `lab-nexthit` and `bin/d2d.lab-nexthit` and point runs back to the main lab binary. If the cherry-picks conflict at that point, the branch tip is only a base for the lab verbs: take the upstream version of the claim.
