#!/bin/bash
# Download United States Report pdfs given the report volume number
set -e

MAX_COUNT=1000
URL_SEARCH='https://www.loc.gov/search/?fo=json&c='${MAX_COUNT}'&fa=partof:u.s.+reports:+volume+%s\n'
URL_MIRROR='https://tile.loc.gov/storage-services/service/ll/usrep/%s/%s/%s.pdf\n'
TIMEOUT=30

function ptr_to_url {
  printf "${URL_MIRROR}" "${1:0:8}" $1 "$1"
}

# i.e. usrep001005a => 001us005a
function convert_ptr {
        set -e
        echo "$1" | grep -q 'usrep[0-9][0-9][0-9]' || (>&2 echo "ERROR: Invalid Format" && exit 1)
        echo "$1" | sed 's/^usrep//;s/^.../&us/'
}

if [ $# -ne 1 ]; then
  echo "usage: $0 241"
  exit 1
fi

if ! ([ "$@" -ge 1 ] && [ "$@" -le 542 ]) 2>/dev/null; then
  echo "usage: $0 x"
  echo "NOTE: 1 <= x <= 542"
  exit 1
fi

DST_DIR=$(printf "%03dus" $@)
mkdir $DST_DIR
cd $DST_DIR
TMP=$(mktemp)
curl -sfL "$(printf $URL_SEARCH $@)" > $TMP || (echo "[ERROR] Download Failed | USR $@" && exit 1)
grep -q '"next": null,' $TMP || (echo "[ERROR] Expected fewer files (< $MAX_COUNT) | USR $@" && exit 1)
TOTAL=$(grep -o '"shelf_id": "[A-Za-z0-9]*",' $TMP | wc -l)
grep -o '"shelf_id": "[A-Za-z0-9]*",' $TMP | awk -F\" '{print $4}' | while read PTR; do
  URL="$(ptr_to_url $PTR)"
  DPTR="$(convert_ptr $PTR)"
  # echo "$DPTR, $URL" # TODO: Uncomment this line to see all downloads
  curl -sfL -o "${DPTR}.pdf" --max-time $TIMEOUT "${URL}" || echo "${DPTR}, ${URL}" >&2
done
rm $TMP
DOWNLOADED=$(ls | wc -l)
if [ $DOWNLOADED -eq $TOTAL ]; then
  echo "[DONE] Total: $TOTAL. U.S. Report $@"
else
  echo "[DONE] Errors: $(($TOTAL - $N)). Total: $TOTAL. U.S. Report $@"
fi
