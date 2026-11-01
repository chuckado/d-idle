## Why

The SKMISS drill recipe (docs/research/sk-missile-trace.md section 8, M1-M5) needs four lab-only control-socket verbs the private d2d fork lacked: set or clear a player state, set a monster's life, damage and resistances, make the player cast a skill through the normal path, refill mana. User standing authorisation 2026-10-31: cheats or functionality may be added to the private fork when documented and vetted, never in an upstream PR.

## What Changes

- Private fork (branch `lab-debug-skills` on the private remote, built from the cheats3 lab branch): `debug mana full`, `debug states set|clear`, `debug mon set`, `debug cast`. Patches stay out of this repo (GPL, docs/decisions.md).
- This repo: `internal/d2dctl` gets a client (`SkillDebug`) that writes every call to the cheat ledger (kind `d2d-debug`, a required reason) before sending it; `docs/forkdbg.md` documents the verbs, safety rules, ledger reasons, fork commits, vetting and the M1-M5 recipe.
- The drills themselves are not run here.
