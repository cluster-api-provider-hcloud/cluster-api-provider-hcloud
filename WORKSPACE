load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "a8d6b1b354d371a646d2f7927319974e0f9e52f73a2452d2b3877118169eb6bb",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.23.3/rules_go-v0.23.3.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.23.3/rules_go-v0.23.3.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "cdb02a887a7187ea4d5a27452311a75ed8637379a1287d8eeb952138ea485f7d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

## Load kubernetes repo-infra for tools like kazel
git_repository(
    name = "io_k8s_repo_infra",
    commit = "967e39a37fb93640a37e272949ddf92a8c96f230",
    remote = "https://github.com/kubernetes/repo-infra.git",
    shallow_since = "1569300445 -0700",
)

# Load repositories from external files
# gazelle:repository_macro hack/build/repos.bzl%go_repositories
load("//hack/build:repos.bzl", "go_repositories")

go_repositories()

## Load rules_docker and depdencies, for working with docker images
# Download the rules_docker repository at release v0.14.1
http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "dc97fccceacd4c6be14e800b2a00693d5e8d07f69ee187babfd04a80a9f8e250",
    strip_prefix = "rules_docker-0.14.1",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.14.1/rules_docker-v0.14.1.tar.gz"],
)

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

container_repositories()

load("@io_bazel_rules_docker//repositories:deps.bzl", container_deps = "deps")

container_deps()

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
