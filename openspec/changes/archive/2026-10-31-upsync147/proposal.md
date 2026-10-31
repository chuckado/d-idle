## Why

Upstream PR #148 (directory per subsystem) merged and its description named our belt PR #147 as needing a rebase. User go 'rebase' 2026-10-31. We never rewrite pushed commits, so the branch gets a merge of upstream/main instead.

## What Changes

- merge upstream/main into the PR branch (fix/belt-scroll-columns) on the worker host, re-prove it, plain push to our fork
- evidence pack gets dated `posted:` lines
