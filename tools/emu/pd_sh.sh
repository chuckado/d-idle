#!/bin/bash
# playerdef: read-only grep inside the upstream read checkout on the worker host.
#   labrun orion -- bash tools/emu/pd_sh.sh PATTERN [path under the checkout]
cd ~/d2d-lab/pr/palskill || exit 2
exec grep -rn -- "$1" "${2:-docs}"
