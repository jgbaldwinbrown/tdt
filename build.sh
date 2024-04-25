#!/bin/bash
set -e

mkdir -p bin

GOOS=windows GOARCH=amd64 go build -o bin/tdtall-windows-amd64.exe cmd/tdtall.go
GOOS=windows GOARCH=386 go build -o bin/tdtall-windows-386.exe cmd/tdtall.go

GOOS=darwin GOARCH=amd64 go build -o bin/tdtall-darwin-amd64 cmd/tdtall.go
# GOOS=darwin GOARCH=386 go build -o bin/tdtall-darwin-386 cmd/tdtall.go

GOOS=linux GOARCH=amd64 go build -o bin/tdtall-linux-amd64 cmd/tdtall.go
GOOS=linux GOARCH=386 go build -o bin/tdtall-linux-386 cmd/tdtall.go
