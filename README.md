# TigerOCR

`tigerocr` is a client used for benchmarking OCR performance of leading cloud providers.

## Install

```
go install ./...
```

## Create Keys

Follow the documentation below to create keys for `AWS`, `Azure`, and `GCP`

```
$ tigerocr run --help
usage: tigerocr run [-keys=~/keydir/] [-aws] [-azure] [-gcp] image.jpg

  -aws
    	Run AWS Textract OCR. Key files: credentials config
    	More info: https://docs.aws.amazon.com/textract/latest/dg/setup-awscli-sdk.html
  -azure
    	Run Azure CognitiveServices OCR. Key file: azure.json
    	More info: https://docs.microsoft.com/azure/cognitive-services/cognitive-services-apis-create-account
    	Note: Create a json file with 'subscription_key' and 'endpoint' items
  -azureR
        Run Azure CognitiveServices Read API.
  -gcp
    	Run GCP Vision OCR. Key file: gcp.json
    	More info: https://cloud.google.com/vision/docs/before-you-begin
  -keys string
    	Path to credentials directory (default "~/.aws")
```

## Available Commands

```
$ tigerocr --help
usage: tigerocr <command> [arguments]

The commands are:

	run     	 execute ocr on selected providers
	annotate	 draw bounding boxes of words on the original image
	editdist	 calculate levenshtein distance of two text files
	convert 	 convert json ocr responses to unified blw format (*)
	extract 	 extract metadata from a blw or json datafile
	explore 	 execute pdf ocr and output results as a web explorer
	serve   	 serve current directory at 127.0.0.1:8080
```

## Example

```
$ tigerocr explore -keys ~/.aws -aws -azure -azureR -gcp book.pdf
[INFO] PDF to PNG (Total: 20) ... 		32 secs
[INFO] PDF to TXT (Total: 20) ... 		9 secs
[ATTN] Estimate: $0.12 (80 ops). Run? [N/y]: 	y
[INFO] Executing OCR (Total: 80) ... 		56 secs
[INFO] JSON to BLW (Total: 80) ... 		5 secs
[INFO] BLW to TXT ... 				0 secs
[INFO] Running Levenshtein Distance ... 	1 secs
[INFO] Creating Explorer ... 			done
[INFO] Comparable Ptrs: 20 (out of 20). aws: 20, azu: 20, azuR: 20, gcp: 20
[DONE] Run: tigerocr serve ./explorer-book
$ tigerocr serve ./explorer-book
```
