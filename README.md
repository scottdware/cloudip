## cloudip

This program will retrieve a list of all public IP address ranges (v4 or v6) for the three major cloud vendors:

Amazon AWS, Microsoft Azure and Google

By default, the ranges are printed to the console/screen. If you would like to save them in a file, the
output format is CSV, and you can use the "--file" flag to specify a file name.

```
Usage:
  cloudip [flags]

Flags:
  -f, --file string     CSV filename to save the output to
  -h, --help            help for cloudip
  -i, --iptype int      IP Type to export - 4|6|all (default 4)
  -v, --vendor string   Cloud vendor to export IP's from - aws|azure|google
```

### Installation

Clone this repo and run the script as follows (must have Golang installed):

`go run main.go --vendor aws --iptype 4 --file AWS_IP_Ranges.csv`
