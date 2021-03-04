#!/bin/bash

if [ $# -ne 2 ]; then
    echo "usage: tigerocr-extractall.sh json-dir/ output-txts/"
    exit 1
fi

mkdir -p $2 $2/aws $2/azu $2/gcp
ls $1 | grep aws | while read f; do
    tigerocr extract -text $1/${f} > $2/aws/${f%.*.*}.txt
done
ls $1 | grep azu | while read f; do
    tigerocr extract -text $1/${f} > $2/azu/${f%.*.*}.txt
done
ls $1 | grep gcp | while read f; do
    tigerocr extract -text $1/${f} > $2/gcp/${f%.*.*}.txt
done
