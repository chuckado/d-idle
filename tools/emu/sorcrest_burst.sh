#!/bin/sh
# SORCREST: run the burst oracle on the worker host. usage: sorcrest_burst.sh fireball|glacialspike [SPAN]
UP=$HOME/d2d-lab/pr/skmiss
cp "$(dirname "$0")/sk_rest_burst.py" $UP/tools/emu/
cd $UP/tools/emu || exit 1
export D2_MPQ_DIR=/tmp/ug-emu D2_PATCH_INSTALLER=$HOME/Downloads/LODPatch_114d.exe
exec nice ${PY:-$HOME/d2d-lab/pr/walk2-venv/bin/python} sk_rest_burst.py "$@"
