#!/usr/bin/env bash
set -euo pipefail

docker build -f benzhi.Dockerfile -t windctl:benzhi .
docker run --rm -d -p 8080:8080 --name windctl-benzhi windctl:benzhi
