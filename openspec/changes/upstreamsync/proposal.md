## Why

The upstream maintainer merged our PR #152 (vendor prices) and #153 (waypoint arrival cooldown) on 2026-11-01, so upstream main moved from 72132b99 to 7c3a50af. Our open PRs (#155 to #158), the local candidates and the private fork branches all sit on the old base. This card measures what still merges cleanly, prepares the trivial rebases, plans the fork rebase and lists the docs that are now outdated.

## What Changes

- Read-only on upstream (fetch only). On the worker host: new branches `<name>-rebased` in new worktrees for the clean candidates; originals untouched.
- In this repo: `docs/upstreamsync.md` (mergeability table, fork rebase plan, superseded items, outdated-docs list, shared-file lines for the merger), this card, OWNERS lines.
- Nothing is pushed or posted; no PR comment; the fork is not rebased or pushed (a later card does that).

## Impact

Docs only here. No spec changes.
