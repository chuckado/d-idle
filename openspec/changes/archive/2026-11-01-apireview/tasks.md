# Tasks

```yaml
change: apireview
status: archived
merge: 97f2fc3
agent: apireview
branch: agent/apireview
worktree: ~/code/d2lod-wt/apireview
skip_specs: true
owns:
  - .github/workflows/pr-review.yml
  - tools/ci/pr_review.py
  - tools/ci/haiku_tick.py
  - tools/ci/anthropic_http.py
  - tools/ci/tests/**
  - docs/process/api-credits.md
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/ci/tests -t tools/ci passes; tools/ci/check.sh passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "offline, not verified live: review script, tick helper, workflow and doc; fake-HTTP tests pass (11); actionlint clean; no real API call, workflow not yet run on GitHub (first real review needs a PR from this repo)"
live: "not verified live: needs the ANTHROPIC_API_KEY secret, then open an owner PR"
```

- [x] review script and workflow
- [x] haiku tick helper
- [x] tests
- [x] docs
- [ ] owner sets the secret; first real review checked
