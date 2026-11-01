# Tasks

```yaml
change: brandfinal
status: archived
agent: brandfinal
branch: agent/brandfinal
worktree: ~/code/d2lod-wt/brandfinal
skip_specs: true
owns:
  - assets/branding/**
  - docs/branding.md
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; check.sh --hygiene-only no FAIL; openspec validate --all --strict"
merge: bbf46c8
result: final brand assets (SVG masters, PNGs, favicon, banners, docs/branding.md); offline, public assets only
```

## 1. Work

- [x] 1.1 final cursor asset set (SVG) with geometry and contrast checks
- [x] 1.2 stdlib PNG/ICO renderer and decoder check
- [x] 1.3 banner variants (big, plain ASCII, one line)
- [x] 1.4 docs/branding.md
- [ ] 1.5 applying anything (needs the user's go; prepared on agent/brandapply)
