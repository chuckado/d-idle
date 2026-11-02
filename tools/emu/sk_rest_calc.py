# SPDX-License-Identifier: GPL-3.0-or-later
"""SORCREST: every calc of the 30 Sorceress Skills.txt rows (ids 36..65) by game.exe's
own calc machine, vs d2d's eval_calc. nec_calc.py (necskill) with its row list swapped.

    python3 sk_rest_calc.py gen N SEED > cases.tsv
    python3 sk_rest_calc.py game cases.tsv > game.tsv
    test_sk_rest_calc --calc cases.tsv > d2d.tsv     # tools/emu/nec_test.cpp, built as test_sk_rest_calc
    diff game.tsv d2d.tsv

Run next to upstream's tools/emu with nec_calc.py (and skill_damage.py) copied beside it.
Limit (as nec_calc.py): the stat getter returns 0 for every stat but clvl and the masteries.
"""
import nec_calc

nec_calc.NEC = list(range(36, 66))

if __name__ == "__main__":
    nec_calc.main()
