#!/bin/bash

echo "Running go fmt..."
gofmt -s -w ./..

echo "Running unit tests..."
go test ./... || exit

echo "Building application..."
go build -ldflags="-s -w" || exit

echo "Running integration tests..."

GREEN='\033[1;32m'
RED='\033[0;31m'
NC='\033[0m'

if ./cs --not-a-real-option > /dev/null ; then
    echo -e "${RED}================================================="
    echo -e "FAILED Invalid option should produce error code "
    echo -e "=======================================================${NC}"
    exit
else
    echo -e "${GREEN}PASSED invalid option test"
fi

echo -e "${NC}Cleaning up..."
rm ./cs

echo -e "${GREEN}================================================="
echo -e "ALL TESTS PASSED"
echo -e "=================================================${NC}"
