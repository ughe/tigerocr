#!/bin/bash
# Download United States Report pdfs given the report id
set -e

MAX_COUNT=1000
URL_SEARCH='https://www.loc.gov/search/?fo=json&c='${MAX_COUNT}'&fa=partof:u.s.+reports:+volume+%s\n'
URL_MIRROR='https://tile.loc.gov/storage-services/service/ll/usrep/%s/%s/%s.pdf\n'

function ptr_to_url {
  printf "${URL_MIRROR}" "${1:0:8}" $1 "$1"
}

if [ $# -ne 1 ]; then
  echo "usage: getptr2.sh 241"
  exit 1
fi

if ! ([ "$@" -ge 1 ] && [ "$@" -le 542 ]) 2>/dev/null; then
  echo "usage: getptr2.sh x"
  echo "NOTE: 1 <= x <= 542"
  exit 1
fi

DST_DIR=$(printf "usrep%03d" $@)
mkdir $DST_DIR
cd $DST_DIR
TMP=$(mktemp)
curl -sfL "$(printf $URL_SEARCH $@)" > $TMP || (echo "[ERROR] Download Failed | USR $@" && exit 1)
grep -q '"next": null,' $TMP || (echo "[ERROR] Expected fewer files (< $MAX_COUNT) | USR $@" && exit 1)
TOTAL=$(grep -o '"shelf_id": "[A-Za-z0-9]*",' $TMP | wc -l)
grep -o '"shelf_id": "[A-Za-z0-9]*",' $TMP | awk -F\" '{print $4}' | while read PTR; do
  URL="$(ptr_to_url $PTR)"
  # echo "$PTR, $URL" # TODO: Uncomment this line to see all downloads
  curl -sfOL --max-time 30 "${URL}" || echo "${PTR}, ${URL}" >&2
done
rm $TMP
DOWNLOADED=$(ls | wc -l)
if [ $DOWNLOADED -eq $TOTAL ]; then
  echo "[DONE] Total: $TOTAL. U.S. Report $@"
else
  echo "[DONE] Errors: $(($TOTAL - $N)). Total: $TOTAL. U.S. Report $@"
fi
