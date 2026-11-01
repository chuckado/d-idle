# Tasks

```yaml
change: skground2
status: in-progress
agent: skground2
branch: agent/skground2
worktree: ~/code/d2lod-wt/skground2
skip_specs: true
owns:
  - docs/research/sk-ground2-trace.md
  - docs/research/pr-blizzard-draft.md
  - tools/emu/sk_gnd2_*
  - openspec/changes/skground2/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs and offline tooling. D1 proven on a real d2d build (14,926 / 14,957 of 100,000 shards off game.exe's grid before, 0 / 0 after one line); local PR candidate fix/blizzard-shard-offsets (9a1a90eb tooling, 681d52a9 claim) on the worker host, pr_guard 0/0, independent verifier next, nothing pushed. D4 proven on a real build (shard 5 / 13 / 21 vs d2d 21 subtiles; Fire Wall 1 / 5 / 9 vs 21). Fire Wall cast, makers and hits traced (V run 80 / 80). Hydra read to the spawn; AI not run. Not verified live."
live: "docs only, nothing run live"
```

- [x] Fire Wall (do 24, maker do 6, hits) read and run against a real d2d build
- [x] Hydra start check and do 144 read
- [x] D4 oracle (game.exe unit search) and d2d real-build measurement
- [x] D1 real-build differential, local PR candidate, pr_guard
- [x] trace doc, PR draft, matrix cells, drill recipe, shared-file lines
