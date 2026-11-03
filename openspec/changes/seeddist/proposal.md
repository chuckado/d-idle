# SEEDDIST: is d2d's RNG stream statistically equivalent to game.exe's per-unit seeds?

Decomposition card (docs only, no Go, no game). d2d draws monster life from one world stream and drops from a unit seed it models as a step of its own game-seed stream; game.exe gives every unit `{step of game+0xd0, 666}`. Seed-for-seed equality is out of reach (docs/research/seed-study.md section 1.3), so the question is whether the OUTCOME DISTRIBUTIONS agree for the Act 1 classes the early game meets (life roll, damage roll, hit roll, first drop). Deliverable: `docs/research/seeddist-trace.md`, with tests, case counts and p-values per quantity; ledger rows only for what is proven.

Overlap (declared): none beyond reading docs/research/monster-stats-trace.md and treasure-trace.md.
