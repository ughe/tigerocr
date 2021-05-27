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
