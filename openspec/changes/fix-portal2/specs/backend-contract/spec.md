# Backend contract specification

## ADDED Requirements

### Requirement: The d2d portal verb never clicks with a dead context
The d2d town portal verb SHALL check its context before every try and, when it is cancelled, SHALL return an error that names the cancel cause without pressing a key or clicking. Source: docs/research/town-portal-trace.md (134 of 778 failures were three tries in 10 ms with a cancelled context).

#### Scenario: Cancelled context
- **WHEN** the portal verb is called with an already-cancelled context
- **THEN** no key and no click is made and the error names the cause

### Requirement: The d2d portal verb clicks only when the player is idle
On d2d the portal verb SHALL NOT click while the player is in hit recovery, attack, block, cast, throw, kick, skill, sequence or knockback mode, because d2d drops the click silently. It SHALL wait up to 2 seconds for an idle mode and then fail with a busy error. Source: docs/research/town-portal-trace.md answer 3.

#### Scenario: Hit recovery ends within the wait
- **WHEN** the player is in hit recovery when the verb starts and becomes idle within 2 seconds
- **THEN** the click happens after the player is idle

#### Scenario: The player stays busy
- **WHEN** the player is still attacking after 2 seconds
- **THEN** no click is made and the error says the player is busy

### Requirement: The d2d portal verb succeeds only on a portal object
The d2d portal verb SHALL report success only when a town portal object (class 59) is within 25 subtiles of the player, never because the scroll left the pack. When the source was spent and no object appeared, it SHALL re-read the pack and charges and retry with the next source; with no source left it SHALL return the no-portal error without another click. Failure lines SHALL name the cause (empty, busy, cancelled, no object after a dropped click, no object after a cut cast).

#### Scenario: Cast cut by a hit
- **WHEN** the first scroll is spent and no portal appears while a second scroll is in the pack
- **THEN** the second scroll is clicked and its portal counts as success

#### Scenario: Last source spent by a cut cast
- **WHEN** the only scroll is spent and no portal appears
- **THEN** no second click is made and the error is the no-portal error
