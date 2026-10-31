# Backend contract specification

## ADDED Requirements

### Requirement: Deadlines of game waits are game time under lockstep
The d2d verbs that wait for the game up to a timeout (the NPC approach) SHALL measure that timeout on the clock of the run: the Pacer's virtual time under lockstep, the wall clock on a free-running d2d. Source: docs/probes/waits-audit.md.

#### Scenario: Lockstep NPC approach that never opens
- **WHEN** the NPC approach has a 6 second timeout under lockstep and the menu never opens
- **THEN** it gives up after about 6 virtual seconds of stepped pauses, however little wall time that took

#### Scenario: Free-running approach
- **WHEN** the clock is the wall clock
- **THEN** the deadline is the wall time, as before

### Requirement: One call installs the lockstep-aware waits
The d2d backend SHALL offer one call that detects lockstep and installs the pause and the deadline clock of the verbs, and every d2d program that calls those verbs (the bot, the lab shop tool) SHALL use it. Source: docs/probes/waits-audit.md.

#### Scenario: Install under lockstep
- **WHEN** the call runs against a lockstep d2d
- **THEN** a later pause steps the game and the deadline clock advances by the paused time

#### Scenario: Install on a free-running d2d
- **WHEN** the call runs against a free-running d2d
- **THEN** a later pause is a wall sleep that sends no step command and the deadline clock is the wall clock
