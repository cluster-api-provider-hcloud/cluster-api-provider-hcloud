#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# Prerequisites
modprobe overlay
modprobe br_netfilter


sysctl --system

# install cri-o
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

git clone https://github.com/cri-o/cri-o
cd cri-o
git checkout release-1.19

make
make install
make install.systemd

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


# Install runc
cd github.com/opencontainers
git clone https://github.com/opencontainers/runc
cd runc

make
make install

# enable systemd service after next boot
systemctl enable crio.service
systemctl daemon-reload
systemctl enable crio
