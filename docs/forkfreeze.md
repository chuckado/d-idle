# FORKFREEZE: the private fork carries the freeze-end resume fix (card forkfreeze, 2026-11-01)

> status: current. Branch `lab-freezeresume` of the private d2d fork, lab binary `bin/d2d.lab-freezeresume` on the worker host. Patches never live in this repo (GPL, docs/decisions.md). Standing user approval 2026-10-31: functionality in the private fork when documented and vetted; never part of an upstream pull request without the user's go (the upstream candidate is the separate `fix/freeze-resume`, [pr-freezeresume-proof.md](research/pr-freezeresume-proof.md), verified by [pr-freezeresume-verifier.md](research/pr-freezeresume-verifier.md), body draft [pr-freezeresume-draft.md](research/pr-freezeresume-draft.md), held for the user's go).

Tags: **V run** measured on a real build in this card; **S** source or earlier doc; **I** inferred; **U** not verified.

## Why

When a freeze ends, game.exe's state-1 callback refiles the monster's next think at the end frame plus its MonStats aidel (6 to 15 frames; Zombie 15 / 14 / 13 by difficulty); d2d let it act at the end frame. Every d2d freeze (Glacial Spike, and Frozen Armor since [forkarmor.md](forkarmor.md)) therefore gave soaks and drills a monster that comes back sooner than in game.exe, which flatters nothing but understates the player's time. The independent verifier, running game.exe's real event queue, found the claimed frozen classes differ in 34,200 of 34,299 cases on upstream main and in 0 on the candidate. The fork carries the fix so cold-skill runs measure game.exe-like freeze timing.

## What the fork carries

- Branch `lab-freezeresume` on the private remote, built from `lab-armorhit` (which holds `lab-nexthit`: the Nova hit-once fix and the Frozen Armor freeze-rule fix). Pushed tip `efe7bc368b361d06b14c116f123291981caa2ce9` (pushed the same day, before this doc cites it).
- Two commits cherry-picked from the upstream candidate with no conflict: the claim (`c6cc4f12`, now `7a86e096`: new `Monster::stunned_until`, set at the one stun site, and `next_act` = freeze end + aidel at the first tick past a stand that was a freeze) and the tooling (`efacc7f4`, now `efe7bc36`: `test_freeze_resume` and the oracle `freeze_resume.py`). Both stay on the fork: the test binary is needed here to vet the build. S, git log on the fork.
- Interaction with the other two fixes: the claim touches `strike.cpp` (the stun site) and `ai.cpp`; the armor claim touches `states.cpp` and `gamedata`; no overlap, and the differentials below show no interference.
- No new debug verb, no change to docs/cheats.md or docs/forkdbg.md. Every other lab behaviour is unchanged.

## Vetting (V run, worker host, fork build `lab-freezeresume` efe7bc36, Release, `D2D_WERROR=OFF`)

Targets built: `d2d`, `test_next_hit`, `test_armor_hit`, `test_freeze_resume`, plus the verifiers' `test_verify_freeze_resume` and `test_verify_armor_hit` copied in untracked from `agent/verifyfreezeresume` and main.

| Check | Result |
|---|---|
| Freeze resume, the author's `freeze_resume.py` / `test_freeze_resume` (37,776 cases per seed, recorded game.exe side, seeds 0xF2E50001 and the second seed of the proof) | frozen in game.exe (plain + minion): **0 of 12,548 and 0 of 12,516** differ, every aidel value 0 (upstream main: 12,503 and 12,459). Specials differ as before (25,136 / 25,148: freeze-rules C3, not this claim) |
| Freeze resume, the verifier's harness (`verify_freeze_resume.py`, `test_verify_freeze_resume --vfr`, 31,484 cases per seed, seeds 0xF3EE0001 and 0xA5C30002, recorded game.exe side, order g = game.exe's) | single, prior freeze, refreeze inside / on the end frame / in the delay window / on the think frame / after, think due on the freeze frame: **0 differ on both seeds** (seed 1 total claimed frozen 17,180; seed 2 17,119). Residuals identical to the verifier's report on the candidate: prior stun 397 of 1,881 and 422 of 1,978, zero-length 308 of 312 and 320 of 320, specials 11,912 of 12,111 (seed 1) |
| The same harness, d2d's own tick order (order n, seed 1) | prior freeze 19, refreeze in the delay window 2, refreeze on the think frame 38 of 38, think due on the freeze frame 253 of 253 differ: the tick-order residual below, same counts as the verifier's report |
| Armor differential, the author's `armor_hit.py` (108,506 cases, seed 0x41524d48) | **0 of 108,506** differ, calc1 differs 0, RESFAIL 0 (as on `lab-armorhit`) |
| Armor differential, the verifier's harness (134,400 cases, seeds 0xA101, 0xB202, 0xC303, 0xD404) | chill and freeze lengths and end frames differ in **0 of 134,400**, calc1 0. The harness's resume column (game.exe resume minus d2d stun end) still differs in 52,714 (14,513 + 12,710 + 12,745 + 12,746), the same counts as before this fix: that column reads the stun end, which the claim does not move (the claim moves the think after it) |
| Hit-once, `test_next_hit --next-hit 1 100000` (docs/forknext.md) | **0 of 100,198** T lines differ, 0 H lines differ: the three fixes do not interfere |

Game.exe's side of every differential is the recorded run of the proof or the verifier's report (it does not depend on the d2d build); only the d2d side was re-run on the fork build.

## Live readback (V run for the numbers, I for what they mean; one hidden slot, `--no-save` throwaway Sorceress, `d2drill -probe`, every debug call ledgered)

Method: `D2D_AGENT=forkfreeze agentrun.sh forkfreeze --ttl 300 -- d2dslot.sh run -n 2 -- d2drill -group D1 -bin <binary> -probe '<verbs>'` on the worker host (idle+nice 26 percent beforehand; `tools/capacity.sh` not used). Blood Moor (`debug goto 2`), `debug kill`, `debug points 55 1` (Glacial Spike), `debug mana full`, `debug spawn zombie1 1 near` (`#17`), `debug mon set 17 life 100000 maxlife 100000 dmg 3 3`, `cmd move 31 48`, `step 10`, `debug cast 55 unit 17`, then `step 1` and `monsters` 120 times. The harness has no verb that reads a monster's next action frame (`next_act`), and no bot code was added; the observable is the Zombie's mode column in `monsters` (NU while frozen, WL once it acts again). Lockstep from the same seed, so the cast lands on the same step in both runs.

| Binary | Zombie #17 mode: walking (WL) at step | frozen (NU) from step | WL again at step |
|---|---|---|---|
| `lab-armorhit` 7c73b881 | 2 | 12 | **37** |
| `lab-freezeresume` efe7bc36 | 2 | 12 | **44** |

The freeze starts on the same step and the Zombie acts 7 `step`s later on `lab-freezeresume`. A Zombie's aidel is 14 on Normal (MonStats, V run); if one `step` is two game frames the shift is the 14-frame aidel, if it is one frame it is half of it. Which one is U (the unit of `step` in lockstep was not measured here); the sign and the freeze start match the differential. One run per binary, no repeat.

## Binary and how runs select it

- File `bin/d2d.lab-freezeresume` on the worker host (md5 prefix f118e19d0a05; Release, `D2D_WERROR=OFF`; build script `build-freezeresume.sh`, worktree `wt-freezeresume`, build dir `build-freezeresume`, vet script `freezeresume-vet.sh`, outputs in `freezeresume-out`, all under the fork directory). Pass it as `D2D_BIN` or `-bin`: `D2D_BIN=<fork>/bin/d2d.lab-freezeresume tools/d2dslot.sh run -n N -- d2drill -bin $D2D_BIN ...`. The default binary of each tool is unchanged.
- Other agents keep `bin/d2d.lab-armorhit`, `bin/d2d.lab-nexthit2`, `bin/d2d.lab-debug-skills` and the cheats3 binary: this card touched none of them.

## Rules for results

- **Every cold-skill soak or drill result (Glacial Spike, Frozen Armor, Shiver Armor, Frost Nova and Nova) names the binary** (file name, branch, hash). Cold-skill soaks should prefer `lab-freezeresume`: it contains the `lab-armorhit` fix (Frozen Armor freezes by the freeze rule) and the `lab-nexthit2` fix (hit-once), so one binary is right for all of them. Numbers from `lab-armorhit` or older are not comparable with it for anything where a freeze ends and the monster acts again; do not merge them into one series.
- Stale-binary risk: a slot still pointing at `d2d.lab-armorhit`, `d2d.lab-nexthit2`, `d2d.lab-debug-skills`, `d2d.lab6` or the cheats3 binary silently reports the old timing. Check the binary in the run log before reading any cold-skill result.
- Frost Nova and Shiver Armor chill on d2d (no freeze): their timing is not changed by this binary, but name it anyway. Glacial Spike and Frozen Armor freeze: those are the results that move.

## Known deviations NOT fixed here

- **d2d's tick order.** `Fight::world` runs `monster_update` before the player's missiles strike; game.exe's frame (FUN_005414d0) runs missiles, then players, then monsters. A think due on the frame a freeze lands still runs in d2d (253 of 253 in the n-order harness), and a refreeze on the resume frame comes after the think (38 of 38). S/V run (the verifier, reproduced above); pre-existing and independent of this claim. Not fixed: a separate card.
- **Prior stun.** game.exe refiles at the freeze end + aidel even under a longer stun; d2d keeps one stand end and refiles nothing (the claim's `// ponytail:`): 397 of 1,881 and 422 of 1,978 differ, d2d later by 5 to 11 frames typically.
- **Zero-length freeze.** game.exe freezes 0 frames and thinks at now + 1 + aidel; d2d's strike with freeze 0 does not freeze, and if it ever did the claim would refile one frame early. Unreachable on upstream and on the fork today (no divisor).
- Specials (champions, uniques, superuniques, boss rows) are still frozen by Glacial Spike on d2d (freeze-rules C3, not carried here) and now also wait aidel.

## Dropping it

When upstream merges the claim (or an equivalent), rebuild from the new upstream plus the lab patch set, rerun the freeze differentials (expect 0 of the claimed frozen classes), both armor differentials and the hit-once differential (expect 0, 0 and 0 of 100,198), then retire `lab-freezeresume` and `bin/d2d.lab-freezeresume` and point runs back to the main lab binary. If the cherry-picks conflict at that point, the branch tip is only a base for the lab verbs: take the upstream version of the claim.

## Shared-file lines for the merger

- docs/BACKLOG.md: `- FORKFREEZE: lab-freezeresume carries the freeze-end resume fix (docs/forkfreeze.md); cold-skill soaks use bin/d2d.lab-freezeresume and name it; separate card for d2d's tick order (monster_update before the player's missiles); PR fix/freeze-resume body revised, held for the user's go (det state hashes still owed)`
- docs/INDEX.md: `- docs/forkfreeze.md: the private fork carries the freeze-end resume fix (lab-freezeresume): vetting, binary, naming rule, remaining deviations`
- CHANGELOG.md (unreleased, docs): `- docs: FORKFREEZE: the private d2d fork carries the freeze-end resume fix as lab-freezeresume (docs/forkfreeze.md).`
- GLOSSARY.md: none missing.
