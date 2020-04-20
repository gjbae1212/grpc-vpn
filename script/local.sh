#!/bin/bash

set -e -o pipefail

trap '[ "$?" -eq 0 ] || echo "Error Line:<$LINENO> Error Function:<${FUNCNAME}>"' EXIT

cd `dirname $0` && cd ..
CURRENT=`pwd`


function test
{
   go test -v $(go list ./... | grep -v vendor) --count 1 -race -covermode=atomic -timeout 120s
}


CMD=$1
shift
$CMD $*
