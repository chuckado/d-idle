# Build catalogue (build text and stars)

> status: current (card buildcatalog). One file, `internal/plandata/build-catalog.json`, loaded by `internal/plandata/catalog.go`; the UI (card E6) reads it.

- A build is one real build, not one plan table: a Hammerdin owns several tables (`tables[]`). Its name, identity, summary, 3 pros, 3 cons and seven ratings live only here.
- Add a build: add an entry (copy one), list its plan tables (file name without `build-` and `.json`), write all seven ratings with `source: estimate`. A new plan table with no entry and no `uncovered` reason fails `TestBuildCatalog`.
- Limits (phone card): identity 80, summary 280, pro/con 90, rating note 100 characters; stars 1 to 5. The loader errors on any unknown class or table, never falls back.
- `estimate` = a veteran's opinion (the d2-sme draft, all unverified). `measured` = turned into stars from a named scenario, with the number in `note`. Never from a CHEATED character.
- To replace an estimate: run the scenario at least 3 times, read the number against the `thresholds` block in the same file (SME ballpark, unverified, expected to be re-set after the first measured runs), set `stars`, `source: measured` and write the number and run id in `note`.
- Small card: Defense, the two highest of the other six, then the lowest of the remaining four; ties go to the earlier key in the order leveling, farmingItems, farmingXP, bossing, defense, clearSpeed, idle. The detail view shows all seven.
- Accessibility text: `Defense: 4 of 5 stars, estimate` (Defense adds `(Normal 5, Hell 4)` where the file has `defenseByDifficulty`).
- `build-catalog.json` shares the `build-` prefix with the plan tables; `plandata.TableIDs` and `buildfile.PlanNames` skip it.
