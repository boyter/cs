#!/usr/bin/env bash

set -e

if [ -t 1 ]
then
  YELLOW='\033[0;33m'
  GREEN='\033[0;32m'
  RED='\033[0;31m'
  NC='\033[0m'
fi

yellow() { printf "${YELLOW}%s${NC}" "$*"; }
green() { printf "${GREEN}%s${NC}" "$*"; }
red() { printf "${RED}%s${NC}" "$*"; }

good() {
  echo "$(green "● success:")" "$@"
}

bad() {
  ret=$1
  shift
  echo "$(red "● failed:")" "$@"
  exit $ret
}

try() {
  "$@" || bad $? "$@" && good "$@"
}


echo "Running go fmt..."
gofmt -s -w ./..

echo "Running unit tests..."
go test -cover -race ./... || exit

{
  {
    opt='shopt -s extglob nullglob'
    gofmt='gofmt -s -w -l !(vendor)/ *.go'
    notice="    running: ( $opt; $gofmt; )"
    prefix="    $(yellow modified)"
    trap 'echo "$notice"; $opt; $gofmt | sed -e "s#^#$prefix #g"' EXIT
  }

  # comma separate linters (e.g. "gofmt,stylecheck")
  additional_linters="gofmt"
  try golangci-lint run --enable $additional_linters ./...
  trap '' EXIT
}

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
echo "   windows..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w"
echo "   linux..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"
GOOS=linux GOARCH=386 go build -ldflags="-s -w"

echo -e "${NC}Cleaning up..."
rm ./cs
rm ./cs.exe

echo -e "${GREEN}================================================="
echo -e "ALL CHECKS PASSED"
echo -e "=================================================${NC}"
