# (c) 2021, Ava Labs, Inc. All rights reserved.
# See the file LICENSE for licensing terms.

!/bin/bash
# TODO wait for node to be up instead of using sleep
set -e

host="$1"

generate_post_data()
{
  cat <<EOF
{
    "jsonrpc":"2.0",
    "id"     :1,
    "method" :"info.isBootstrapped",
    "params": {
        "chain": "X"
    }
}
EOF
}

run_check(){
    check="$(curl -s -H "Accept: application/json" -H "Content-Type:application/json" -X POST --data-raw "$(generate_post_data)" "https://api.avax.network/ext/info")"

    echo $check
    if [ -z "${check##*"true"*}" ];
    then
        return 0
    fi
    return 1
}


shift

until run_check; do
  >&2 echo "Node ${host} is not bootstrapped - sleeping"
  sleep 1
done

node1=$(getent hosts node-1 | awk '{ print $1 }')
echo $node1
echo "$($@ --bootstrap-ips=${node1}:9651 )"
