# Tasks

```yaml
change: brandapply
status: in-progress
agent: brandapply
branch: agent/brandapply
worktree: ~/code/d2lod-wt/brandapply
skip_specs: true
owns:
  - README.md
  - cmd/d2bot/banner.go
  - cmd/d2bot/banner_test.go
  - cmd/d2bot/main.go
  - cmd/d2dview/web/icon.svg
  - cmd/d2dview/web/favicon.ico
  - cmd/d2dview/web/apple-touch-icon.png
  - cmd/d2dview/brand_test.go
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; check.sh --hygiene-only no FAIL; openspec validate --all --strict; go test -p 1 ./cmd/d2bot ./cmd/d2dview; whole-repo vet and test -short on the worker host"
merge:
result: PREPARED, NOT MERGED: awaiting the user's go (and brandfinal merged first). Offline-tested only.
```

## 1. Work

- [x] 1.1 README header image and Discord line
- [x] 1.2 favicon files in cmd/d2dview/web, test; head links as a patch (mapitem owns the two pages)
- [x] 1.3 usage banner in d2bot (-h only), tests
- [ ] 1.4 merge: needs the user's go
