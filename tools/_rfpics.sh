set -e
o=/tmp/rf-pics; rm -rf $o; mkdir -p $o
S='900x700 1280x620 360x640 412x780'
for th in dark light; do
 tools/uishots.sh --char --out $o --url "/watch?first=0&theme=$th&needs=0" --sizes "$S" --tag idle-$th --theme $th
 tools/uishots.sh --char --out $o --url "/watch?first=0&theme=$th&needs=0&reset=resetting" --click '#btn-reset' --sizes "$S" --tag pressed-$th --theme $th
done
ls $o
