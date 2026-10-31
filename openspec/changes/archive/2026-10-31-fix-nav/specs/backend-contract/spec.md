# Backend contract specification

## ADDED Requirements

### Requirement: A bounced step away from a cave mouth changes direction
The navigator SHALL, when a step away from a tile exit ends in another level, step in a different direction from that exit on the next attempt instead of repeating the same step. Source: docs/probes/postfix-soak1.md (6 -> 20, 16 times).

#### Scenario: Bounce then retry
- **WHEN** the character starts a walk inside the trigger of an exit it does not want and the first step away fires that exit
- **THEN** the next step away from the same exit is rotated away from the first direction

#### Scenario: Memory expires
- **WHEN** ten minutes have passed since the bounce
- **THEN** the step away is the straight line away from the exit again

### Requirement: Touched waypoints reach the planner
A waypoint the navigator touched or travelled to SHALL count as active for the level picker and the travel planner of the same character and difficulty. Source: docs/probes/postfix-soak1.md (Black Marsh touched at 15:23, still `no activated waypoint`).

#### Scenario: Touched by the town trip
- **WHEN** the waypoint-to-town trip opened the waypoint of area 6
- **THEN** the next activeWaypoints read includes area 6
