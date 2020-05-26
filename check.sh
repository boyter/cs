#!/usr/bin/env bash

GREEN='\033[1;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "Running go fmt..."
gofmt -s -w ./..

echo "Running unit tests..."
go test -cover -race ./... || exit

# Race Detection
echo "Running race detection..."
if go run --race . t NOT something test~1 "ten thousand a year" "/pr[a-z]de/" 2>&1 >/dev/null | grep -q "Found" ; then
    echo -e "${RED}======================================================="
    echo -e "FAILED race detection run 'go run --race . test' to identify"
    echo -e "=======================================================${NC}"
    exit
else
    echo -e "${GREEN}PASSED race detection${NC}"
fi

echo "Building application..."
go build -ldflags="-s -w" || exit

echo -e "${NC}Checking compile targets..."

echo "   darwin..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w"
GOOS=darwin GOARCH=386 go build -ldflags="-s -w"
echo "   windows..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w"
GOOS=windows GOARCH=386 go build -ldflags="-s -w"
echo "   linux..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"
GOOS=linux GOARCH=386 go build -ldflags="-s -w"

echo -e "${NC}Cleaning up..."
rm ./cs
rm ./cs.exe

echo -e "${GREEN}================================================="
echo -e "ALL CHECKS PASSED"
echo -e "=================================================${NC}"
