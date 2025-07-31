#!/bin/bash
project="kohme"

set -e

go mod tidy

go run ./cmd/plugin

go build -ldflags "-s -w" -o $project ./cmd/bot


# docker build -t $project .
# rm -rf $project
# echo "build success $project:$version"