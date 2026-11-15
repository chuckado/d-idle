# Before/after of the survival replay for card powerwire (docs/survival/power.md): run on the worker host, from the repo root.
# usage: sh tools/levelsim/replay_compare.sh [INSTALL_DIR]
set -e
INSTALL=${1:-/gamedata}
L=internal/survival/testdata/replay-sandbox-clvl11to14.txt
go build -o ${TMPDIR:-/tmp}/ls_pw ./tools/levelsim/cmd/levelsim
run() { name=$1; shift
  ${TMPDIR:-/tmp}/ls_pw -controller $L -install "$INSTALL" -arch bruiser -dial balanced "$@" > ${TMPDIR:-/tmp}/pw_$name.txt
  echo "## $name: $*"
  grep -E '^pass rule|^minutes |^k_hero' ${TMPDIR:-/tmp}/pw_$name.txt
  awk '/^== controller/{exit} $8=="bad"||$8=="other"||$8=="good"{ if($11=="risky") r+=$4; else if($11=="too") t+=$4; else if($11=="safe") s+=$4 } END{printf "stay minutes: safe %.0f risky %.0f too_hard %.0f\n", s,r,t}' ${TMPDIR:-/tmp}/pw_$name.txt
  awk '/^== controller/{c=1} c' ${TMPDIR:-/tmp}/pw_$name.txt | grep -c "Cave Level 1" | sed 's/^/controller rows naming Cave Level 1: /'
}
run before
run build_noweapon -build paladin-normal30
run build_w6_12 -build paladin-normal30 -weapon 6-12 -ar 150
run build_w10_20 -build paladin-normal30 -weapon 10-20 -ar 200
