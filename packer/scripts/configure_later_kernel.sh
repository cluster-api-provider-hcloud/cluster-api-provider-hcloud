#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

cat > /etc/yum.repos.d/simonswine-kernel-lts.repo <<EOF
[copr:copr.fedorainfracloud.org:simonswine:kernel-lts-5.4]
name=Copr repo for kernel-lts-5.4 owned by simonswine
baseurl=https://raw.githubusercontent.com/simonswine/centos-kernel-lts-yum/main/kernel-lts-5.4-epel-\$releasever/
type=rpm-md
skip_if_unavailable=True
gpgcheck=1
gpgkey=https://download.copr.fedorainfracloud.org/results/simonswine/kernel-lts-5.4/pubkey.gpg
repo_gpgcheck=0
enabled=1
enabled_metadata=1
EOF

# update kernel
yum -y install kernel-5.4.43-300.el7
