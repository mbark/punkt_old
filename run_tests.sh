#!/usr/bin/env bash
set -o pipefail
set -e

docker build -t dotool:test .
docker run dotool:test
