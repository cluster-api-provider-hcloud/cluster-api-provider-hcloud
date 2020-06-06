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
yum -y install git vim tmux socat at jq unzip htop ipvsadm

# disable kdump service
systemctl disable kdump.service

# Set SELinux in enforcing mode (effectively disabling it)
setenforce 1
sed -i 's/^SELINUX=permissive$/SELINUX=enforcing/' /etc/selinux/config
