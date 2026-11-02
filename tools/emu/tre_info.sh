#!/bin/bash
# treasure card: where the d2d drop-dump build lives and which sources the read checkout has (worker host, read-only).
D2=${D2:-$HOME/d2d-lab/pr/skmana}
find $HOME/d2d-lab -name drop-dump -type f -not -path "*/CMakeFiles/*" 2>/dev/null | head
ls $HOME/d2d-lab/pr/skresistab-d2d-build 2>/dev/null | head -5
ls $D2/tools/drop-dump
sed -n 1,40p $D2/tools/drop-dump/README.md
sed -n 55,75p $D2/tools/emu/emu.py
