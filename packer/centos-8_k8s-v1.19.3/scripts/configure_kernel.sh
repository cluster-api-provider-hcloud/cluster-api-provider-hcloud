#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

# Add ELRepo
dnf -y install https://www.elrepo.org/elrepo-release-8.el8.elrepo.noarch.rpm
rpm --import https://www.elrepo.org/RPM-GPG-KEY-elrepo.org

# install kernel
dnf -y --enablerepo=elrepo-kernel install kernel-ml

# update grub to use latest kernel
grub2-set-default 0
grub2-mkconfig -o /boot/grub2/grub.cfg
