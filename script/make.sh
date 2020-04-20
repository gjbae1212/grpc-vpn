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

function build
{
   echo "build"
}

CMD=$1
shift
$CMD $*
