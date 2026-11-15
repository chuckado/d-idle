#!/bin/bash
# decomp29 (worker host): game.exe's element min / max / length (FUN_00644d50 / e40 / f20) vs d2d's elem_damage, all class skills.
# Run: tools/emu/run-on-worker.sh --ttl 900 dc29_cmp.sh   (prints counts; d2d = the fork's pr/skilldmg build, post-#146)
set -u
here=$(cd "$(dirname "$0")" && pwd)
S=${D2_SCRATCH:-/tmp/d2agent-$USER}
mkdir -p "$S"
bash "$here/run-py.sh" dc29_cases.py > "$S/dc29_cases.tsv" || exit 2
bash "$here/run-py.sh" skill_damage.py game "$S/dc29_cases.tsv" > "$S/dc29_game.tsv" || exit 3
T=${D2D_TEST_GAME:-$HOME/d2d-lab/pr/skilldmg/build/tests/test_game}
export D2_MPQ_DIR=${D2_MPQ_DIR:-${D2_EMU_DIR:-$HOME/d2d-lab/emu}/mpq}
export D2_PATCH_INSTALLER=${D2_PATCH_INSTALLER:-${D2_EMU_DIR:-$HOME/d2d-lab/emu}/patch/LODPatch_114d.exe}
nice "$T" --skill-damage "$S/dc29_cases.tsv" | grep -E '^[0-9]+ ' > "$S/dc29_d2d.tsv"
wc -l "$S/dc29_game.tsv" "$S/dc29_d2d.tsv"
diff "$S/dc29_game.tsv" "$S/dc29_d2d.tsv" | grep -c '^<'
diff "$S/dc29_game.tsv" "$S/dc29_d2d.tsv" | head -20
md5sum "$S/dc29_cases.tsv" "$S/dc29_game.tsv"
