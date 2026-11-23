set -e
. ~/d2d-lab/emu/env.sh
R=${EMURUN_WT:?set to the synced worktree}
W=$R/tools/emu
cd ~/d2d-lab/emu/scratch-emurun
export PYTHONPATH=$W:$EMU
time nice -n 10 $PY $W/block_len.py game $R/.tmp-emurun/d2d_block_before.tsv > game_block.tsv
wc -l game_block.tsv
