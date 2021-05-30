#!/bin/bash
set -e

if [ $# -lt 2 ]; then
	echo "usage: $0 fst snd ..."
	exit 1
fi

tmp0=`mktemp`
tmp1=`mktemp`

join -t, $1 $2 > $tmp0
mv $tmp0 $tmp1
for f in "${@:3}"; do join -t, $tmp1 $f > $tmp0 && mv $tmp0 $tmp1 ; done
cat $tmp1
rm $tmp1
