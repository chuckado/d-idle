# Stash from anywhere (backlog idea, not built)

> status: backlog idea, user decision 2026-11-14. Default stays: stash and take need the hero at the chest (the strict item-verb rule, docs/design/item-verbs.md). This page records the idea so it is not lost.

User decision (2026-11-14, answering the question from the character-screen design pass): stash from anywhere ("virtual stash") is NOT the default. It is added to the backlog as a possible unlockable feature.

What it would be: stash and take (and gold deposit and withdraw) work from the watch UI while the hero is anywhere, instantly, with no trip to the chest.

Why it is not the default: it changes a game rule (the real game needs the stash panel open next to the chest), so it trades fidelity for convenience. It would be an unlock-ladder reward, gated like the order queue (`unlockorderqueue.go` is the pattern), and always visibly marked as a convenience feature.

Open design points for whoever picks it up: the unlock trigger and tier; how the UI marks items that moved without a chest visit; whether the d2d fork verb needs an explicit flag (the stash refusal rule is `far` today); interaction with the loot-rules engine (stash actions) and with measured runs (the CRUTCH rule: a convenience that bypasses game rules must be declared in the run record and must not count toward gates or baselines).
