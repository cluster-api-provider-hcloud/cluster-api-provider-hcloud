# How to add new packer config

1. Add new directory to packer use the same structure as the provided images.
1. Add under BUILD.bazel the new packer path
    - filegroup `"//packer/<directory-name>:all-srcs", `
1. Add under cmd/cluster-api-provider-hcloud BUILD.bazel the new packer path.
    - pkg_tar `"//packer/<directory-name>:packer",`
    - container_layer `"//packer/<director-name>:packer",`
1. Change under your new directory the BAZEL.build file under pkg_tar the key package_dir to `<directory-name>-packer-config`
