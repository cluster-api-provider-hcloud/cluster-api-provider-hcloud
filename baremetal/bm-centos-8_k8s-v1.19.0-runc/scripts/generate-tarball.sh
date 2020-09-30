#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

rm -f /installimage.conf /installimage.debug

dnf -y install tar
tar zcvf /CentOS-82-64-minimal.tar.gz --exclude=/dev --exclude=/proc --exclude=/sys --exclude=/tmp/scripts /
