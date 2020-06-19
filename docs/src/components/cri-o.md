# cri-o

[CRI-O] is an implementation of the Kubernetes CRI (Container Runtime Interface)
to enable using OCI (Open Container Initiative) compatible runtimes. It is a
lightweight alternative to using Docker as the runtime for kubernetes. It
allows Kubernetes to use any OCI-compliant runtime as the container runtime for
running pods. Today it supports runc and Kata Containers as the container
runtimes but any OCI-conformant runtime can be plugged in principle.

CRI-O supports OCI container images and can pull from any container registry.
It is a lightweight alternative to using Docker, Moby or rkt as the runtime for
Kubernetes.

## Packages

As there is not a good package repo for Centos packages of CRI-O, this projects
builds its own packages through [COPR]

### How to build a new cri-o version

```bash
# Create a clone of package source
git clone git@github.com:simonswine/fedora-rpm-crio.git
cd fedora-rpm-crio
git remote add upstream https://src.fedoraproject.org/rpms/cri-o.git

# Checkout a minor release
git checkout 1.18

# Potentially get updates from fedora
git pull upstream 1.18

# Submit package build to COPR
# Eventually the source needs to be downloaded before this step
rpkg build simonswine/cri-o --spec cri-o.spec

# Wait for successful build
copr watch-build $NUMBER
```

[CRI-O]: https://cri-o.io/
[COPR]: https://copr.fedorainfracloud.org/coprs/simonswine/cri-o/
