# Tasks

```yaml
change: playerroll
status: in-progress
agent: playerroll
branch: agent/playerroll
worktree: ~/code/d2lod-wt/playerroll
skip_specs: true
owns:
  - docs/research/pr-playerroll-draft.md
  - tools/emu/det_scenarios_lab.py
  - openspec/changes/playerroll/**
exclusive:
  - fork-branch:pr/player-roll
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "fork branch pr/player-roll head 28c0540a (4 commits on upstream main 6935cf64) built and measured on the worker host; draft written; not verified live (offline, no game); held for the independent verifier and the user go"
live: "no game; emulator differentials, d2d unit tests and the headless det scenarios on the worker host"
```

- [x] build pr/player-roll on upstream main (roll-top pair, physical roll, -100 floor)
- [x] differentials on the combined head
- [x] unit tests, pr_guard, presubmit, det scenarios
- [x] write docs/research/pr-playerroll-draft.md
