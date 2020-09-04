#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# Ensure we don't leave SSH host keys
rm -rf /etc/ssh/ssh_host_*
