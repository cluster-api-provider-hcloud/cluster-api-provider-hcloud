#!/bin/sh
PACKER=$1
shift
export HCLOUD_TOKEN=test
find .
exec $PACKER validate "$@"
