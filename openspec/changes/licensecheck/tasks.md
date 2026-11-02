# Tasks

```yaml
change: licensecheck
status: done
agent: licensecheck
branch: agent/licensecheck
worktree: ~/code/d2lod-wt/licensecheck
skip_specs: true
owns:
  - docs/research/lockstep-licences.md
  - openspec/changes/licensecheck/**
exclusive: []
gates:
  - go run ./cmd/changecheck: 0 warnings
  - go run ./cmd/leakcheck: 0 hits
result: "docs only; licences fetched and read; not legal advice"
```

- [x] Fetch and read licences (d2d, D2MOO, libTAS, speedhack-rs, PlugY)
- [x] Fetch Blizzard EULA and Legal FAQ, mods and Warden evidence
- [x] Write clean-room procedure, PROVENANCE format, notice, recommendation
