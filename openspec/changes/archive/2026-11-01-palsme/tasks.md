# Tasks

```yaml
change: palsme
status: archived
agent: palsme
branch: agent/palsme
worktree: ~/code/d2lod-wt/palsme
skip_specs: true
owns:
  - docs/palsme.md
  - openspec/changes/palsme/**
exclusive: []
gates:
  - "gofmt, go vet, go test -p 1 per touched package, whole repo on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Zealot-lite table adjusted per the SME review (Prayer 3 -> 1, Cleansing and Vigor dropped, Defiance 1 kept, points to Might and Holy Fire); legality passes incl. the install check on the worker host; Paladin policy items recorded as ready cards; not verified live"
live: "offline only"
merge: f0e7163
```

- [x] Zealot-lite table (Prayer 1, Might, Defiance 1, Holy Fire 2)
- [x] legality tests, install check on the worker host
- [x] docs/palsme.md with ready cards
