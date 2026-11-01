# Tasks

```yaml
change: brand
status: in-progress
agent: brand
branch: agent/brand
worktree: ~/code/d2lod-wt/brand
skip_specs: true
owns:
  - assets/branding/**
  - docs/branding.md
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; check.sh --hygiene-only no FAIL; openspec validate --all --strict"
merge:
result:
```

## 1. Work

- [x] 1.1 pixel font, generator, three concepts
- [x] 1.2 contact sheet, docs/branding.md
- [ ] 1.3 applying anything (needs the user's go; not in this card)
