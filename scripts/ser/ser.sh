#!/bin/bash
set -e

# Calculate the spelling-error-rate (SER) of a file
# Replaces long s (ſ) with short s (s) and removes all non ascii chars
# Removes punctuation and any words that only have digits as chars
# Writes the SER to stdout and the misspelled word list to stderr

if [ $# -ne 1 ] && [ $# -ne 2 ] && [ -t 0 ] ; then
  echo "usage: $0 file.txt [stoplist.txt]"
  exit 1
fi

tmp1=`mktemp`
tmp2=`mktemp`

# Strip UTF-8 (and ſ with s); punctuation; one word per line; remove words that are only digits
cat $1 |
unutf_ſ |
tr '[:punct:]' ' ' |
tr -s '[:space:]' '\n' |
grep -v '^[0-9]*$' |
sort -u > $tmp1

# Send words that are not in the stoplist to v10spell
comm -23 $tmp1 ${2-/dev/null} | v10spell > $tmp2

# Write misspelled words to stderr
>&2 cat $tmp2

# Write SER to stdout
echo "`wc -w < $tmp2`,`wc -w < $tmp1`" | awk -F, '{print ($2==0?($1==0?0:1):$1/$2)}'

# Cleanup
rm -f $tmp1 ; rm -f $tmp2
