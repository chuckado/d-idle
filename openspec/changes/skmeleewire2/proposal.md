## Why

On the d2d backend the player's stat vector has no stats 19 / 21 / 22 / 25 (d2d keeps no player stat for them), so `combat.meleeBlow` returns nil for any wielded weapon and the melee model is never fed. The private fork's `lab2-stats` branch adds `gear` and `weapons` to `dump stats` (docs/forkdbg.md).

## What Changes

- `internal/d2dsrc/gearstats.go` (+ test): maps `gear` sums (stats 19, 25, flats 21 / 22) and the right-hand `weapons` base range into stats 19, 21, 22, 25 of `Player.Stats`; no `gear` key means nothing is added (defaults unchanged). Hooked in `playerStats`.
- `docs/skmeleewire2.md`: what is read, what is not, open live checks.

## Impact

Offline Go and docs. Overlap (declared): `internal/d2dsrc/player.go` (small hook) is outside this card's own melee files; no other card owns it in docs/OWNERS. Builds on skmeleewire (combat/melee.go unchanged).
