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

## tdtall

Tdtall runs the TDT test on a pedigree, either reporting just one test result
for a focal individual or reporting one result for each individual in the
pedigree, taking them as the focal individual one by one.

```
Usage of tdtall:
  -f string
    	IndividualID for focal individual (default is to do TDT for all males)
  -i string
    	path to input .ped file
  -o string
    	path to write output
```

Once tdtall is installed, all you need to do is run the following in the directory that contains your pedigree file:

```sh
tdtall -i myped.ped -o myout.json
```

Here, `-i` specifies the input file and `-o` specifies the output file. If the input or output files end in ".gz", the files will be handled as gzipped files.

## tdtmonte

Tdtmonte runs a monte carlo simulation of the TDT test by randomly generating
families of the same size as in the background results (-b), then seeing if the
true family (-a) is more significant than the simulated families

```
Usage of tdtmonte:
  -a string
    	path to .json containing actual family results
  -b string
    	path to .json containing background families
  -r int
    	Replicates (default 1)
  -s int
    	Random seed
```

## pedshufsex

Pedshufsex shuffles either the sex or the phenotype of all individuals in a
pedigree the number of times specified, putting each shuffled pedigree in a separate .ped file.

```
Usage of pedshufsex:
  -i string
    	input .ped path (default stdin)
  -o string
    	output prefix (default "shuf_ped_sex_out")
  -p	huffle phenotype instead of sex
  -r int
    	shuffle replicates (default 1)
  -s int
    	random seed
```

## outlier

Outlier takes the output of a set of background pedigrees (usually shuffled
with pedshufsex) and the output of a real pedigree of interest, then provides
statistics on whether the real pedigree contains individuals more likely to
carry a distorter than those in background pedigrees.

```
Usage of outlier:
  -b string
    	Path to list of paths containing warp output for background data
  -bh
    	Background data has a header line
  -c string
    	Chosen individual ID to run rank order statistics on
  -r string
    	Path to output of warp for real data
  -rh
    	Real data has a header line
  -t int
    	Top number of individuals to average to get score (default 1) (default -1)
```
