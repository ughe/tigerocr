#!/bin/bash

if [ $# -ne 1 ]; then
    echo "usage: ./combine.sh ocr_json_dir"
    exit 1
fi

# Create ocr.txt (list of all pointers which passed all 3 services)
ls $1 | grep aws | cut -d. -f1 | sort | uniq > .aws.tmp
ls $1 | grep azu | cut -d. -f1 | sort | uniq > .azu.tmp
ls $1 | grep gcp | cut -d. -f1 | sort | uniq > .gcp.tmp
comm -12 .aws.tmp .azu.tmp > .awsazu.tmp 
comm -12 .awsazu.tmp .gcp.tmp > ocr.txt
rm -f .aws.tmp .azu.tmp .gcp.tmp .awsazu.tmp 

# Create combined.csv (with all levenshtein distances and run times)
for p in `cat ocr.txt`; do
    rm -f .aws.tmp .azu.tmp .gcp.tmp
    unset AWS_MILLIS AZU_MILLIS GCP_MILLIS AWS_AZU_LEV AZU_GCP_LEV GCP_AWS_LEV
    extractor -text $1/${p}.jpg.aws.json > .aws.tmp
    extractor -text $1/${p}.jpg.azure.json > .azu.tmp
    extractor -text $1/${p}.jpg.gcp.json > .gcp.tmp
    AWS_MILLIS=`extractor -milliseconds $1/${p}.jpg.aws.json`
    AZU_MILLIS=`extractor -milliseconds $1/${p}.jpg.azure.json`
    GCP_MILLIS=`extractor -milliseconds $1/${p}.jpg.gcp.json`
    AWS_AZU_LEV=`editdist .aws.tmp .azu.tmp`
    AZU_GCP_LEV=`editdist .azu.tmp .gcp.tmp`
    GCP_AWS_LEV=`editdist .gcp.tmp .aws.tmp`
    echo "$p,$AWS_AZU_LEV,$AWS_MILLIS,$AZU_GCP_LEV,$AZU_MILLIS,$GCP_AWS_LEV,$GCP_MILLIS" >> combined.csv
done
rm -f .aws.tmp .azu.tmp .gcp.tmp
