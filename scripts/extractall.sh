#!/bin/bash

if [ $# -ne 2 ]; then
    echo "usage: ./extractall.sh ocr_json_dir out_dir"
    exit 1
fi

mkdir -p $2
for p in `cat ocr.txt`; do
    tigerocr extract -text $1/${p}.jpg.aws.json > $2/${p}.aws.txt
    tigerocr extract -text $1/${p}.jpg.azure.json > $2/${p}.azu.txt
    tigerocr extract -text $1/${p}.jpg.gcp.json > $2/${p}.gcp.txt
done
