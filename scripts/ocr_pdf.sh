#!/bin/bash
set -e

DPI=400
PLT_MAX_TIME=12
PLT_MAX_DIST=1500

if [ $# -ne 1 ]; then
    echo "usage: ./ocr.sh file.pdf"
    exit 1
fi

echo "  OP | Seconds "
echo " --- + ------- "
echo -n " JPG | "
NAME=`basename $1 .pdf`
IMAGES=${NAME}_jpg${DPI}
mkdir ${IMAGES}
START=`date +'%s'`
gs -dSAFER -dQUIET -dBATCH -dNOPAUSE -sDEVICE=jpeg -r$DPI -o $IMAGES/$NAME-%03d.jpg $1
STOP=`date +'%s'`
echo "$(( STOP - START ))"

echo -n " OCR | "
JSON=${NAME}_ocr${DPI}
mkdir $JSON && cd $JSON
START=`date +'%s'`
for f in ../$IMAGES/*; do tigerocr -aws -azure -gcp $f >>stdout.txt 2>> stderr.txt; done
STOP=`date +'%s'`
echo "$(( STOP - START ))"
cd ..

echo -n " LEV | "
START=`date +'%s'`
combine.sh $JSON
STOP=`date +'%s'`
echo "$(( STOP - START ))"

echo -n " TXT | "
TXT=${NAME}_txt${DPI}
START=`date +'%s'`
extractall.sh $JSON $TXT
STOP=`date +'%s'`
echo "$(( STOP - START ))"

echo -n " PLT | "
START=`date +'%s'`
plot_lev.py combined.csv lev.png
plot_lev_vs_time.py combined.csv lev_vs_time.png $PLT_MAX_TIME $PLT_MAX_DIST
STOP=`date +'%s'`
echo "$(( STOP - START ))"
echo " --- + ------- "
echo "Done."
