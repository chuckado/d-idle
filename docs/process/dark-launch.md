# Dark-launch rule: built behind a default-off knob means a ledger row

> status: current (card darklaunch, 2026-11-25). Why: the arbiter, `town.plan` and `quests.boardSM` were built, shipped at 0 and never enabled (docs/research/fable-logic-review-2026-11-25.md section 2.5).

## Rule
1. A feature that ships behind a knob with default 0 because it is "built but not enabled" MUST, in the same PR:
   - put `[dark-launch]` in the knob's description (it shows in docs/knobs.md);
   - add a row to `data/darklaunch.json`: `name`, `built_by` (card and commit), `dark_since`, `why_off`, `enable_criteria` (measurable: "invariant counts for 7 days", "drill X green", "owner go"), `owner` (coordinator or owner), `review_by` (at most 14 days out), `status`, `criteria_met` (yes, no, U).
2. Knobs that are off by design (a user decision, a numeric "0 = off") need a row only when they carry the marker. Existing dark features are listed by name without the marker.
3. `status` is one of `dark`, `trial` (on sandboxes), `default-on`, `retired`. A `default-on` or `retired` row stays as history.

## Enforcement
`tools/ci/check.sh` runs `python3 -I tools/darklaunch/check.py`. It fails when a marked default-0 knob has no row, a row lacks a field, or a `dark`/`trial` row is past `review_by`. Tests: `python3 -I tools/darklaunch/test_check.py`.

## Review
The coordinator runs `python3 -I tools/darklaunch/report.py` weekly (name, age dark, criteria met, owner, due). Every item past due or with `criteria_met: yes` gets one decision: enable (flip the default, status `default-on`), extend (new `review_by` plus a changed `why_off` that says why), or retire (delete the code path or the knob, status `retired`).

## Standard enable rule (user decision 2026-11-26)
Flip on the sandboxes once the scenario, drill or replay passes, watch 24 h with the invariant monitor, and revert automatically on a hard breach (a death, a stall, a potion not drunk, a refusal-as-failure storm); otherwise the knob stays on and the row closes. No criterion longer than 24 h of live observation; longer evidence comes from scenarios and replays. The owner of each flip is the coordinator, due next session.

## How to flip (tool `tools/flip/flip.py`, card autorevert1)
1. Precondition: the knob has a `dark` row in `data/darklaunch.json` and its scenario, drill or replay passed. Never in a measured run (`D2BOT_RUN` set refuses); sandbox containers only; one open window per sandbox.
2. Hosts come from `local.env` through `tools/env.sh`: `D2_FLIP_RUN` (command prefix on the sandbox host, default `labrun $D2_LABRUN_HOST --`), optional `D2_FLIP_TUNING` (default `/data/instances/main/state/tuning.json` in the container), `D2_FLIP_BIN`, `D2_FLIP_STATE`. The container is named like the instance.
3. Flip: `python3 tools/flip/flip.py <instance> <knob>=<value> --why "scenario X green" --window 24h` (`--dry-run` prints every step and writes nothing; run it first). The tuning file is the store envelope `{version, data:{profile, overrides}}`: the knob goes into `data.overrides` (V: real sandbox file shape), the rest stays; the old file is the backup tag `bak-flip-<ts>`; the write is an atomic rename; the restart is the whole container (`D2_FLIP_RESTART=container`, default) or only the bot launcher (`bot`, U: not drilled). Then `--verify`: it waits (`D2_FLIP_VERIFY_S`, 240) until `ops status --why` shows the value under `knobs`; if it never does, the backup is restored and nothing is recorded. On success the row is `trial` with `trial_started`.
4. Watch (one watcher for all windows, under the process registry): `tools/agentrun.sh <agent> --ttl 90000 -- python3 tools/flip/flip.py watch`. It reads `GET /api/explain` data through `d2bot ops status --why` (`invariants`, Hard rows: potion-not-drunk, life-zero-alive, refusal-storm, xp-frozen, mode-no-exit, corpse-unresolved) every 60 s.
5. A hard counter above its baseline (taken right after the restart; a child restart inside the window rebaselines at zero) reverts: backup restored, child restarted, row `dark` with `revert_reason` and `reverted_at`, one line `REVERT ...`. Ten unreadable polls in a row count as a breach. A failed restore or restart prints `REVERT FAILED` and is retried every poll; the file is never left half set.
6. No breach to the end of the window: row `enabled` (on in the sandbox, `enabled_at`), backup removed, line `CLOSE ...`. The coordinator then flips the code default and sets `default-on`.
7. `python3 tools/flip/flip.py status` lists open windows. Tests (fake sandbox and clock): `python3 -I tools/flip/test_flip.py`. U: not yet run against a real container.
