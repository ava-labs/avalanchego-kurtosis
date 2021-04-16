#!/bin/sh

# Set up the versions to be used
# Don't export them as their used in the context of other calls
avalancheGoVersion=${AVALANCHEGO_VERSION:-'v1.3.2'}
corethVersion=${CORETH_VERSION:-'v0.4.0-rc.8'}
goEthereum=${GO_ETHEREUM:-'v1.9.21'}