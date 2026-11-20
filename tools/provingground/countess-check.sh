for c in "$@"; do set -- $c
 echo "=== RUN $1 $2"
 tools/provingground/run.sh -id act1.forgotten.countess -class $1 -seed $2 -parallel 1 2>&1 | tail -4
 d=$(ls -dt provingground-out/*/ | head -1)
 grep -h -v d2dtrace $d/act1.forgotten.countess_$1_$2.bot.log | grep -E '"msg":"(boss presets|boss run)"|nav mouth gate|quest board: step|task step' | cut -c1-900 | tail -30
 cp $d/act1.forgotten.countess_$1_$2.bot.log /data/d2d-lab/provingground/cnt-$1-$2-$(date +%H%M).log
done
