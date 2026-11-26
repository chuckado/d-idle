# Necromancer caps, curse memory and the corpse rule (cards necrocaps and necrocorpse, 2026-11-26)

> status: current. Tags: V verified (how said), S read in code or data, U unverified.

## Caps and reach numbers are profile data now

Before: the minion caps sat in `necro.summons[].max` without a note, and the reach and curse-memory numbers were constants in `internal/combat/necro.go`. Now (S: `internal/combat/classdata.go`, `internal/combat/classdata/profiles.json`):

- every `summons[].max` carries a `source`, and `validateNecro` refuses a profile whose cap has no tag (first word V, S or U);
- `necro.rules` holds `summonReach`, `corpseSearch`, `curseRange`, `curseLifeSec`, `curseCover` with one `source` string; a zero field reads the old constant, a set field needs the source;
- all three Necromancer profiles list today's values, so behaviour is unchanged (test `TestNecroRulesAreData` pins that the profile numbers equal the old constants).

Source status of the numbers: summonReach 4 is V (docs/classes/smoke-necromancer.md finding 2) with the exact limit U; curseRange 20 is a U middle between V 11 (landed) and V 28 (missed); curseLifeSec 30, curseCover 5 and corpseSearch 14 are U; the caps 12, 4 and 8 are U guesses.

NOT done: reading the caps from skills.txt. The table columns the loader keeps (`internal/data/skills.go` calc1 to calc4, Param1 to Param8) are raw strings; the summon count formulas are not parsed (U: whether they sit in calc1). The runtime `capSeen` memory (a paid cast that raised nothing marks the cap) is the evidence path and stays. A follow-up card should print calc1 for Raise Skeleton, Raise Skeletal Mage and Revive from the install tables before parsing anything.

Curse durations: the same gap. `Skill.AuraFrames` parses only the `ln12`/`ln34` form of `auralencalc`; curses use other forms (U). The cast memory `curseLifeSec` stays a profile number until a read of the curse rows says otherwise.

## Corpse rule (what exists, S)

- Golems (Clay, Fire) need no corpse: `planSummon` returns them first when no golem stands (`internal/combat/necro.go`).
- Skeleton, mage and revive take a corpse: `Corpses` lists dead hostile-class monsters in mode 12 within reach, minus friendlies, NPCs, non-killable and bosses (U: bosses and some classes cannot be raised), minus corpses a cast already aimed at (`tried`; V a used corpse vanishes and a second cast spent no mana). Revive takes the strongest corpse, the others the nearest.
- A raise farther than `summonReach` makes the fighter walk (only between fights); inside a fight only corpses within reach are used.
- Corpse Explosion runs only when no summon wants a corpse (`necroTick`: curse, then summon, then `corpseExplode`), needs `cost x 2` mana and a corpse near a pack (`pickExplosionCorpse`).

## Corpse Explosion points (card necrocorpse)

NOT changed. The proposal (5 points to 1 in the default table, SME delta 1 of the readiness audit) needs a scenario pair first (corpse-free fallback held by Teeth, Raise Skeletal Mage and Clay Golem taking the points) and the SME numbers are U. No table edit without that pair; the card stays open.

## Not verified live

No live Necromancer fight read these numbers back. Checklist: a lab Necromancer fighting a pack; the log lines `curse` and `corpse-explosion` and the minion count of the minute line (card necrosummoninv, `minions=N`) show whether the caps and reach hold.
