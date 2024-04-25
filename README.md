# Pedviz: a program for testing and visualizing 

# Installation

## Simple method: binary release

1. Download the latest release tarball
2. Copy the binary for your architecture to a new file named "tdtall"
3. Copy "tdtall" into a directory that is in your PATH

## Advanced method: build from source

0. Make sure you have a version of the Go compiler and build system installed. Minimum version 0.18.
1. Clone this repository
2. In the main directory, run `go mod tidy`
3. In the main directory, run `go build cmd/tdtall.go`
4. Copy "tdtall" into a directory that is in your PATH

# Usage

Once tdtall is installed, all you need to do is run the following in the directory that contains your pedigree file:

```sh
tdtall -i myped.ped -o myout.json
```

Here, `-i` specifies the input file and `-o` specifies the output file. If the input or output files end in ".gz", the files will be handled as gzipped files.
