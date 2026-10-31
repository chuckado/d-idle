# Town basics specification

## ADDED Requirements

### Requirement: ToTown never reads a scroll while an open portal is in reach
`ToTown` SHALL enter a town portal object within `portalReuseDist` of the player on every try and SHALL read a scroll or tome only when no such portal is in reach. When every try used an open portal and none led to town, it SHALL take the no-scroll path (walk or waypoint) instead of casting. Source: drill D2.t (all 3 scrolls spent on casts beside a portal that could not be entered); a new cast only replaces the pair with the same kind of portal.

#### Scenario: First entry ignored
- **WHEN** a portal stands 10 subtiles from the player in the field with 3 scrolls in the pack and the first entry does nothing
- **THEN** ToTown enters the same portal on the next try, arrives in town and still holds 3 scrolls

#### Scenario: No portal in reach
- **WHEN** no portal object is within reach and the pack holds 2 scrolls
- **THEN** one scroll is read, its portal entered, and 1 scroll is left

#### Scenario: Open portal never leads
- **WHEN** the open portal in reach never takes the player and the pack holds 2 scrolls
- **THEN** no scroll is read and ToTown walks back from Blood Moor to the camp
