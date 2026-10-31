# navsoak3: third hidden soak, current main after TOWNWEDGE and NOWALL (2026-10-31)

> status: finished 2026-10-31 12:22 (exit 0, 240 wall minutes). Final numbers: [navsoak3-summary.md](../navsoak3-summary.md).

## Built from
origin/main 35ce4c2 (bot binary `bin/d2bot.navsoak3`, sha256 prefix dbcfbd9eef92, d2d `d2d.lab6`). Proof, run on origin/main:
- `git grep -n badEdgeTTLSec origin/main -- internal/env`: `tuning_nav.go:121` (`nav.badEdgeTTLSec`).
- `git grep -n stallWarnMin origin/main -- internal/env`: `tuning_bot.go:125` (`bot.stallWarnMin`).
- `git grep -n recoveryKitGate origin/main -- internal/resource internal/env`: `internal/resource/tuning.go:130`.
- `git grep -n tpScrollPotions origin/main -- internal/env`: `tuning_town.go:110` and `:112` (`potionsFirst`, `tpFirstLevel` text), knob at line 113.

## What it is
Recipe of [navsoak2](../navsoak2/README.md): fresh uncheated Sorceress, seed 5, `--backend d2d --managed --d2d-class 4 --toggle autoloot=off,autoparty=off --d2d-args "--lockstep" --goals campaign`, `D2D_RUN=on`, nice 10, no viewer, no stream. Character `Navy4`, instance `navsoak3`, supervisor `bin/navsoak3-sup.sh` (up to 3 restarts with `--d2d-class -1 --clean-start`), started with `D2D_AGENT=navsoak3 bin/agentrun.sh navsoak3 --ttl 15000 -- bin/d2dslot.sh run -w 600 -t 15000 -- bash bin/navsoak3-sup.sh`. Other slots ran m2early baselines at the same time.

## Where (worker host, relative to `~/d2d-lab/navsoak3/`)
`logs/supervisor.log` (restart reasons), `logs/bot-0.log` (then `bot-1.log`..`bot-3.log`; last line `exit <code>`; exit 5 = the stall ladder asked for a supervisor restart), `data/` (saves), `bin/`.

## Analysis (read-only)
```
cd ~/d2d-lab/navsoak3
cat logs/supervisor.log; bin/agentreap.sh --list
grep '"msg":"minute"' logs/bot-*.log | tail -3     # clvl, kills, deaths, gold, area, step
grep '"msg":"pace"' logs/bot-*.log | tail -1       # virtual vs wall
grep -c -E 'campaign stalled|portal did not lead' logs/bot-*.log
grep -E 'campaign stalled|no eligible reachable area' logs/bot-*.log | tail
grep -E 'died|death|failsafe|panic|runner loop made no progress' logs/bot-*.log | tail
grep -o -E '"step":"Quest[^"]*"' logs/bot-*.log | sort | uniq -c | sort -rn | head
```

## Stop
`tools/agentreap.sh navsoak3` (worker host; takes supervisor, bot and d2d down, frees the slot). Touch no other agent's entries.

## Compare with navsoak and navsoak2 ([navsoak-compare.md](../navsoak-compare.md))
How far Act 1 gets in 240 minutes (virtual time at the end); the first death (clvl, area, virtual time); the stall ladder firing (`campaign stalled` lines, restart reasons in `logs/supervisor.log`); `portal did not lead` recoveries; quests done: Den, Blood Raven (250 gold gate: is gold the limit?), Dark Wood; whether the `no eligible reachable area` wall still shows up; whether the town nav wedge (old end state at wall min 179 and 45) is gone.

## First 16 minutes (wall; checked 08:38 worker clock)
clvl 6, 680 kills, 0 deaths, virtual 3.5 h (ratio 13x), gold 152, step "Quest: fighting (Level up to 9)", log restarts counter 2 (in-process), no supervisor restart.
