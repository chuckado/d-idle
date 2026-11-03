# Tasks

```yaml
change: botab
status: in-progress
agent: botab
branch: agent/botab
worktree: ~/code/d2lod-wt/botab
skip_specs: true
owns:
  - docs/research/botab-20261102.md
  - docs/research/botab-trace.md
  - tools/baseline/botab_extract.py
  - tools/baseline/botab_run.py
  - openspec/changes/botab/**
exclusive: []
gates:
  - "python3 -I tools/baseline/botab_run.py <gate2 minute file> runs; changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Docs and offline scripts: bot-level d2d vs game.exe table from committed minute lines; every XP/h verdict INCONCLUSIVE, kills/h d2d about 0.76 of game.exe. No game run."
live: "offline only, nothing run live"
```

- [x] extractor and driver on the committed minute lines (d2d gates 1 and 2, game.exe late set, baselines.json bands)
- [x] per-metric table, comparability table, gap list, minimum new runs
- [x] trace doc with D/B lists, limits, shared-file lines
