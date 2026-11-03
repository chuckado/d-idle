#!/bin/bash
# verifydmg3: compile tools/emu/w3_d2d.cpp against the rules headers of base, r2 and r3 (exported by w3_git.sh), nice 19, one at a time.
#   tools/emu/run-on-worker.sh w3_build.sh
set -eu
S=${D2_SCRATCH:?}
here=$(cd "$(dirname "$0")" && pwd)
for l in base r2 r3; do
	IN=""
	for d in components/rules components/d2s components/txt components/log; do IN="$IN -I$S/t_$l/$d"; done
	DEF=""; [ $l = base ] && DEF=-DW3_BASE; [ $l = r2 ] && DEF=-DW3_R2
	nice -n 19 c++ $IN $DEF -O1 -std=c++26 -Wno-missing-field-initializers -o "$S/w3_d2d_$l" "$here/w3_d2d.cpp"
	echo "built w3_d2d_$l"
done
