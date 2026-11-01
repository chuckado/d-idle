## Why

Upstream merged our PRs #152, #153, #155, #156, #157 and #158. The private fork stack (`lab-debug-skills` < `lab-nexthit` < `lab-armorhit` < `lab-freezeresume`) conflicted with the new main in two files and carried commits that upstream now has. Standing user authorisation 2026-10-31: functionality may live in the private fork when documented and vetted.

## What Changes

- Private fork: new branches `lab2-debug-skills`, `lab2-armorhit`, `lab2-freezeresume` on upstream main `52362d52` (the lab shift-click and `belt_slot` replaced by upstream's belt code, ctrl-click sell kept; the hit-once commits dropped as merged). Old branches untouched; no force-push. Binary `bin/d2d.lab2-freezeresume`.
- This repo: `docs/forkrebase.md` (conflicts, redundant commits, vetting, binary and naming rule, bot/drill impact, how to drop branches). Docs only, no patch text (GPL, docs/decisions.md).
