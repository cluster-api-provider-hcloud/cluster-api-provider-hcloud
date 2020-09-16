#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# Set locale
localectl set-locale LANG=en_US.UTF-8 
localectl set-locale LANGUAGE=en_US.UTF-9

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



# Enable Repos
dnf config-manager --set-enabled AppStream BaseOS extras PowerTools 

# update all packages
dnf update -y

# install basic tooling
dnf -y install \
    git vim tmux at jq unzip htop wget\
    socat ipvsadm iperf3 mtr\
    nfs-utils \
    iscsi-initiator-utils \
    firewalld
