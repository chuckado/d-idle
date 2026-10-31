## Why

PR psi29a/d2decomped#152 (price fixes) fell behind upstream main after #149 merged. The user's go (2026-10-31, "rebase 152") rebases the PR branch onto upstream main and force-pushes it; the evidence pack records the push.

no observable behaviour change: a branch rebase and an evidence-pack line, no spec number changes.

## What Changes

- rebase fix/vendor-prices (8 commits) onto upstream/main, re-verify, force-push with a pinned lease
- append the dated `posted:` line to docs/research/pr-price-evidence-pack.md
