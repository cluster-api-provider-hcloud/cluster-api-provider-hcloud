#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# install cri-o
cat << 'EOF' > /etc/yum.repos.d/crio.repo
[copr:copr.fedorainfracloud.org:simonswine:cri-o]
name=CRI-O Packages for EL 7
baseurl=https://copr-be.cloud.fedoraproject.org/results/simonswine/cri-o/epel-7-$basearch/
type=rpm-md
skip_if_unavailable=True
gpgcheck=1
gpgkey=https://copr-be.cloud.fedoraproject.org/results/simonswine/cri-o/pubkey.gpg
repo_gpgcheck=0
enabled=1
enabled_metadata=1
EOF

yum -y install cri-o cri-tools

# remove default CNIs
rm -f /etc/cni/net.d/100-crio-bridge.conf /etc/cni/net.d/200-loopback.conf

# add default cni directory the config
perl -i -0pe 's#plugin_dirs\s*=\s*\[[^\]]*\]#plugin_dirs = [\n  "/opt/cni/bin",\n  "/usr/libexec/cni"\n]#g' /etc/crio/crio.conf

# enable systemd service after next boot
systemctl enable crio.service
