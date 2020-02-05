# gazelle:repository_macro hack/build/repos.bzl%go_repositories

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

http_archive(
    name = "io_bazel_rules_go",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.21.2/rules_go-v0.21.2.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.21.2/rules_go-v0.21.2.tar.gz",
    ],
    sha256 = "f99a9d76e972e0c8f935b2fe6d0d9d778f67c760c6d2400e23fc2e469016e2bd",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

http_archive(
    name = "bazel_gazelle",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/bazel-gazelle/releases/download/v0.19.1/bazel-gazelle-v0.19.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.19.1/bazel-gazelle-v0.19.1.tar.gz",
    ],
    sha256 = "86c6d481b3f7aedc1d60c1c211c6f76da282ae197c3b3160f54bd3a8f847896f",
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

# Load repositories from external files
load("//hack/build:repos.bzl", "go_repositories")

go_repositories()

## Load rules_docker and depdencies, for working with docker images
git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    commit = "80ea3aae060077e5fe0cdef1a5c570d4b7622100",
    shallow_since = "1561646721 -0700",
)

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

container_repositories()

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
http_archive(
    name = "io_bazel_rules_jsonnet",
    sha256 = "68b5bcb0779599065da1056fc8df60d970cffe8e6832caf13819bb4d6e832459",
    strip_prefix = "rules_jsonnet-0.2.0",
    urls = ["https://github.com/bazelbuild/rules_jsonnet/archive/0.2.0.tar.gz"],
)

load("@io_bazel_rules_jsonnet//jsonnet:jsonnet.bzl", "jsonnet_repositories")

jsonnet_repositories()

load("@jsonnet_go//bazel:repositories.bzl", "jsonnet_go_repositories")

jsonnet_go_repositories()

load("@jsonnet_go//bazel:deps.bzl", "jsonnet_go_dependencies")

jsonnet_go_dependencies()

git_repository(
    name = "com_github_jemdiggity_rules_os_dependent_http_archive",
    remote = "https://github.com/jemdiggity/rules_os_dependent_http_archive.git",
    commit = "b1e3ed2fd829dfd1602bc31df4804ff34149f659",
)

load("@com_github_jemdiggity_rules_os_dependent_http_archive//:os_dependent_http_archive.bzl", "os_dependent_http_archive")

# Packer binary dependency
http_archive(
    name = "packer_linux_amd64_bin",
    urls = ["https://releases.hashicorp.com/packer/1.5.1/packer_1.5.1_linux_amd64.zip"],
    sha256 = "3305ede8886bc3fd83ec0640fb87418cc2a702b2cb1567b48c8cb9315e80047d",
    build_file_content = '''filegroup(
    name="bin",
    srcs=["packer"],
    visibility = ["//visibility:public"],
)''',
)

# kubecfg binary dependency
http_file(
    name = "kubecfg_linux_amd64_bin",
    urls = ["https://github.com/bitnami/kubecfg/releases/download/v0.14.0/kubecfg-linux-amd64"],
    sha256 = "bb1455ec70f93d6f0fd344becec2f1617837a879e8363272d3216bf44c04cb2c",
    downloaded_file_path = "kubecfg",
    executable = True,
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
