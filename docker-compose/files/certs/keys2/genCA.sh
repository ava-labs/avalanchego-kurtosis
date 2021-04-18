# (c) 2021, Ava Labs, Inc. All rights reserved.
# See the file LICENSE for licensing terms.

!/bin/sh
set -ex

openssl genrsa -out `dirname "$0"`/rootCA.key 4096
openssl req -x509 -new -nodes -key `dirname "$0"`/rootCA.key -sha256 -days 365250 -out `dirname "$0"`/rootCA.crt
