#!/bin/bash

# Ensures the local run is using the correct version of libraries
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source ${__dir}/dep-versions.sh "${@}"

go mod edit -require github.com/ava-labs/avalanchego@"${avalancheGoVersion}"
go mod edit -require github.com/ava-labs/coreth@"${corethVersion}"
go mod edit -require github.com/ethereum/go-ethereum@"${goEthereum}"

source ${__dir}/build-and-run.sh "${@}"

