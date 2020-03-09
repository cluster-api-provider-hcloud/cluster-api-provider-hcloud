#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

KUBERNETES_VERSION=${PACKER_KUBERNETES_VERSION:-1.16.6}
KUBERNETES_VERSION_RPM=${PACKER_KUBERNETES_VERSION_RPM:-$KUBERNETES_VERSION-0}

kubernetes_selinux_fix_versions=(
    "1.16.6"
    "1.17.2"
)

kubernetes_selinux_fix_hash=(
    "aba7f145f269a9f5ffa765b89f6627ab79d88d0d23e565ed9bb20c0b5e754ea8"
    "4064e6f5679912334ee3df4187ac050309fe69e51b2631a76ad4f0f9f897b36a"
)

# Tests kubernetes version (returns 0 if everthing is ok, 1 if it needs to
# download a selinux fix, 2 if the version would not work)
test_version () {
    local IFS=.
    local version_split=($1)
    unset IFS

    # Return early if we are not speaking about 1.x
    if (("${version_split[0]}" != 1)); then
        return 0
    fi

    # Everything before 1.16 is fine
    if ((10#${version_split[1]} < 16 )); then
        return 0
    fi

    # Everything after 1.17 is fine
    if ((10#${version_split[1]} > 17 )); then
        return 0
    fi

    # 1.16.8 will release a fix
    if ((10#${version_split[1]} == 16 )) && ((10#${version_split[2]} > 7 )); then
        return 0
    fi

    # 1.17.4 will release a fix
    if ((10#${version_split[1]} == 17 )) && ((10#${version_split[2]} > 3 )); then
        return 0
    fi

    for e in ${kubernetes_selinux_fix_versions[@]}; do
        [[ "$e" == "$1" ]] && return 1
    done

    return 2
}

test_version_string () {
    test_version $1
    case $? in
        0) str='fine';;
        1) str='selinux-fixed';;
        2) str='not-avail';;
    esac
    echo $str
}

echo $(test_version_string $KUBERNETES_VERSION)
echo $(test_version_string 2.18.0)
echo $(test_version_string 1.18.0)
echo $(test_version_string 1.17.0)
echo $(test_version_string 1.17.4)
echo $(test_version_string 1.15.34)
echo $(test_version_string 1.16.1)
echo $(test_version_string 1.16.6)
echo $(test_version_string 1.16.7)
echo $(test_version_string 1.16.8)

exit 1

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
