# SPDX-License-Identifier: GPL-3.0-or-later
"""SORCREST: SKMISS's missile-hit oracle (sk_missile_hit.py) for the two Sorceress
missile skills SKMISS did not cover: Fire Ball (47, fire) and Glacial Spike
(55, cold).

    python sk_rest_hit.py gen N SEED > cases.tsv
    python sk_rest_hit.py game cases.tsv > game.tsv
    test_missile_hit --missile-hit cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
    diff game.tsv d2d.tsv

Same cases format and d2d test binary as sk_missile_hit.py (the d2d side picks
the skill row by id, so no rebuild). Synergies from Skills.txt EDmgSymPerCalc:
Fire Ball 36 + 56 (Fire Bolt, Meteor), Glacial Spike 39 + 45 + 64 (Ice Bolt,
Ice Blast, Frozen Orb). Run it next to upstream's tools/emu with
sk_missile_hit.py copied beside it.
"""
import sk_missile_hit as base

base.SKILLS = {47: "fire", 55: "cold"}
base.SYNERGIES = {47: (36, 56), 55: (39, 45, 64)}

if __name__ == "__main__":
    base.main()
