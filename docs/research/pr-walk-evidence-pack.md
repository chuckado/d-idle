# PR candidate evidence pack: walk velocity

> status: local candidate, verifier NOT run, nothing posted (card decompwalk, 2026-10-31). Proof: docs/research/pr-walk-proof.md; body: docs/research/pr-walk-draft.md.

## Branch

`fix/walk-velocity` in the worker host's `~/d2d-lab/pr/walk` worktree, base upstream/main d0ea14b7; commits 96953bfe, 7e8feb87, 97d55732 (author shatteredeagle, body `Agent: Claude Code`, no Claude trailer); 5 files, +23 / -15 (ai.cpp, fight/strike.cpp, world.hpp, world/tick.cpp, rules/combat.hpp). Local only, not pushed.

## Re-run

```sh
# on the worker host, patched tables via the mpq-cat recipe into DATA (monstats.txt, monstats2.txt)
python3 tools/emu/walk_speed.py --data DATA gen cases.tsv            # 712,362 cases
g++ -O2 -std=c++20 -o h tools/emu/walk_speed_harness.cpp && ./h < cases.tsv > d2d.tsv
D2_MPQ_DIR=<mpq dir> PYTHONPATH=<upstream>/tools/emu python tools/emu/walk_speed.py --exe <game.exe> --data DATA game cases.tsv > game.tsv   # about 66 s
python3 tools/emu/walk_speed.py cmp cases.tsv game.tsv d2d.tsv
```

Result (game.exe md5 2b3600c1...; md5 prefixes cases e8fb857a, game 5eb97059, d2d 49662822):

```
cases 712362 (not a moving mode: 0); before 425665 differ; after 0 differ
  player: 226548 cases, before 208586, after 0
  monster: 485814 cases, before 217079, after 0
  term M1 34190, M2 124301, M3 58588, P1 207466, P2 1120
```

## Upstream checks

- Build: ninja `test_game test_rules test_combat d2d` with `-Wno-error=maybe-uninitialized -Wno-error=array-bounds`: OK.
- `test_game` (D2_MPQ_DIR set) rc 0, `test_rules` ok, `test_combat` ok. The baseline `test_rules` ok; ctest and `presubmit.py --quick` not run.
- `tools/pr_guard.py upstream/main`: 0 FAIL, 0 WARN.
- Callers of the changed `World::walk_percent(data, ...)`: tick.cpp (player), tests (default velocity_pct 0: unchanged); `view.walk_percent` now includes velocitypercent raw.

## Still to do before any post

Independent Opus verifier on the proof; a det/state-hash run (`tools/d2ddet`) showing only movement moved; the user's go.
