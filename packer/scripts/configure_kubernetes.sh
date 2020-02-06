#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

KUBERNETES_VERSION=${PACKER_KUBERNETES_VERSION:-1.15.6-0}

env

cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF

yum install -y kubelet-${KUBERNETES_VERSION} kubeadm-${KUBERNETES_VERSION} kubectl-${KUBERNETES_VERSION} yum-plugin-versionlock bash-completion --disableexcludes=kubernetes
yum versionlock kubelet kubectl kubeadm
systemctl enable kubelet

mkdir -p /etc/systemd/system/kubelet.service.d
cat <<EOF > /etc/systemd/system/kubelet.service.d/11-cgroups.conf
[Service]
CPUAccounting=true
MemoryAccounting=true
EOF

cat <<EOF > /etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd
EOF

cat <<EOF > /etc/modules-load.d/k8s.conf
# load bridge netfilter
br_netfilter
EOF

cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
EOF

systemctl start crio
kubeadm config images pull

semanage fcontext -a -t container_file_t /var/lib/etcd
mkdir -p /var/lib/etcd
restorecon -rv /var /etc

# temporary fix for SELinux in kubernetes 1.16+
curl -Lo /usr/bin/kubelet https://github.com/simonswine/kubernetes/releases/download/v$(echo "${KUBERNETES_VERSION}" | cut -d "-" -f 1)/kubelet-linux-amd64
chmod +x /usr/bin/kubelet

# enable completion
echo 'source <(kubectl completion bash)' >>~/.bashrc

# set the kubeadm default path for kubeconfig
echo 'export KUBECONFIG=/etc/kubernetes/admin.conf' >>~/.bashrc
