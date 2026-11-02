## MODIFIED Requirements

### Requirement: Gate 1 deaths and progression
A Gate 1 run MUST follow the death rule (death-rule spec, `d2dsoak report`) and SHALL have no band with a FAIL verdict under "Gate 1 XP per clvl band", against the baseline files of the fidelity-gate spec, without copying their numbers here. No single state may last more than 10 minutes without progress. Source: G1 criteria 2 and 3; ADR 0009 item 2; docs/research/baseline-audit.md section 5.

#### Scenario: XP below the d2d baseline
- **WHEN** a band has a hard verdict and the run's XP/h in it is below the band baseline by more than the band tolerance
- **THEN** the run is a FAIL even with no crash

#### Scenario: Whole-run mean
- **WHEN** a run crosses several bands
- **THEN** its whole-run XP/h is shown as information only and is never compared with one baseline mean

## ADDED Requirements

### Requirement: Gate 1 XP per clvl band
The XP/h of a run SHALL be computed per clvl band (1-4, 5-7, 8-10, 11-14, 15-20, inclusive), from the first minute line at or above the band's lower clvl to the first line above its upper clvl, so the bands tile the run, and compared with the band baseline: the mean of the baseline runs with at least 0.5 h in that band (`d2dsoak bands`). Source: docs/research/baseline-audit.md sections 4b, 6 item 5 and the After section.

#### Scenario: Fresh run crossing the bands
- **WHEN** a fresh run reaches clvl 14
- **THEN** the report lists one XP/h row per band it entered, each against its own baseline

#### Scenario: Band entered briefly
- **WHEN** the run spent less than 0.5 h in a band
- **THEN** that band's row is ADVISORY whatever its value

### Requirement: Band baseline sample
A band verdict SHALL be ADVISORY, never FAIL, when fewer than 3 baseline runs have at least 0.5 h in the band. Source: docs/research/baseline-audit.md section 3 (the interval of a 2-run mean is wider than the mean); `d2dsoak bands`.

#### Scenario: Band with one baseline run
- **WHEN** the baseline holds one run in clvl 5-7
- **THEN** the clvl 5-7 verdict is ADVISORY and does not decide Gate 1

### Requirement: Band tolerance
With 3 or more baseline runs in the band, the verdict SHALL be FAIL only when the run's XP/h is below the band mean by more than 1.96 x sqrt(CV^2 / k + CV^2 / n) of the mean, k the d2d runs pooled (1 for Gate 1), n the baseline runs, CV the larger of the band's own CV and 0.14; above the mean it passes. Source: docs/research/baseline-audit.md section 3 (single-seed noise 14 percent).

#### Scenario: Band with four baseline runs
- **WHEN** four baseline runs have at least 0.5 h in clvl 11-14 and a Gate 1 run's XP/h there is below the mean by more than the tolerance
- **THEN** the clvl 11-14 verdict is FAIL

#### Scenario: Faster than the baseline
- **WHEN** the run's XP/h in a hard band is above the band mean
- **THEN** the verdict is PASS whatever the size of the gap
