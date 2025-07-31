#!/bin/bash
project="kohme"

set -e

bash build.sh
docker build -t $project .
rm -f $project
echo "build success $project:$version"