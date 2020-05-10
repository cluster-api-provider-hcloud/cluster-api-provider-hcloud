#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# update all packages
yum update -y

# install basic tooling
yum -y install epel-release
yum -y install \
    git vim tmux at jq unzip htop \
    socat ipvsadm \
    nfs-utils \
    iscsi-initiator-utils \
    firewalld

# disable portmapper rpcbind
systemctl disable rpcbind.service rpcbind.socket

# disable firewalld
systemctl disable firewalld.service

# disable kdump service
systemctl disable kdump.service

# mount bpfs for calico
cat > /etc/systemd/system/sys-fs-bpf.mount <<EOF
[Unit]
Description=Cilium BPF mounts
Documentation=http://docs.cilium.io/
DefaultDependencies=no
Before=local-fs.target umount.target
After=swap.target

[Mount]
What=bpffs
Where=/sys/fs/bpf
Type=bpf
Options=rw,nosuid,nodev,noexec,relatime,mode=700

[Install]
WantedBy=multi-user.target
EOF
systemctl enable sys-fs-bpf.mount

# Set SELinux in enforcing mode (effectively disabling it)
setenforce 1
sed -i 's/^SELINUX=permissive$/SELINUX=enforcing/' /etc/selinux/config
