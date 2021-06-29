# Do not load `http_archive` or `http_file` in this file.
# All fetches of external repositories should go in deps_bazel.bzl
# so that statements in this file are all order-dependent.
load("//:WORKSPACE_deps.bzl", "fetch_deps")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")
fetch_deps()

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.16.2")

# Load repositories from external files
# gazelle:repository_macro hack/build/repos.bzl%go_repositories
load("//hack/build:repos.bzl", "go_repositories")

go_repositories()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

container_repositories()

load("@io_bazel_rules_docker//repositories:deps.bzl", container_deps = "deps")

container_deps()

# load("@io_bazel_rules_docker//repositories:pip_repositories.bzl", "pip_deps")

# pip_deps()

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_pull",
)

## Use 'static' distroless image for all builds
container_pull(
    name = "static_base",
    registry = "index.docker.io",
    repository = "library/alpine",
    #tag = "3.11",
    digest = "sha256:ddba4d27a7ffc3f86dd6c2f92041af252a1f23a8e742c90e6e1297bfa1bc0c45",
)

## Setup jsonnet

load("@io_bazel_rules_jsonnet//jsonnet:jsonnet.bzl", "jsonnet_repositories")

jsonnet_repositories()

load("@jsonnet_go//bazel:repositories.bzl", "jsonnet_go_repositories")

jsonnet_go_repositories()

load("@jsonnet_go//bazel:deps.bzl", "jsonnet_go_dependencies")

jsonnet_go_dependencies()

load("@com_github_jemdiggity_rules_os_dependent_http_archive//:os_dependent_http_archive.bzl", "os_dependent_http_archive")

# Packer binary dependencies
PACKER_VERSION = "1.5.4"

http_archive(
    name = "packer_linux_amd64_bin",
    urls = ["https://releases.hashicorp.com/packer/%s/packer_%s_linux_amd64.zip" % (PACKER_VERSION, PACKER_VERSION)],
    sha256 = "c7277f64d217c7d9ccfd936373fe352ea935454837363293f8668f9e42d8d99d",
    build_file_content = '''filegroup(
    name="bin",
    srcs=["packer"],
    visibility = ["//visibility:public"],
)''',
)

http_archive(
    name = "packer_darwin_amd64_bin",
    urls = ["https://releases.hashicorp.com/packer/%s/packer_%s_darwin_amd64.zip" % (PACKER_VERSION, PACKER_VERSION)],
    sha256 = "dab5ab9d4801da5206755856bc3f026942ce18391419202a1b0b442c1c2e591d",
    build_file_content = '''filegroup(
    name="bin",
    srcs=["packer"],
    visibility = ["//visibility:public"],
)''',
)

# kubebuilder for testing our controllers
http_archive(
    name = "kubebuilder_linux_amd64_bin",
    urls = ["https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.2.0/kubebuilder_2.2.0_linux_amd64.tar.gz"],
    strip_prefix = "kubebuilder_2.2.0_linux_amd64/bin/",
    sha256 = "9ef35a4a4e92408f7606f1dd1e68fe986fa222a88d34e40ecc07b6ffffcc8c12",
    build_file_content = '''filegroup(
    name="bin",
    srcs=["etcd","kubectl","kube-apiserver", "kubebuilder"],
    visibility = ["//visibility:public"],
)''',
)

http_archive(
    name = "kubebuilder_darwin_amd64_bin",
    urls = ["https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.2.0/kubebuilder_2.2.0_darwin_amd64.tar.gz"],
    strip_prefix = "kubebuilder_2.2.0_darwin_amd64/bin/",
    sha256 = "5ccb9803d391e819b606b0c702610093619ad08e429ae34401b3e4d448dd2553",
    build_file_content = '''filegroup(
    name="bin",
    srcs=["etcd","kubectl","kube-apiserver", "kubebuilder"],
    visibility = ["//visibility:public"],
)''',
)

# kubectl binary dependencies
KUBECTL_VERSION = "1.17.2"

http_file(
    name = "kubectl_linux_amd64_bin",
    urls = ["https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubectl" % KUBECTL_VERSION],
    sha256 = "7732548b9c353114b0dfa173bc7bcdedd58a607a5b4ca49d867bdb4c05dc25a1",
    downloaded_file_path = "kubectl",
    executable = True,
)

http_file(
    name = "kubectl_darwin_amd64_bin",
    urls = ["https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubectl" % KUBECTL_VERSION],
    sha256 = "5d5bd9f88cc77fc51057641c46a2a73e6490550efa7c808f2d2e27a90cfe0c6e",
    downloaded_file_path = "kubectl",
    executable = True,
)
