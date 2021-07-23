#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# Set locale
localectl set-locale LANG=en_US.UTF-8 

# Ensure that the correct repos are installed
cat > /etc/yum.repos.d/CentOS-Base.repo <<EOF

[BaseOS]
name=CentOS-\$releasever - Base
mirrorlist=http://mirrorlist.centos.org/?release=\$releasever&arch=\$basearch&repo=BaseOS&infra=\$infra
#baseurl=http://mirror.centos.org/\$contentdir/\$releasever/BaseOS/\$basearch/os/
gpgcheck=1
enabled=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-centosofficial

EOF

cat > /etc/yum.repos.d/CentOS-AppStream.repo <<EOF

[AppStream]
name=CentOS-\$releasever - AppStream
mirrorlist=http://mirrorlist.centos.org/?release=\$releasever&arch=\$basearch&repo=AppStream&infra=\$infra
#baseurl=http://mirror.centos.org/\$contentdir/\$releasever/AppStream/\$basearch/os/
gpgcheck=1
enabled=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-centosofficial
EOF

cat > /etc/yum.repos.d/CentOS-Extras.repo <<EOF

#additional packages that may be useful
[extras]
name=CentOS-\$releasever - Extras
mirrorlist=http://mirrorlist.centos.org/?release=\$releasever&arch=\$basearch&repo=extras&infra=\$infra
#baseurl=http://mirror.centos.org/\$contentdir/\$releasever/extras/\$basearch/os/
gpgcheck=1
enabled=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-centosofficial
EOF

# Add Extra Packages for Enterprise Linux (EPEL) 8
dnf -y install dnf-plugins-core

dnf install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm

# Enable Repos
dnf config-manager --set-enabled AppStream BaseOS powertools

# update all packages
dnf update -y

# install basic tooling
dnf -y install \
    git vim tmux at jq unzip htop wget\
    socat ipvsadm iperf3 mtr\
    nfs-utils \
    iscsi-initiator-utils \
    firewalld

# disable portmapper rpcbind
systemctl disable rpcbind.service rpcbind.socket

# disable firewalld
systemctl disable firewalld.service

# disable kdump service
systemctl disable kdump.service

# mount bpfs for cilium
cat > /etc/systemd/system/sys-fs-bpf.mount <<EOF
[Unit]
Description=Cilium BPF mounts
Documentation=https://docs.cilium.io/
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
sed -i 's/^SELINUX=permissive\$/SELINUX=enforcing/' /etc/selinux/config
