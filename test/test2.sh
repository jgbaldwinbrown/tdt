#!/bin/sh
set -e

go build ../cmd/inbreeding_coefficients.go
./inbreeding_coefficients <inbred.ped
