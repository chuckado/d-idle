# Backend contract specification

## ADDED Requirements

### Requirement: The d2d state source lists missiles
When the d2d offers the `debug missiles` listing, the d2d state source SHALL fill `Snapshot.Missiles` with one unit per missile (Class = missiles.txt id, OwnerType and OwnerID of the shooter, X and Y in subtiles), as Game.exe's reader does. A d2d that answers `err` SHALL be asked once and never again. Source: card fix-missile; internal/d2dsrc/missiles.go (FillMissiles); tests TestMissilesListed, TestMissilesAbsent.

#### Scenario: A shooter's missile heads at the player
- **WHEN** the listing holds a missile owned by monster 752 and one owned by the player
- **THEN** the snapshot has two missiles, the first with OwnerType 1 and OwnerID 752, the second with the player's own id, and the combat dodge treats only the first as hostile

#### Scenario: Stock d2d without the verb
- **WHEN** the d2d answers `err` to `debug missiles`
- **THEN** every snapshot has no missiles and the verb is sent once
