## cloudip

This program will retrieve a list of all public IP address ranges (v4 or v6) for the three major cloud vendors:

Amazon AWS (aws), Microsoft Azure (azure) and Google (google)

Flags:

For the --vendor (-v) flag, you must specify one of these options: aws | azure | google
For the --iptype (-i) flag, you must specify either: 4 or 6

By default, the ranges are printed to the console/screen. If you would like to save them in a file, the
output format is CSV, and you can use the "--file" (-f) flag to specify a file name.

```
Usage:
  cloudip [flags]

Flags:
  -f, --file string     CSV filename to save the output to
  -h, --help            help for cloudip
  -i, --iptype int      IP Type to export - 4|6 (default 4)
  -v, --vendor string   Cloud vendor to export IP's from - aws|azure|google
```

### Installation

The easiest way to run this program is to download the binary for your OS of choice from the [Releases](https://github.com/scottdware/cloudip/releases/latest) section.

You can optionally choose to clone this repo and run the script as follows (must have Golang installed):

`go run main.go --vendor aws --iptype 4 --file AWS_IP_Ranges.csv`
