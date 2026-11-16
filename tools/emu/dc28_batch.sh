#!/bin/bash
# decomp28: worker host. ms_spawn.py over a list of Levels.txt ids, seed batches one after the other (each under 5 minutes).
# Usage: bash tools/emu/dc28_batch.sh OUT SEEDLO SEEDHI BATCH IDS [DIFF]    IDS: comma list; DIFF 0 Normal (default)
set -e
here=$(cd "$(dirname "$0")" && pwd)
out=${1:?out}; lo=${2:?lo}; hi=${3:?hi}; b=${4:-3}; lids=${5:?ids}; diff=${6:-0}
s=$lo
while [ "$s" -le "$hi" ]; do
  e=$((s+b-1)); [ "$e" -gt "$hi" ] && e=$hi
  nice -n 10 timeout 290 bash "$here/am_env.sh" py ms_spawn.py "$s-$e" $lids $diff >> "$out" || echo "# batch $s-$e failed" >> "$out"
  s=$((e+1))
done
echo "# done $lo-$hi" >> "$out"
