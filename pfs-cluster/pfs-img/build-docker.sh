#!/bin/bash
function fatal {
    echo -e "\033[31m$* \033[0m"
    exit 1
}
function assert()
{
    local cmd=$*
	echo "Run:$cmd" > /dev/stderr
	eval '${cmd}'
    if [ $? -ne 0 ]; then
        fatal "Failed to run:$cmd"
    fi
}

docker build -f Dockerfile -t pureflash/pureflash-k8s:1.8.2 .
