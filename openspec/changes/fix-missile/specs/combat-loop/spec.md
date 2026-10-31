# combat-loop Specification

## ADDED Requirements

### Requirement: The cast gap covers the cast animation
The default of the `combat.castGapMs` knob SHALL be at least the cast time of the early Sorceress bolts (560 ms), so the fighter does not re-issue a cast the game is still animating. Source: card fix-missile; docs/d2d-casts-real.md (V-d2d: 1.6 to 1.75 commands per cast at 450 ms, 1.02 to 1.04 at 650 ms); internal/env/tuning_combat.go; test TestCastGapCoversEarlyCast.

#### Scenario: Default tuning
- **WHEN** the default combat tuning is read
- **THEN** CastGapMs is at least 560
