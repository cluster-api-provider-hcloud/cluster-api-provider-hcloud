#!/bin/sh

sh ./configure_base.sh
sh ./configure_cri.sh
sh ./configure_kubernetes.sh
sh ./configure_kernel.sh
sh ./generate-tarball.sh