#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

KUBERNETES_VERSION=${PACKER_KUBERNETES_VERSION:-1.16.7}
KUBERNETES_VERSION_RPM=${PACKER_KUBERNETES_VERSION_RPM:-$KUBERNETES_VERSION-0}

kubernetes_selinux_fix_versions=(
    "1.16.6"
    "1.16.7"
    "1.17.2"
    "1.17.3"
)

kubernetes_selinux_fix_hash=(
    "aba7f145f269a9f5ffa765b89f6627ab79d88d0d23e565ed9bb20c0b5e754ea8"
    "e1d293305de50ee5fca821324ce298166617a9965161155e9e8fbd4fd43ff1f0"
    "4064e6f5679912334ee3df4187ac050309fe69e51b2631a76ad4f0f9f897b36a"
    "73633120d7863ea0b56dfcaa37ca8e9826a4fa45075036dfd2c165c9bbecbc36"
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
    if ((${version_split[1]} < 16 )); then
        return 0
    fi

    # Everything after 1.17 is fine
    if ((${version_split[1]} > 17 )); then
        return 0
    fi

    # 1.16.8 will release a fix
    if ((${version_split[1]} == 16 )) && ((${version_split[2]} > 7 )); then
        return 0
    fi

    # 1.17.4 will release a fix
    if ((${version_split[1]} == 17 )) && ((${version_split[2]} > 3 )); then
        return 0
    fi

    for e in ${kubernetes_selinux_fix_versions[@]}; do
        [[ "$e" == "$1" ]] && return 1
    done

    return 2
}

download_hotfix_pos () {
    local path=/usr/bin/kubelet
    curl -Lo "${path}" "https://github.com/simonswine/kubernetes/releases/download/v${1}/kubelet-linux-amd64"
    echo "${kubernetes_selinux_fix_hash[$2]} ${path}" | sha256sum -c
    chmod +x "${path}"
}

download_hotfix () {
    echo "Download hotfix for ${1}"
    local pos=0
    for e in ${kubernetes_selinux_fix_versions[@]}; do
        [[ "$e" == "$1" ]] && download_hotfix_pos $1 $pos
        pos=$((pos+1))
    done
}

# test if the kubernetes version is supported
handle_version () {
    test_version $1 && ret=$? || ret=$?
    case $ret in
        1)
            echo "Kubernetes release ${1} has an existing SELinux hotfix"
            download_hotfix ${1}
            ;;
        2)
            echo "Kubernetes release ${1} is not working with SELinux as other 1.16 or 1.17 releases (cf.)" > /dev/stderr
            exit 2
            ;;
    esac
}


cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF

yum install -y kubelet-${KUBERNETES_VERSION_RPM} kubeadm-${KUBERNETES_VERSION_RPM} kubectl-${KUBERNETES_VERSION_RPM} yum-plugin-versionlock bash-completion --disableexcludes=kubernetes
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

# temporary fix for SELinux in kubernetes 1.16+,1.17+
handle_version $KUBERNETES_VERSION

# enable completion
echo 'source <(kubectl completion bash)' >>~/.bashrc

# set the kubeadm default path for kubeconfig
echo 'export KUBECONFIG=/etc/kubernetes/admin.conf' >>~/.bashrc
