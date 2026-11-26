# Necromancer keys beyond F8 (card necrokeys, 2026-11-26)

> status: current. Closes card necrokeys with evidence: on the d2d backend the F-key limit does not apply. Tags: V verified (how said), S read in code, U unverified.

## Question

The Necromancer profiles bind eight skills to f1..f8 (`internal/combat/classdata/profiles.json`, `layout`). Weaken, Decrepify, Lower Resist and Fire Golem have no key in the summoner layout (Amplify Damage, Corpse Explosion and Revive take f6 to f8). Does the bot then fail to cast them?

## Evidence

1. S: `Fighter.cast` (`internal/combat/fighter.go`) needs either `sp.Key != ""` or the skill already on the right hand; otherwise `ErrNoHotkey`. `sp.Key` comes from `Fighter.keys`, filled by `SetHotkeys`.
2. S: on d2d the production stack fills it with `Rig.BindSkills` (`internal/d2dcombat/rig.go`, called from `cmd/d2bot/d2dwire.go` in `prepare` and in the leveler `Rebind` after every point spent). It gives EVERY learned skill the combat catalog knows a pseudo key `sk<id>`, mapped in `Controller.Skills`. The profile `layout` is not read on this path.
3. S: `Controller.CastAt` for a key found in `Controller.Skills` sends `cmd skill <id> x y`; no F key, no popup, no keyboard. A skill learned later gets its key at the next `Rebind`.
4. V (earlier live work, `cmd/d2drill/d22_lab.go` and the Paladin runs of the matrix): a cast through `sk<id>` spends mana and moves the hero. Not read for a Necromancer curse or the Fire Golem (U); the path is class-blind code, so no new mechanism is involved.

## Verdict

The card's premise (a second bar or a swap rule for keys beyond F8) holds only for the game.exe path (`EnsureHotkeys`, popup binding, the `layout` map). That path is not the live backend. No profile change is made: widening `layout` to f9 and beyond would be a guess about a path nobody runs (U), and it would add drift to the ratchet of card hotkeydrift.

If the game.exe path is ever revived for the Necromancer: add the four skills to the `layout` of the summoner profile with keys the popup binder can read back (the binder verifies each key by selecting it and reading the hand), and extend `knownHotkeyDrift` accordingly.

## Not verified live

A Necromancer fight log showing a Weaken, Decrepify, Lower Resist or Fire Golem cast on d2d. Checklist for the next Necromancer drill: grep the bot log for `cast` lines naming those skills and confirm mana drops; the curse state readback (`debug skillstate`) is still not wired (22 Nov audit S3).
