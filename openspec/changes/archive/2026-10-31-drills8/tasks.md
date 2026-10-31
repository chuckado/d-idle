# Tasks

```yaml
change: drills8
status: done
agent: drills8
branch: agent/drills8
worktree: ~/code/d2lod-wt/drills8
skip_specs: true
owns:
  - openspec/changes/drills8/**
exclusive:
  - instance:i0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge: "75b2469"
result: "H.pause x3 PARTIAL (F13), D6.d x3 PASS 9 s, D-sync.wp cause was the drill (mana), NPC smoke D3 12/12 and D8 a m q w PASS, full suite table in README DRILLS8; harness edits: Mana 35 in D-sync, clearDiag. Overlap note: docs/probes/drills/** belongs to m1-liveproof, this card writes only *drills8* files and its README section."
live: "verified live on the hidden slot, lab build cheats3, code at 1628eeb (fixstall3 merged)"
```

- [x] 1.1 H.pause x3 invocations
- [x] 1.2 D6.d x3
- [x] 1.3 D-sync waypoint arrival x3
- [x] 1.4 NPC interaction smoke (D3, D8)
- [x] 1.5 full drill suite regression table
