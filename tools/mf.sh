#!/bin/sh
export PATH=$PATH:/opt/go/bin
[ -e /tmp/d2mpq ] || ln -s /data/d2d-lab/data /tmp/d2mpq
go test -count=1 -p 1 -v -run "LooseScroll" ./internal/town ./internal/loot 2>&1 | grep -E "^(---|ok|FAIL)"
echo "== must-fail (cap off)"
sed -i 's/ && have < n.l.e.Tune.Town.LooseScrollKeep//' internal/loot/loot.go
go test -count=1 -p 1 -run "LooseScrollsCapped" ./internal/loot 2>&1 | grep -E "^(---|ok|FAIL)"
git checkout internal/loot/loot.go 2>/dev/null; true
