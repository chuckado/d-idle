export PG_D2D_BIN=/data/build/devp15-f212b77a-d2d/apps/d2d/d2d
for c in "Paladin 11" "Sorceress 3" "Paladin 11"; do set -- $c
 echo "=== RUN $1 $2"
 tools/provingground/run.sh -id act1.forgotten.countess -class $1 -seed $2 -parallel 1 2>&1 | tail -4
 d=$(ls -dt provingground-out/*/ | head -1)
 grep -h -v d2dtrace $d/act1.forgotten.countess_$1_$2.bot.log | grep -E 'boss presets|"boss|preset .* not found|goto|nav mouth|Countess|quest board: (step|forgotten)|task step|summary"' | cut -c1-420 | tail -40
 cp $d/act1.forgotten.countess_$1_$2.bot.log /data/d2d-lab/provingground/cnt-$1-$2-$(date +%H%M).log
done
