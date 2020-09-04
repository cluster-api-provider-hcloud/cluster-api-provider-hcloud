#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# hardcode a particular centos release
centos_release=${PACKER_CENTOS_RELEASE:-7}
cat > /etc/yum.repos.d/CentOS-Base.repo <<EOF
# CentOS-Base.repo
[base]
name=CentOS-${centos_release} - Base
baseurl=http://mirror.centos.org/centos/${centos_release}/os/\$basearch/
        http://vault.centos.org/centos/${centos_release}/os/\$basearch/
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
#released updates
[updates]
name=CentOS-${centos_release} - Updates
baseurl=http://mirror.centos.org/centos/${centos_release}/updates/\$basearch/
        http://vault.centos.org/centos/${centos_release}/updates/\$basearch/
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
#additional packages that may be useful
[extras]
name=CentOS-${centos_release} - Extras
baseurl=http://mirror.centos.org/centos/${centos_release}/extras/\$basearch/
        http://vault.centos.org/centos/${centos_release}/extras/\$basearch/
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
EOF

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
