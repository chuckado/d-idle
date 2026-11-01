# SKMELEEWIRE2: d2d gear and weapon fields into the melee model

> status: current (2026-11-01). Offline. Follows docs/skmeleewire.md; source fields docs/forkdbg.md "dump stats gear and weapons". Tags: V verified, S secondary, U unverified.

## What changed

SKMELEEWIRE reads player stats 19, 21, 22, 25 and returns no Blow for a wielded weapon when 21 / 22 are absent. d2d keeps no such player stats, so on d2d the model was never fed (V, forkdbg.md reading of the fork). `internal/d2dsrc/gearstats.go` now derives them from `dump stats`:

| Stat | From |
|---|---|
| 19, 25 | `gear` sum of stat 19 / 25 (param 0), 0 when the entry is missing |
| 21, 22 | right-hand (slot 4) `weapons` mindam / maxdam plus the `gear` sums of 21 / 22 flats |

No `gear` key (older fork binary) adds nothing: Stats keeps today's contents, the model keeps its defaults. A left-hand-only weapon gives no 21 / 22 (nil Blow, as before). No class or build is named. Test: `go test -p 1 -run GearStats ./internal/d2dsrc` (table, synthetic dumps shaped like the doc's example; the staff row is the live read from forkdbg.md, V).

## Not wired

- Weapon speed: not in the dump, the model has no speed term (U, no source).
- Stats 17 / 18 (enhanced min / max damage percent): the model uses reading A (0); `gear` carries them if reading B is ever needed.
- Whether the base range should take the weapon's enhanced damage before the model applies stat 25 again: the model applies `DmgPct` to Min / Max, so base + flats is passed (S, matches melee.go comment; U against the character panel).
- Dual wield, swap weapon set: `weapons` lists worn weapons of the set in use only (S).

## Live checklist (needs the lab2-stats binary, a hidden slot)

1. `dump stats` on a character with a weapon and gear: compare `gear` stat 19 / 25 with the panel.
2. Log the fed Blow (AR, Min, Max) and compare with the panel damage range and attack rating (drill R1).
