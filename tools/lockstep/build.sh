#!/bin/sh
# Build the P0 probes (32-bit Win32 PE) with clang + lld against Wine's headers
# and import libraries; no mingw-w64 needed. Run on the worker host:
#   labrun orion -- tools/lockstep/build.sh      (from inside the worktree)
set -eu
cd "$(dirname "$0")"
OUT=${OUT:-build}
mkdir -p "$OUT"
W=/usr/lib/wine/i386-windows
CF="--target=i686-windows-gnu -ffreestanding -fno-stack-protector -O1 -fno-builtin -Wall -I/usr/include/wine/msvcrt -I/usr/include/wine/windows"
LF="-fuse-ld=lld -nostdlib -L$W"
clang $CF $LF -shared -Wl,-e,_DllMain@12 -Wl,--out-implib,$OUT/libfakeijl.a hook.c -o $OUT/fakeijl.dll -lkernel32 -lwinmm -luser32
clang $CF $LF -shared -Wl,-e,_DllMain@12 -Wl,--out-implib,$OUT/libgamedll.a gamedll.c -o $OUT/gamedll.dll -lkernel32 -lwinmm -luser32
clang $CF $LF -Wl,-e,_mainCRTStartup -L$OUT probe.c -o $OUT/probe.exe -lfakeijl -lgamedll -lkernel32 -lwinmm -luser32
ls -l $OUT/*.exe $OUT/*.dll
