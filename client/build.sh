#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

mkdir -p dist

GOOS=linux   GOARCH=arm64 go build -o dist/client.linux-arm64.out .
GOOS=linux   GOARCH=amd64 go build -o dist/client.linux-amd64.out .
GOOS=windows GOARCH=amd64 go build -o dist/client.windows-amd64.exe .

rm -rf dist/web
cp -r web dist/
