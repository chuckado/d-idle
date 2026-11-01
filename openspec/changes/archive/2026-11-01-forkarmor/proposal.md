## Why

Upstream d2d freezes and chills every melee striker of a Frozen Armor wearer; game.exe freezes a plain monster only at a negative cold effect and at most chills a champion, unique or boss (ARMORHIT, verified independently: upstream differs on 109,573 of 134,400 cases, the candidate on 0). Andariel soaks and Frozen Armor drills on a stock or lab-nexthit2 binary are optimistic. User standing authorisation 2026-10-31: functionality may be added to the private fork when documented and vetted, never in an upstream PR without approval.

## What Changes

- Private fork: branch `lab-armorhit` (from `lab-nexthit`) carries the ARMORHIT claim and tooling commits; lab binary `bin/d2d.lab-armorhit`. No new debug verb.
- This repo: `docs/forkarmor.md` (what the fork carries, vetting, binary selection, naming rule, remaining deviation, how to drop it). Docs only, no patch text (GPL, docs/decisions.md).
