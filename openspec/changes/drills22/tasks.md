# Tasks

```yaml
change: drills22
status: in-progress
agent: drills22
branch: agent/drills22
worktree: ~/code/d2lod-wt/drills22
skip_specs: true
owns:
  - openspec/changes/drills22/**
  - internal/drill/d22.go
  - cmd/d2drill/d22_lab.go
  - docs/probes/drills/2026-11-01-drills22-*.txt
exclusive:
  - instance:drills-slot-0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "see DRILLS22 in docs/probes/drills/README.md"
live: "see DRILLS22"
```

- [x] 1.1 group T drills (control, Teleport legs, bind, town, refused hops)
- [x] 1.2 live runs on the fixtelekey build
- [ ] 1.3 regression D2, D3, DW.m3
