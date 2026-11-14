set -e
. ~/d2d-lab/emu/env.sh
W=${EMURUN_WT:?set to the synced worktree}/tools/emu
mkdir -p ~/d2d-lab/emu/scratch-emurun && cd ~/d2d-lab/emu/scratch-emurun
export PYTHONPATH=$W:$EMU
nice -n 10 $PY $W/fl_gold.py gen > gen_cases.tsv
wc -l gen_cases.tsv
time nice -n 10 $PY $W/fl_gold.py game gen_cases.tsv > game_gold.tsv
wc -l game_gold.tsv
