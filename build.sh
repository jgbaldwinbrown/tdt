Jim@T480:im$ cat build.sh 
#!/bin/bash
set -e

mkdir -p bin

GOOS=windows GOARCH=amd64 go build -o bin/client-windows-amd64.exe client.go
GOOS=windows GOARCH=386 go build -o bin/client-windows-386.exe client.go

GOOS=darwin GOARCH=amd64 go build -o bin/client-darwin-amd64 client.go
# GOOS=darwin GOARCH=386 go build -o bin/client-darwin-386 client.go

GOOS=linux GOARCH=amd64 go build -o bin/client-linux-amd64 client.go
GOOS=linux GOARCH=386 go build -o bin/client-linux-386 client.go

GOOS=windows GOARCH=amd64 go build -o bin/server-windows-amd64.exe server.go
GOOS=windows GOARCH=386 go build -o bin/server-windows-386.exe server.go

GOOS=darwin GOARCH=amd64 go build -o bin/server-darwin-amd64 server.go
# GOOS=darwin GOARCH=386 go build -o bin/server-darwin-386 server.go

GOOS=linux GOARCH=amd64 go build -o bin/server-linux-amd64 server.go
GOOS=linux GOARCH=386 go build -o bin/server-linux-386 server.go
