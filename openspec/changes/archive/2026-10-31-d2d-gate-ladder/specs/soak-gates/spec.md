## ADDED Requirements

### Requirement: Gate ladder on d2d
Release gates SHALL be a ladder on a hidden d2d slot, each rung started only after the one below passed on the same build: Gate 0 the live drill suite (every S and A primitive of docs/primitives.md passes in `cmd/d2drill`), Gate 1 a 2 h unattended soak with one injected fault, Gate 2 an 8 h soak (profile m24), Gate 3 the 24 h soak (profile m24). Cheated gold or levels never count. Source: ADR 0013; docs/review-strategy.md section 3 (d2d gate ladder); docs/probes/drills/README.md.

#### Scenario: Skipping a rung
- **WHEN** Gate 1 has not passed on a build
- **THEN** a Gate 2 run on that build does not count as a gate

#### Scenario: Drill failing
- **WHEN** one S or A primitive drill is FAIL or BLOCKED
- **THEN** Gate 0 is not passed and no Gate 1 soak starts

### Requirement: Gate 1 stability and save semantics
A Gate 1 run SHALL show zero unplanned process exits other than the injected fault, zero panics and zero manual interventions, the bot and its d2d alive at 2 h, and after a graceful stop a saved XP equal to the character XP. Source: translated from docs/review-strategy.md G1 criterion 1 (S).

#### Scenario: Clean two hour run
- **WHEN** a Gate 1 run ends at 2 h
- **THEN** the report shows 0 panics, no exit other than the injected one, and `saveXp == charXp` after the graceful stop

### Requirement: Gate 1 deaths and progression
A Gate 1 run MUST follow the death rule (death-rule spec, `d2dsoak report`) and SHALL earn true XP per hour at or above the d2d baseline at the same clvl, read from the baseline file named in the fidelity-gate spec (Early baseline fields; docs/d2d-baselines-early.md, docs/probes/navsoak3-summary.md for the pre-fix state), without copying its numbers here. No single state may last more than 10 minutes without progress. Source: translated from G1 criteria 2 and 3; ADR 0009 item 2.

#### Scenario: XP below the d2d baseline
- **WHEN** true XP/h is under the d2d baseline at the same clvl
- **THEN** the run is a FAIL even with no crash

### Requirement: Gate 1 belt and portal
A Gate 1 run SHALL have no window longer than 5 minutes with zero healing potions outside town, measured from the minute-line fields packHP, packMP and beltBoxes (pack plus belt), and at least one portal trip completed by town portal, not by the waypoint fallback. Source: translated from G1 criterion 4 (the old "bought and used" is now measurable); docs/beltfill.md.

#### Scenario: Empty field window
- **WHEN** packHP is 0 and no belt box holds a healing potion outside town for 5 minutes
- **THEN** the run is a FAIL

#### Scenario: Waypoint only
- **WHEN** every town return in the run used the waypoint fallback
- **THEN** the run is a FAIL on the portal criterion

### Requirement: Gate 1 route guard respected
A Gate 1 run MUST start no boss run while the route guard or the boss gate refuses (log `blocked: not ready`, `route guard:` floors of the merc-policy spec) and SHALL show fewer than 1 boss-run start per 30 s. Source: translated from G1 criterion 5.

#### Scenario: Refused boss
- **WHEN** the log shows `blocked: not ready` for a boss and a run for that boss starts afterwards without the gate opening
- **THEN** the run is a FAIL

### Requirement: Injected fault recovery on d2d
At about 60 minutes of Gate 1 exactly one fault SHALL be injected: the d2d process of that slot only is killed through the process registry (`tools/agentreap.sh`, never `pkill`, never a pid not started by the run). The supervisor MUST restart it within 3 minutes with exactly one restart counted, the character saved, and no process of the slot orphaned afterwards (`tools/agentreap.sh --list` and `--orphans`). Source: translated from G1 criteria 6 and 7; ADR 0013.

#### Scenario: Injected kill
- **WHEN** the slot's d2d is killed at about 60 minutes
- **THEN** the bot is back in game in at most 3 minutes, `restarts` rises by exactly 1, the saved XP does not drop below its value before the kill and no process of the slot is orphaned

#### Scenario: Double restart
- **WHEN** the kill causes 2 restarts
- **THEN** the run is a FAIL on this criterion

### Requirement: Gate 2 and Gate 3 on d2d
Gate 2 (8 h) and Gate 3 (24 h) SHALL be soaks of profile m24 on a hidden d2d slot, judged by the Gate 1 criteria that apply over their length, the m24 limits and floors of this spec, and `d2dsoak report` ending with `INVARIANTS: PASS` including the primitives invariant. Gate 3 additionally records the d2d memory curve and the periodic d2d restart interval. Source: docs/review-strategy.md "Next week" table (24 h soak); docs/d2d-container/risks.md step 10.

#### Scenario: Gate 2 with a violated invariant
- **WHEN** the Gate 2 report ends with `INVARIANTS: VIOLATED(...)`
- **THEN** the run is not a pass even if the m24 limits hold

## MODIFIED Requirements

### Requirement: G2 meaning and floors
The 8 h profile m24 run (Gate 2 on d2d, 24 h Gate 3) SHALL be called a stability and economy soak, not a progression soak, until Andariel and Act 2 work; it SHALL reach clvl 16, attempt Andariel at clvl 14 or higher, hire the merc and keep gold from staying flat. Source: docs/plan/plan-v2.md section 5 and section 10 (approved 2026-10-30); ADR 0013.

#### Scenario: Cheated run
- **WHEN** a run used cheated gold or cheated levels
- **THEN** it does not count toward any gate

### Requirement: Start guard
A soak start MUST be refused when free disk is under 10 GB, when the 1-minute load exceeds the backstop 28, or when `tools/capacity.sh --add <soak cost>` prints `NO`; there is no fixed slot cap. A multi-hour soak is also a big run (lab-rules spec). Source: docs/soak24.md section 2; docs/process/worker-capacity.md (code change to soak24.sh is a follow-up, S).

#### Scenario: Busy host
- **WHEN** the predicted idle+nice after adding the soak is under 25 %
- **THEN** `soak24.sh start` refuses with the capacity.sh output

#### Scenario: Many instances, quiet host
- **WHEN** three slots already run but the predicted idle+nice stays at or above 25 % with iowait under 40 %
- **THEN** the start is allowed

## REMOVED Requirements

### Requirement: G1 stability
**Reason**: G1 on the original Game.exe is retired as a release gate (ADR 0013); replaced by "Gate 1 stability and save semantics" on d2d. The Game.exe numbers stay historical in docs/g1-log.md.

### Requirement: G1 progression
**Reason**: replaced by "Gate 1 deaths and progression" against the d2d baseline; the Game.exe baseline (Soaky) stays historical.

### Requirement: G1 belt
**Reason**: replaced by "Gate 1 belt and portal", now measurable from the minute-line fields.

### Requirement: G1 kill recovery
**Reason**: the Game.exe kill (100 s, winedevice) is replaced by "Injected fault recovery on d2d" (3 min, d2d process, process registry).
