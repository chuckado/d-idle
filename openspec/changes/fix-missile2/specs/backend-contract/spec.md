# Backend contract specification

## MODIFIED Requirements

### Requirement: The d2d state source lists missiles
The d2d state source SHALL fill `Snapshot.Missiles` from `debug missiles` with one unit per moving missile (speed above 0; Class, OwnerType, OwnerID, subtile X Y) and SHALL exclude speed 0 hit-effect rows. A d2d that answers `err` SHALL be asked once. Source: cards fix-missile, fix-missile2; internal/d2dsrc/missiles.go; tests TestMissilesListed, TestMissilesRealVerb.

#### Scenario: A shooter's missile heads at the player
- **WHEN** the listing holds a missile owned by monster 752 and one owned by the player
- **THEN** the snapshot has two missiles, the first with OwnerType 1 and OwnerID 752, the second with the player's own id, and the combat dodge treats only the first as hostile

#### Scenario: A hit effect is listed
- **WHEN** the listing holds a speed 0 row with owner -1 next to a moving shaman bolt
- **THEN** the snapshot holds only the bolt

#### Scenario: Stock d2d without the verb
- **WHEN** the d2d answers `err` to `debug missiles`
- **THEN** every snapshot has no missiles and the verb is sent once
