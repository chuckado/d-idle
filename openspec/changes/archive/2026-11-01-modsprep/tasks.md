# Tasks

```yaml
change: modsprep
status: in-progress
agent: modsprep
branch: agent/modsprep
worktree: ~/code/d2lod-wt/modsprep
skip_specs: true
owns:
  - docs/research/pr-mods-draft-v2.md
  - docs/research/pr-mods-evidence-addendum.md
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "draft v2 + addendum; det hashes: 10 of 14 scenarios equal, 4 differ only by the claimed unique/champion life and minion level, 0 unexplained; READY. Not verified live (no game.exe)."
live: "not applicable: docs and a headless d2d run, no game.exe"
```

- [x] draft v2 with the four verifier edits
- [x] d2d builds: base and patched (c048a356)
- [x] det scenarios on one hidden slot (14 x base, boss), table in the addendum
- [x] queue line, INDEX, CHANGELOG, BACKLOG, OWNERS
