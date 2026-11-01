# Tasks

```yaml
change: brand
status: archived
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
merge: brandfinal
result: concepts delivered; continued and finished by card brandfinal (contains every brand commit); on this branch the card is done and its OWNERS lines moved to brandfinal
```

## 1. Work

- [x] 1.1 pixel font, generator, three concepts
- [x] 1.2 contact sheet, docs/branding.md
- [ ] 1.3 applying anything (needs the user's go; not in this card)
