# Backend contract specification

## ADDED Requirements

### Requirement: Field objects are listed and operable on d2d
The d2d state source SHALL list placed objects within 40 cells of the player (chests, shrines, the cage, the Tristram portal) with the objects.txt class of their preset row, and the d2d input SHALL operate one with `cmd interact <index>`, never a walk onto it. Source: drill H.obj.

#### Scenario: Chest near the player
- **WHEN** the level has a chest preset within 40 cells of the player
- **THEN** the snapshot lists an object with the chest's class and Interact on it sends `cmd interact` with the preset index

#### Scenario: Hidden or far preset
- **WHEN** a preset row is hidden or farther than 40 cells
- **THEN** it is not listed

### Requirement: Waits step the game under lockstep
The d2d backend SHALL make the pauses of the hover, menu and NPC-approach verbs advance the game by the requested time when d2d runs under lockstep, and SHALL keep them wall sleeps that end with the context on a free-running d2d. Source: docs/probes/drills/2026-10-31-h.txt (H.pause).

#### Scenario: Lockstep pause
- **WHEN** a pause of 4 seconds is requested under lockstep
- **THEN** the game is stepped by 4 seconds of ticks and almost no wall time passes

#### Scenario: Free-running pause
- **WHEN** a pause is requested on a free-running d2d
- **THEN** no step command is sent and the pause lasts the requested wall time
