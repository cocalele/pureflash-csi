#!/bin/bash
set -v

go mod tidy
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ./bin/pureflash_csi .
strip ./bin/pureflash_csi



docker build -t pureflash/pureflash-csi:1.8.3 .
docker tag pureflash/pureflash-csi:1.8.3 docker.io/pureflash/pureflash-csi:1.8.3


