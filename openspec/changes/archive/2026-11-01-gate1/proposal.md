## Why

Gate 1 of the d2d gate ladder has not passed on one seed (docs/probes/postfix-soak2.md, belt window and no-kill window in virtual minutes). The user asked for a rerun on current main (2026-11-01).

## What Changes

- One 2 h unattended hidden-d2d soak with one injected d2d kill at about 60 minutes, judged against the Gate 1 requirements of the soak-gates spec with `d2dsoak report`.
- Docs only: the result entry docs/soaks/gate1-rerun.md, a g1-log.md line, no bot or spec change.
