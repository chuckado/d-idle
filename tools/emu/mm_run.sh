#!/bin/sh
# MONSTERMISSILE (the worker host): copy mm_launch.py into the scratch upstream checkout and run it.
#   tools/emu/mm_run.sh [UPSTREAM] TABLES_DIR
up=${1:-$HOME/d2d-lab/pr/skmiss}; t=$2
here=$(cd "$(dirname "$0")" && pwd)
cp "$here/mm_launch.py" "$up/tools/emu/mm_launch.py"
cd "$up/tools/emu"
mkdir -p "$HOME/mmtab/emu/bin"; ln -sf "$HOME/d2d-lab/pr/tome-scratch/game.exe" "$HOME/mmtab/emu/bin/game.exe"; for f in "$HOME"/d2d-lab/data/*; do ln -sf "$f" "$HOME/mmtab/emu/$(basename "$f")"; done
export D2_MPQ_DIR=${D2_MPQ_DIR:-$HOME/mmtab/emu} D2_PATCH_INSTALLER=${D2_PATCH_INSTALLER:-$HOME/Downloads/LODPatch_114d.exe}
exec nice "${PY:-$HOME/d2d-lab/pr/walk2-venv/bin/python}" mm_launch.py "$t"
