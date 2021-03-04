#!/bin/bash
# Generate a results.csv given a json or blw OCR results directory
# There should only be files of one extension type (not multiple) in dir

OUT="results.csv"

tmpOUT=`mktemp`
tmpAWS=`mktemp`
tmpAzu=`mktemp`
tmpGCP=`mktemp`
tmpCOMM=`mktemp`

if [ $# -ne 1 ]; then
    echo "usage: tigerocr-combine.sh json-dir/"
    exit 1
fi

# Create list of all pointers which passed all 3 services
ls $1 | grep aws | cut -d. -f1 | sort | uniq > $tmpAWS
ls $1 | grep azu | cut -d. -f1 | sort | uniq > $tmpAzu
ls $1 | grep gcp | cut -d. -f1 | sort | uniq > $tmpGCP
tmp=`mktemp`
comm -12 $tmpAWS $tmpAzu > $tmp
comm -12 $tmp $tmpGCP > $tmpCOMM
rm -f $tmpAWS $tmpAzu $tmpGCP $tmp

# Create $OUT (with all levenshtein distances and run times)
echo "ptr,AWS-Azu Lev,Azu-GCP Lev,GCP-AWS Lev,AWS Millis,Azu Millis,GCP Millis" > $tmpOUT
for p in `cat $tmpCOMM`; do
    rm -f $tmpAWS $tmpAzu $tmpGCP
    unset AWS_MILLIS AZU_MILLIS GCP_MILLIS AWS_AZU_LEV AZU_GCP_LEV GCP_AWS_LEV
    tigerocr extract -text $1/${p}*aws*> $tmpAWS
    tigerocr extract -text $1/${p}*azu* > $tmpAzu
    tigerocr extract -text $1/${p}*gcp* > $tmpGCP
    AWS_MILLIS=`tigerocr extract -speed $1/${p}*aws*`
    AZU_MILLIS=`tigerocr extract -speed $1/${p}*azu*`
    GCP_MILLIS=`tigerocr extract -speed $1/${p}*gcp*`
    AWS_AZU_LEV=`tigerocr editdist $tmpAWS $tmpAzu`
    AZU_GCP_LEV=`tigerocr editdist $tmpAzu $tmpGCP`
    GCP_AWS_LEV=`tigerocr editdist $tmpGCP $tmpAWS`
    echo "$p,$AWS_AZU_LEV,$AZU_GCP_LEV,$GCP_AWS_LEV,$AWS_MILLIS,$AZU_MILLIS,$GCP_MILLIS" >> $tmpOUT
done
ruby -rcsv -e 'puts CSV.parse(STDIN).transpose.map &:to_csv' < $tmpOUT > $OUT

rm -f $tmpOUT $tmpAWS $tmpAzu $tmpGCP $tmpCOMM
