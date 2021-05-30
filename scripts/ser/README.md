## Spelling Error Rate (SER)

The SER is the number of unique spelling misstakes divided by the total number of unique words in a file. A stoplist may be provided to ignore known words during the spell check.

## v10spell

The v10spell is from: https://github.com/arnoldrobbins/v10spell
Commit hash: `39b16d4866ad806dbeaccd7d55677a5614e4bdab`

Before sending text to v10spell, all punctuation is removed as well as any words that are completely digits. All UTF-8 characters are removed. The long s (ſ) is replaced with the short s (s).

## Usage:

Prints the error rate on standard out:

```
ser.sh file.txt stoplist.txt 2> misspelled.txt
```

Install `ser.sh` to the `$GOPATH/bin` or `/usr/local/bin` using `make install`.

## Long Example

```
P=(AWS AzureOCR GCP AzureRead macOS Tesseract)
SRC=~/obo.cs.princeton.edu/data2021/oldbailey/txts/
for s in $P; do echo $s && ls ${SRC}/$s > ls_${s}.txt ; time cat ls_${s}.txt | while read f ; do ser.sh $SRC/$s/$f stoplist_ob21.txt 2>> spell_mistakes_${s}.txt >> spell_ser_${s}.txt ; done ; done

for s in $P; do echo $s ; paste -d, ls_${s}.txt spell_ser_${s}.txt > _${s} ; done

printf "ptr" ; for d in $P ; do printf ",$d SER" ; done ; printf "\n" > _ser.txt
joins.sh `for s in $P; do printf "_$s " ; done` | sed 's/\.txt//' | awk -F, '{printf "%s,",$1 ; for (i=2;i<NF;i++) { if(length($i) > 7){printf "%.5f,", $i}else{printf "%s,", $i} } ; if(length($NF) > 7) {printf "%.5f\n", $NF}else{printf "%s\n", $NF}}' >> _ser.txt
```
