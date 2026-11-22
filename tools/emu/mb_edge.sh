#!/bin/sh
# decomp43: the exit codes of monster_block.py compare / game / gen on bad input (builder). Needs /data/d43/r from the oracle run.
export D2_MPQ_DIR=/data/prfix6-emu/gd D2_PATCH_INSTALLER=/data/p565-emu/LODPatch_114d.exe D2_BUILD=/data/build/cr3
M=${1:?path of tools/emu with monster_block.py}
export PYTHONPATH=/data/d43/tools/emu:$M
R=/data/d43/r; T=/tmp/mbedge; mkdir -p $T; : > $T/empty
cd /data/d43/tools/emu
head -n 9 $R/cases.tsv > $T/c8.tsv; head -n 8 $R/game.tsv > $T/g8.tsv; head -n 8 $R/d2d-claim4.tsv > $T/d8.tsv
t() { python3 $M/monster_block.py "$@" >/dev/null 2>&1; echo "$* -> $?"; }
t compare $T/c8.tsv $T/g8.tsv $T/d8.tsv
t compare $T/c8.tsv $T/empty $T/d8.tsv
t compare $T/c8.tsv $T/g8.tsv $T/empty
t compare $T/empty $T/g8.tsv $T/d8.tsv
head -n 5 $T/d8.tsv > $T/dshort.tsv; t compare $T/c8.tsv $T/g8.tsv $T/dshort.tsv
sed '3s/.*/x y z/' $T/d8.tsv > $T/dbad.tsv; t compare $T/c8.tsv $T/g8.tsv $T/dbad.tsv
sed '3s/^[0-9]* /99 /' $T/d8.tsv > $T/did.tsv; t compare $T/c8.tsv $T/g8.tsv $T/did.tsv
sed '4s/ -\?[0-9]*$/ 77/' $T/d8.tsv > $T/dval.tsv; t compare $T/c8.tsv $T/g8.tsv $T/dval.tsv
t compare $T/c8.tsv $T/g8.tsv $T/missing
t
t game $T/empty
t gen extra
t cap extra
