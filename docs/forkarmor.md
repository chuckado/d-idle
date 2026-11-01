# FORKARMOR: the private fork carries the Frozen Armor freeze-rule fix (card forkarmor, 2026-11-01)

> status: current. Branch `lab-armorhit` of the private d2d fork, lab binary `bin/d2d.lab-armorhit` on the worker host. Patches never live in this repo (GPL, docs/decisions.md). Standing user approval 2026-10-31: functionality in the private fork when documented and vetted; never part of an upstream pull request without the user's go (the upstream candidate is the separate `fix/armor-hit`, docs/research/pr-armorhit-proof.md, verified by docs/research/pr-armorhit-verifier.md).

Tags: **V run** measured on a real build in this card; **S** source or earlier doc; **I** inferred.

## Why

Upstream d2d freezes and chills every melee striker whose blow did damage to a Frozen Armor wearer. game.exe never does both: a plain monster is frozen only at a negative cold effect (for length / MonsterFreezeDivisor) and not chilled; a champion, unique, superunique or MonStats boss (Andariel) is at most chilled, and not at all at cold effect 0; every length is cut by the striker's cold resistance less the wearer's Cold Mastery pierce (S, pr-armorhit-proof.md). The independent verifier's differential: upstream main differs from game.exe on 109,573 of 134,400 cases, the candidate on 0. Andariel soaks and Frozen Armor drills on a binary without the fix are optimistic (the boss is frozen and chilled by every blow). The fork carries the fix so our runs measure game.exe-like Frozen Armor.

## What the fork carries

- Branch `lab-armorhit` on the private remote, built from `lab-nexthit` (the Nova hit-once fix with its end-frame commit, docs/forknext.md, so every NextHit result also holds here). Pushed tip `7c73b881ca247745b76cc36ed16e7b8498d56469` (pushed the same day, before this doc cites it).
- Two commits cherry-picked from the upstream candidate with no conflict: the claim (`36cbe015`, now `e1eb0c3c`: `buff_events` follows game.exe's freeze rule, one new `MonsterFreezeDivisor` member read from DifficultyLevels) and the tooling (`d5bd19c1`, now `7c73b881`: `test_armor_hit` and the oracle). Both stay on the fork: the test binary is needed here to vet the build.
- The fork does not carry the freeze-rules v2 branch (checked: `lab-nexthit` has one `MonsterFreezeDivisor` load line, in `gamedata_load.cpp`), so the verifier's duplicate-load-line merge note does not arise. If freeze-rules v2 is ever added here, drop the duplicate member and load line.
- No new debug verb, no change to docs/cheats.md or docs/forkdbg.md. The Act 3 merc's Frozen Armor (`merc_buff_events`) is NOT changed, as in the candidate (S, verifier finding 1). Every other lab behaviour is unchanged.

## Vetting (V run, worker host, fork build `lab-armorhit` 7c73b881, Release, `D2D_WERROR=OFF`)

| Check | Result |
|---|---|
| Armor differential, the author's `armor_hit.py` (108,506 cases, seed 0x41524d48; cases md5 521488a95640, game.exe side a707f7181af0, d2d side 2b0833de80b6) | **0 of 108,506** differ (upstream main: 95,920); calc1 differs 0, RESFAIL 0 |
| Armor differential, the verifier's harness (`tools/emu/verify_armor_hit.py` and `test_verify_armor_hit.cpp` as on main; 134,400 cases, seeds 0xA101 (with edge cases), 0xB202, 0xC303, 0xD404) | chill and freeze end frames differ in **0 of 134,400** (upstream main: 109,573); calc1 differs 0. The AI resume frame differs on 52,714 (14,513 + 12,710 + 12,745 + 12,746), the same counts as the verifier's report on the candidate: the known residual below |
| Hit-once differential, `test_next_hit --next-hit 100000` seed 1 (docs/forknext.md) | **0 of 100,198** per-monster hit lines differ, 0 H lines differ: the two fixes do not interfere |

Game.exe's side of both armor differentials is the same recorded run as in the proof and the verifier's report (the game side does not depend on the d2d build); only the d2d side was re-run on the fork build.

## Live readback (V run, one hidden slot, `--no-save` throwaway Sorceress, `d2drill -probe`, every debug call ledgered)

Method: `D2D_AGENT=forkarmor tools/agentrun.sh forkarmor --ttl 300 -- tools/d2dslot.sh run -n 2 -- d2drill -group D1 -bin <binary> -probe '<verbs>'`, capacity checked first (idle + nice 62 percent). Blood Moor (`debug goto 2`), `debug kill`, `debug states set 40 20` (Frozen Armor level 20), `debug spawn <row> 1 near [champion|unique]` (always `#17`), `debug mon set 17 life 100000 maxlife 100000 dmg 3 3` (so she survives and the striker lives), `cmd move 31 48` to walk up to it (a spawned monster idles until she is close), `step 10` and `debug states` thirty times. The chill value is not in any shipped verb: both builds below are the shipped branch heads plus one scratch change that prints `chill_until` in the monster lines of `debug states` (`bin/d2d.probe-nexthit2` = `lab-nexthit` and `bin/d2d.probe-armorhit` = `lab-armorhit`, unpushed, not for runs). The two runs are in lockstep from the same seed, so the first hit lands on the same frame and the end values compare directly (lab clock units as `debug states` prints them; the first reading after the first hit):

| Striker (one melee blow landed) | `lab-nexthit2` stun / chill end | `lab-armorhit` stun / chill end |
|---|---|---|
| plain Zombie | 10560 / 10560 | 10560 / **0** (frozen, not chilled) |
| plain Fallen | 6800 / 6800 | 6800 / **0** |
| champion Zombie | 9520 / 9520 | **0** / 9520 (chilled only) |
| unique (Mold Maim the Mad) | 10560 / 10560 | **0** / 10560 |
| Andariel (boss column) | 5600 / 5600 | **0** / **3840** (chilled only, shorter: her cold resistance) |

Reading: on `lab-nexthit2` every striker is frozen and chilled to the same end; on `lab-armorhit` bosses are never frozen. The plain Zombie and Fallen are still frozen: the rule freezes a plain monster at a negative cold effect, so their rows must have one (I: their MonStats cold effect was not read; no plain row with cold effect 0 or more was tried, the differential covers that case with 0 differences). Later hits in the same runs extend the end values the same way on both builds. A quill rat never struck within the window on either build.

## Binary and how runs select it

- File `bin/d2d.lab-armorhit` on the worker host (Release, `D2D_WERROR=OFF`; build script `build-armorhit.sh`, worktree `wt-armorhit`, build dir `build-armorhit`, targets `d2d`, `test_next_hit`, `test_armor_hit`, plus the verifier's `test_verify_armor_hit` copied in untracked). Pass it as `D2D_BIN` or `-bin`: `D2D_BIN=<fork>/bin/d2d.lab-armorhit tools/d2dslot.sh run -n N -- d2drill -bin $D2D_BIN ...`. The default binary of each tool is unchanged.
- Other agents keep `bin/d2d.lab-nexthit2`, `bin/d2d.lab-debug-skills` and the cheats3 binary: this card touched none of them.

## Rules for results

- **Every Frozen Armor or Shiver Armor soak or drill result names the binary** (file name, branch, hash). Numbers from `lab-nexthit2` or older are optimistic for these armors (every attacker frozen and chilled) and are not comparable with `lab-armorhit`; do not merge them into one series. Andariel soaks with a Frozen Armor Sorceress must use `lab-armorhit`.
- Nova and Frost Nova results should use `lab-armorhit` as well: it contains the `lab-nexthit2` fix (the 0 of 100,198 line above), so one binary is right for both.
- Stale-binary risk: a slot still pointing at `d2d.lab-nexthit2`, `d2d.lab-debug-skills`, `d2d.lab6` or the cheats3 binary silently reports the old numbers. Check the binary in the run log before reading any Frozen Armor result.
- Shiver Armor: the claim changes the melee striker answer of Frozen Armor's aura event (`buff_events`), whose length comes from the Frozen Armor skill; Shiver Armor's own missile-side chill is unchanged (S, verifier CI note: Chilling Armor and Shiver Armor missiles are the freeze-rules candidate). Name the binary anyway.

## Known deviation not fixed here

When a freeze ends, game.exe resumes the monster's AI at the end frame plus its aidel (13 to 15 frames for the common rows, S, verifier section 1: FUN_0057b170); d2d acts at the end frame. This is the 52,714 resume differences above and applies to every d2d freeze, cold-skill ones included. Cold-skill results on d2d (Frost Nova, Ice Blast, Frozen Armor) are therefore still pessimistic by that delay. Card FREEZERESUME covers it; it is not built into this binary.

## Dropping it

When upstream merges the claim (or an equivalent), rebuild from the new upstream plus the lab patch set, rerun both armor differentials and the hit-once differential (expect 0, 0 and 0 of 100,198), then retire `lab-armorhit` and `bin/d2d.lab-armorhit` and point runs back to the main lab binary. If the cherry-picks conflict at that point, the branch tip is only a base for the lab verbs: take the upstream version of the claim.
