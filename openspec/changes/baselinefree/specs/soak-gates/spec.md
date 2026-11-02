## MODIFIED Requirements

### Requirement: Gate 1 deaths and progression
A Gate 1 run MUST follow the death rule (death-rule spec, `d2dsoak report`) and SHALL show true XP per hour per clvl band that is not FAIL under the "Gate 1 XP per clvl band" requirement, against the baseline files named in the fidelity-gate spec (docs/d2d-baselines-early.md, docs/probes/navsoak3-summary.md for the pre-fix state), without copying their numbers here. No single state may last more than 10 minutes without progress. Source: translated from G1 criteria 2 and 3; ADR 0009 item 2; docs/research/baseline-audit.md section 5 (the whole-run clause compared a clvl 1-14 run with a clvl 1-7 mean).

#### Scenario: XP clearly below the band baseline
- **WHEN** a band has a hard verdict (see the sample requirement) and the run's XP/h in it is below the band baseline by more than the band tolerance
- **THEN** the run is a FAIL even with no crash

#### Scenario: Whole-run mean
- **WHEN** a run crosses several bands
- **THEN** its whole-run XP/h is shown as information only and is never compared with a single baseline mean

## ADDED Requirements

### Requirement: Gate 1 XP per clvl band
The XP/h of a run SHALL be computed per clvl band, with bands 1-4, 5-7, 8-10, 11-14 and 15-20 (inclusive), each from the first minute line at or above the band's lower clvl to the first line above its upper clvl (the level-up out of the band belongs to the band, so the bands tile the run), and compared with the band baseline: the mean over the baseline runs that spent at least 0.5 h in that band (`d2dsoak bands -baselines`). Source: docs/research/baseline-audit.md sections 4b and 6 item 5; bands chosen from the Fresha chain and a14 to a18 plus G2 data (docs/research/baseline-audit.md, after section).

#### Scenario: Fresh run crossing the bands
- **WHEN** a fresh run reaches clvl 14
- **THEN** the report lists one XP/h row per band it entered, each against its own baseline, and not one number for the whole run

#### Scenario: Band entered briefly
- **WHEN** the run spent less than 0.5 h in a band
- **THEN** that band's row is ADVISORY whatever its value

### Requirement: Band baseline sample and tolerance
A band verdict SHALL be ADVISORY (never a FAIL) when fewer than 3 baseline runs have at least 0.5 h in the band; with 3 or more it SHALL be FAIL only when the run's XP/h is below the band mean by more than tol = 1.96 x sqrt(CV^2 / k + CV^2 / n), k being the number of d2d runs pooled (1 for Gate 1), n the baseline runs in the band and CV the larger of the band's own CV and 0.14; a value above the mean passes. Source: docs/research/baseline-audit.md section 3 (single-seed noise 14 percent, interval of the baseline mean); implemented by `d2dsoak bands`.

#### Scenario: Band with one baseline run
- **WHEN** the baseline holds one run in clvl 5-7
- **THEN** the clvl 5-7 verdict is ADVISORY and does not decide Gate 1

#### Scenario: Band with four baseline runs
- **WHEN** four baseline runs have at least 0.5 h in clvl 11-14 and a Gate 1 run's XP/h there is below the mean by more than tol
- **THEN** the clvl 11-14 verdict is FAIL

#### Scenario: Faster than the baseline
- **WHEN** the run's XP/h in a hard band is above the band mean
- **THEN** the verdict is PASS whatever the size of the gap
