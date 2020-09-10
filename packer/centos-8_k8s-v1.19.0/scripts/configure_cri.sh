#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# Prerequisites
modprobe overlay
modprobe br_netfilter

sysctl --system

## for cri-o
dnf module -y install go-toolset

dnf install -y \
  containers-common \
  device-mapper-devel \
  git \
  make \
  glib2-devel \
  glibc-devel \
  glibc-static \
  runc \
  go \
  gpgme-devel \
  libassuan \
  libassuan-devel \
  libgpg-error \
  libgpg-error-devel \
  libseccomp \
  libselinux \
  libseccomp-devel \
  libselinux-devel \
  pkgconfig \
  pkgconf-pkg-config \
  runc

go get github.com/cpuguy83/go-md2man  


# Install runc
wget https://github.com/opencontainers/runc/releases/download/v1.0.0-rc92/runc.amd64 -O /usr/local/sbin/runc && chmod +x /usr/local/sbin/runc

# Install conmon
wget https://github.com/containers/conmon/releases/download/v2.0.20/conmon -O /usr/local/bin/conmon && chmod +x /usr/local/bin/conmon

# install cri-o
wget https://github.com/cri-o/cri-o/archive/v1.19.0-rc.1.tar.gz
mkdir /tmp/crio && tar zxvf v1.19.0-rc.1.tar.gz -C /tmp/crio --strip-components 1
(cd /tmp/crio && make)
(cd /tmp/crio && sudo make install)
(cd /tmp/crio && sudo make install.config)
(cd /tmp/crio && make install.systemd)
rm -f v1.19.0-rc.1.tar.gz
rm -rf /tmp/crio

# cri-tool https://github.com/kubernetes-sigs/cri-tools
# Install crictl
wget https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.19.0/crictl-v1.19.0-linux-amd64.tar.gz
tar zxvf crictl-v1.19.0-linux-amd64.tar.gz -C /usr/local/bin 
rm -f crictl-v1.19.0-linux-amd64.tar.gz

# Install critest
wget https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.19.0/critest-v1.19.0-linux-amd64.tar.gz
tar zxvf critest-v1.19.0-linux-amd64.tar.gz -C /usr/local/bin
rm -f critest-v1.19.0-linux-amd64.tar.gz

# remove default CNIs
rm -f /etc/cni/net.d/100-crio-bridge.conf /etc/cni/net.d/200-loopback.conf

# add default cni directory the config
perl -i -0pe 's#plugin_dirs\s*=\s*\[[^\]]*\]#plugin_dirs = [\n  "/opt/cni/bin",\n  "/usr/libexec/cni"\n]#g' /etc/crio/crio.conf


# enable systemd service after next boot
systemctl enable crio.service
systemctl daemon-reload
systemctl enable crio
