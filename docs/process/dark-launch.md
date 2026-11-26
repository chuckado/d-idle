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
