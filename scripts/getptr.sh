#!/bin/bash

if [ $# -ne 1 ]; then
  echo "usage: ./getptr.sh OA167605170001"
  exit 1
fi

MIRROR0="https://www.dhi.ac.uk/san"
#MIRROR1="https://www.hrionline.ac.uk/san"
#MIRRORN="MIRROR$(( $RANDOM % 2 ))"
#MIRROR="${!MIRRORN}"
MIRROR="${MIRROR0}"
for PTR in "$@"; do
  if [ "${PTR:0:2}" = "OA" ]; then
    if [[ "${PTR:10:1}" =~ [0-9] ]]; then
      URL="${MIRROR}/oa/${PTR:0:10}/${PTR}.jpg"
    else
      URL="${MIRROR}/oa/${PTR:0:-4}/${PTR}.jpg"
    fi
  elif [ "${PTR:0:4}" -lt 1834 ] 2>/dev/null; then
    URL="${MIRROR}/ob/${PTR:0:3}0s/${PTR}.gif"
  else
    URL="${MIRROR}/ccc/${PTR:0:8}/${PTR}.jpg"
  fi
  echo "${PTR}, ${URL}"
  ## Download. Note: should check error list and re-try
  # curl -sOJL --fail --max-time 7 "${URL}" || echo "${PTR}, ${URL}" >> ${ERRORS}
done
