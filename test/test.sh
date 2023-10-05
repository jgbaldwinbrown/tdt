#!/bin/bash
set -e

go build
cat ex.ped | ./tdt -f 1
