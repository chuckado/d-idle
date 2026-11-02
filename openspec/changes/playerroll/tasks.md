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
result: "fork branch pr/player-roll built and tested on the worker host (see the draft); draft held for the independent verifier and the user's go; not verified live (offline, no game)"
live: "no game; emulator differentials, d2d unit tests and the headless det scenarios on the worker host"
```

- [x] build pr/player-roll on upstream main (roll-top pair, physical roll, -100 floor)
- [ ] differentials on the combined head
- [ ] unit tests, pr_guard, presubmit, det scenarios
- [ ] write docs/research/pr-playerroll-draft.md
