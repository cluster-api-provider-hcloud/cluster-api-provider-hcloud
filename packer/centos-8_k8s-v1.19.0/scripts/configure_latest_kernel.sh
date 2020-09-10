#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# configure elrepo
dnf -y install https://www.elrepo.org/elrepo-release-7.0-4.el7.elrepo.noarch.rpm

# enable repo and install latest kernel
dnf --enablerepo=elrepo-kernel install -y kernel-ml

# update grub to use latest kernel
grub2-set-default 0
grub2-mkconfig -o /boot/grub2/grub.cfg
