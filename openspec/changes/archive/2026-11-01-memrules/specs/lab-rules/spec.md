## ADDED Requirements

### Requirement: Memory access
Reading game memory SHALL always be allowed; writing SHALL be allowed only for SETUP, INSTRUMENTATION, DIAGNOSIS and FIXTURE capture, through the logged tool (`internal/cheat` / `cmd/d2cheat` or a declared instrumentation module) with a reason and a ledger entry. Source: user decision 2026-11-02, ADR 0014; docs/cheats.md.

#### Scenario: Ad-hoc write
- **WHEN** code or an agent writes game memory outside the logged tool or declared module
- **THEN** it violates the rule and the result is not used

#### Scenario: Injected code
- **WHEN** a hook or DLL is injected into the game process
- **THEN** it is INSTRUMENTATION only, has the user's go for that hook set, is documented and logged

### Requirement: Crutch rule
In a measured run (gate, soak, baseline, fidelity) the bot MUST act only through the real input channel and nothing SHALL write gameplay state (life, mana, position, items, monsters, gold, XP, skills, RNG state) after the run start marker. Source: ADR 0014; docs/cheats.md.

#### Scenario: Setup before start
- **WHEN** a character gets level or gold writes before `d2cheat runstart <run>`
- **THEN** they are declared in the run record and the run may still count where the spec allows cheated setup

#### Scenario: Write after start
- **WHEN** `d2cheat ledger --since-run <run>` reports `writes after start: 1` or more
- **THEN** the run is reported as CHEATED and counted toward no gate, baseline or verdict

#### Scenario: Missing marker
- **WHEN** the run has no start marker in the ledger
- **THEN** the command fails and the record cannot claim `writes after start: 0`

### Requirement: Cheated runs and the writes line
A run with a gameplay-affecting write after its start, or whose bot path needs a write to proceed, SHALL be reported as CHEATED and SHALL NOT count toward a gate, baseline or verdict; each soak or gate record SHALL carry `writes after start: N` from `d2cheat ledger --since-run <run>`. Source: ADR 0014; docs/cheats.md.

#### Scenario: Records
- **WHEN** a soak or gate record is written
- **THEN** it contains the line `writes after start: N` taken from the ledger
