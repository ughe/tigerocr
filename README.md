# TigerOCR

`tigerocr` is a client used for benchmarking OCR performance of leading cloud providers.

## Install

```
pip install -r requirements.txt
go install ./...
cp scripts/* $GOPATH/bin/
```

## Create Keys

Follow the documentation below to create keys for `AWS`, `Azure`, and `GCP`

```
$ tigerocr --help
usage: tigerocr <command> [arguments]

The commands are:

	run     	 execute ocr on selected providers
	annotate	 draw bounding boxes of words on the original image
	editdist	 calculate levenshtein distance of two plaintext files
	convert 	 convert json ocr responses to unified blw files
	extract 	 extract metadata from a blw or json datafile
```

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
  -gcp
    	Run GCP Vision OCR. Key file: gcp.json
    	More info: https://cloud.google.com/vision/docs/before-you-begin
  -keys string
    	Path to credentials directory (default "~/.aws")
```

## OCR PDF

```
ocr_pdf.sh file.pdf
```
