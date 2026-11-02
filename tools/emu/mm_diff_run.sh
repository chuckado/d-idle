#!/bin/sh
# MISSILEDIFF (the worker host): run mm_diff.py from a scratch upstream checkout (default: the fork worktree wt-missilediff).
#   tools/emu/mm_diff_run.sh game TABLES_DIR N SEED OUTDIR | tools/emu/mm_diff_run.sh cmp OUTDIR D2D_OUTPUT
up=${UP:-$HOME/d2d-lab/fork/wt-missilediff}
here=$(cd "$(dirname "$0")" && pwd)
cp "$here/mm_launch.py" "$here/mm_diff.py" "$up/tools/emu/"
cd "$up/tools/emu"
mkdir -p "$HOME/mmtab/emu/bin"; ln -sf "$HOME/d2d-lab/pr/tome-scratch/game.exe" "$HOME/mmtab/emu/bin/game.exe"
export D2_MPQ_DIR=${D2_MPQ_DIR:-$HOME/mmtab/emu} D2_PATCH_INSTALLER=${D2_PATCH_INSTALLER:-$HOME/Downloads/LODPatch_114d.exe}
exec nice "${PY:-$HOME/d2d-lab/pr/walk2-venv/bin/python}" mm_diff.py "$@"
