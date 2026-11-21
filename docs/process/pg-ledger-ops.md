# Proving Ground results ledger: how it is turned on (card pgledgeron)

> status: current, 2026-11-21. Code of the ledger line and the status page: `internal/scenario/ledger.go`, `cmd/d2scenario/ledger.go` (card pgledger, TASK-743). Design: `docs/design/fable-review-2026-11-21/proving-ground-gate.md` section 5; owner answer Q6 (the ledger lives on the board host next to the status files). The ingest command, the pull script and their timer are in the private ops repo and are described here by role.

## Path of one result

1. `d2scenario exec-one` writes one ledger line per run into `<out>/ledger.jsonl` on the host that ran it (the spool: a lost labrun pull cannot lose it) and, when `PG_LEDGER_CMD` is set, pipes the same line into that command.
2. The builder wrapper `tools/provingground/run.sh` sources `tools/provingground/pgenv.sh`: it reads host config from an env file on the builder (`PG_ENV`, default `/data/pg/env`, may set `PG_LEDGER_CMD`; the address of the board host is host config and is never in this repo) and fills `PG_EVENT` (default `manual`; the scheduler sets `merge`, `nightly` or `smoke`), `PG_BRANCH` (`agent/<worktree name>`, from the mirror path), `PG_JOB` (the labrun job id). A value the caller already set wins. `PG_CARD` and `PG_LANE` are passed through when set.
3. The board host holds the ledger (`$D2_PG_LEDGER`, default next to the agent status files) and one ingest command: validate, refuse a line without provenance or setup tag, append under a lock, skip a run id it has. The builder has no credential for the board host (V: none was created for this card; a push needs one), so the board host PULLS: a 10-minute user timer reads every `provingground-out/*/ledger.jsonl` of the last 3 days from the builder and feeds each file to the ingest command (idempotent, a refused file is reported and does not stop the others). A host that does have a credential can set `PG_LEDGER_CMD` in its env file and push as well; both paths are safe together.
4. The status page `docs/process/proving-ground-status.md` is generated from the ledger by `d2scenario ledger status-doc` (ops wrapper `pg-status-doc.sh`); it is committed like any generated file, by hand now and by the nightly job later (card pgnightly).

## State on 2026-11-21

- V: `tools/provingground/pgenv_test.sh` (3 checks) and the ops tests `test_pg_ledger_pull.sh` (pull, idempotent re-pull, a spool without provenance refused and nothing of it written) and `test_pg_ingest.sh`; the pull timer is installed on the board host.
- U: no scenario has run on the builder since the wiring, so the live ledger holds no line yet and the status page shows no runs. The first builder run is the proof (checklist below). The builder was unreachable by ssh for a long stretch of this session (memory at its cap), so the pull itself has not touched the live builder.
- Nothing is EARNED yet (no earned-save card); runs BLOCKED before exec-one are not ledgered.

## Checklist for the first run (needs the builder and a hidden-instance-free scenario)

1. `labrun BUILDER --lane long --detach --notify -- tools/provingground/run.sh -id town.resupply -class Paladin -seed 11 -parallel 1` from a worktree on this change.
2. After the job: `<mirror>/provingground-out/<stamp>/ledger.jsonl` has one line with `event` manual and `branch` agent/<worktree>.
3. Within 10 minutes the board host's `pg-ingest.py --count` increases by one; run `pg-ledger-pull.sh` by hand to skip the wait.
4. `go run ./cmd/d2scenario ledger status-doc -in <ledger> -out docs/process/proving-ground-status.md` and commit the page.
