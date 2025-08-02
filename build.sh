#!/bin/bash
project="kohme"

set -e

go mod tidy

go generate

CGO_ENABLED=1 go build -ldflags "-s -w" -o $project ./cmd/bot


# docker build -t $project .
# rm -rf $project
# echo "build success $project:$version"