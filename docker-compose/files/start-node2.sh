#!/bin/bash
# (c) 2021, Ava Labs, Inc. All rights reserved.
# See the file LICENSE for licensing terms.

set -e
sleep 5
node1=$(getent hosts node-1 | awk '{ print $1 }')
node2=$(getent hosts node-2 | awk '{ print $1 }')
/avalanchego/build/avalanchego --http-host=${node2} --public-ip=${node2} --network-id=local --staking-enabled=true --http-port=9650 --staking-port=9651 --log-level=debug --bootstrap-ids=NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg --bootstrap-ips=${node1}:9651 --staking-tls-cert-file=/files/certs/keys2/staker.crt --staking-tls-key-file=/files/certs/keys2/staker.key
