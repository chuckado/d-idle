#!/bin/sh
export PATH=$PATH:/opt/go/bin:$HOME/.local/bin
[ -e /tmp/d2mpq ] || ln -s /data/d2d-lab/data /tmp/d2mpq
go test -count=1 -p 1 ./internal/loot ./internal/town ./internal/scenario 2>&1 | tail -15
echo "== RED control"
tools/provingground/run.sh -id town.tp-scroll-hoard.red -class Paladin -seed 11 -parallel 1
echo "== GREEN"
tools/provingground/run.sh -id town.tp-scroll-hoard -class Paladin -seed 11 -parallel 1
tools/provingground/run.sh -id town.tp-scroll-hoard -class Sorceress -seed 3 -parallel 1
