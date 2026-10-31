# PR candidate evidence pack: walk velocity

> status: local candidate, nothing pushed or posted (card walkfix, 2026-10-31; earlier: decompwalk). Proof: docs/research/pr-walk-proof.md (author's differential, the verifier's report, the fix pass); body: docs/research/pr-walk-draft.md.

PROOF: complete for the five claims. READY pending the user's go (and the det state hashes, see Residual risks).

## The claim

A unit in a moving mode walks at FUN_00623f50's rate: base MonStats Velocity (a monster unit, a merc and a pet in WL and RN), walk % = diminished faster run/walk + velocitypercent, at least 25; chill adds coldeffect to velocitypercent.

## Branch

`fix/walk-velocity-v2` in the worker host's `~/d2d-lab/pr/walk2` worktree, base upstream/main 72132b99 (the tip when checked), head 10338345. Commits: d8e63d4a monster RN at Velocity, a5cf6415 chill adds coldeffect + floor 25, 1c3fa4e5 player velocitypercent raw, 455e366f merc, 85d50c78 pet, 10338345 tooling (test_game `--walk-velocity`, `tools/emu/walk_velocity.py`, README row). Author and committer the user's noreply identity, body `Agent: Claude Code`, no trailer. 10 files, +407 / -20. Local only, not pushed. The author's `fix/walk-velocity` (97d55732 on d0ea14b7) is superseded and untouched.

## Numbers (122,646 cases, seed 9191, game.exe md5 2b3600c1...)

| claim | mismatches before | after |
|---|---|---|
| monster WL / RN (claims 1, 2) | 15,371 / 23,172 | 0 |
| player (claim 3) | 15,146 | 0 |
| merc WL / RN | 815 / 7,472 | 0 |
| pet WL / RN | 1,968 / 5,493 | 0 |
| total | 69,437 | 0 |

The author's earlier exhaustive run (712,362 cases, 425,665 differ, 0 after) is in the proof doc; it used a transcription of d2d, this one the real build.

## Commands (worker host; reviewers use the body's)

```sh
cd <walk2>/tools/emu   # D2_MPQ_DIR, D2_PATCH_INSTALLER, D2_BUILD set; venv with unicorn, capstone, pefile
python walk_velocity.py gen 100000 9191 > cases.tsv
python walk_velocity.py game cases.tsv <game.exe> > game.tsv
<walk2>/../walk2-out/build/tests/test_game --walk-velocity cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
python walk_velocity.py compare cases.tsv game.tsv d2d.tsv     # mismatches 0
```

## Upstream checks (branch head)

pr_guard 0 FAIL / 0 WARN (range and each commit); presubmit --quick 0 FAIL, 3 WARN answered in the body; clang-tidy names and includes: none in touched files; gcc -Wshadow / -Wunused / -Wsign-compare: 0 on added lines (one `rng` shadow in the first draft of the test mode found and fixed); ruff (upstream rule set) clean; test_game (data dir set), test_rules, test_combat, test_protocol rc 0.

## Residual risks

- Det state hashes (`tools/d2ddet`) not run: need a d2d slot. Movement of monsters, mercs and pets changes the hashes by design; run before posting if the user wants the proof of "nothing else moved".
- A Shadow pet's base and an untraced pet's chase are not claimed (ponytails in the code, a line in the body); skill moves keep a multiplied chill; negative IAS/FHR/FCR/FBR is a separate candidate.
- Stat 0x43 inputs are d2d's model on both sides; game.exe's integer truncation vs d2d's float step (under 1/4096 subtile a tick).
- The merc and pet claims rely on "a merc and a pet are monster units" (the verifier's disassembly reads of FUN_00621360); no live trace of a merc's path velocity.
- Opus verifier's verdict (HOLDS WITH LIMITS) is on the three original commits; the merc and pet commits and the tooling have only their own differential, not a second reviewer.

## Recommendation

Go for the bundle slot, after the user's go. Not run, for the user to approve:

```sh
git -C ~/d2d-lab/pr/walk2 push origin fix/walk-velocity-v2
gh pr create --repo psi29a/d2decomped --base main --head shatteredeagle:fix/walk-velocity-v2   # body: docs/research/pr-walk-draft.md
```
