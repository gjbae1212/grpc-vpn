#!/bin/bash

set -e -o pipefail

trap '[ "$?" -eq 0 ] || echo "Error Line:<$LINENO> Error Function:<${FUNCNAME}>"' EXIT

cd `dirname $0`
CURRENT=`pwd`

function gen_grpc
{
   rm -rf $CURRENT/../grpc/go/*
   protoc --go_out=plugins=grpc:$CURRENT/../grpc/go vpn.proto vpn-struct.proto
}

function build_vpn_server
{
   cd $CURRENT/../cmd/vpn-server
   GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o $CURRENT/../dist/vpn-server-linux
}

function build_vpn_client
{
   cd $CURRENT/../cmd/vpn-client
   GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o $CURRENT/../dist/vpn-client-darwin
   GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o $CURRENT/../dist/vpn-client-linux
}

CMD=$1
shift
$CMD $*
