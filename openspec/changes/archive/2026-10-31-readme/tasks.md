# Tasks

```yaml
change: readme
status: archived
agent: readme
branch: agent/readme
worktree: ~/code/d2lod-wt/readme
owns:
  - README.md
  - docs/README.md
  - GLOSSARY.md
  - CHANGELOG.md
  - deploy/nas/my-d2idle.xml
  - docs/d2d-container/README.md
  - docs/d2d-container/my-d2dbot.xml.draft
  - docs/d2d-container/gpl.md
  - docs/quickstart.md
  - docs/troubleshooting.md
  - docs/research/README-public-draft.md
  - docs/stream/kick-api.md
  - docs/stream/kick-title.md
  - web/setup.js
  - openspec/changes/readme/**
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: e62f519 (agent/readme tip, merge --no-ff)
result: README rewritten, name restyled to d.idle in docs, container templates and 21 web/setup.js strings; text only; verified offline on the worker host, no game
```

## 1. Work

- [x] 1.1 Name styled d.idle in public prose and UI strings (git grep -n -E 'D\.[iI]dle'); ADRs, archive, history lines left
- [x] 1.2 README.md refreshed (status, backends, how we check it, layout, contributing); CHANGELOG entry
