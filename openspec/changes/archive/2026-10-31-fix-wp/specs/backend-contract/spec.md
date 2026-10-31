# Backend contract specification

## ADDED Requirements

### Requirement: The waypoint unit is found on d2d
On the d2d backend the state reader SHALL give every npcs row named "Waypoint" an object class that `d2map.IsWaypointObject` accepts, so the navigator's nearest-waypoint lookup finds the unit of every act. A named object that is not a waypoint SHALL keep class -1. Source: docs/research/waypoint-trace.md W1 (295 of 295 failures were `unit not found`).

#### Scenario: Waypoint row
- **WHEN** the npcs list holds a row named "Waypoint"
- **THEN** the snapshot has an object of a waypoint class at its position and the navigator returns it as the nearest waypoint

#### Scenario: No waypoint near
- **WHEN** the snapshot has no waypoint object
- **THEN** the waypoint lookup reports unit not found

### Requirement: Waypoint rows are clicked at their panel centres
The waypoint verb SHALL click the tab of the target act and then the row at the centre of game.exe's hit box: tab centres x 112, 176, 240, 304, 368 at y 77, row centres y 135, 171, 207, 243, 280, 316, 352, 388, 424 at x 230, for rows 0 to 8 of the act's Levels.txt waypoint order. Source: docs/research/waypoint-trace.md W2.

#### Scenario: Row 8 of an act
- **WHEN** the verb travels to the ninth waypoint of an act
- **THEN** it clicks that act's tab and then (230, 424)
