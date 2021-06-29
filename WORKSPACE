load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "69de5c704a05ff37862f7e0f5534d4f479418afc21806c887db544a316f3cb6b",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.27.0/rules_go-v0.27.0.tar.gz",
    ],
)

http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "59d5b42ac315e7eadffa944e86e90c2990110a1c8075f1cd145f487e999d22b3",
    strip_prefix = "rules_docker-0.17.0",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.17.0/rules_docker-v0.17.0.tar.gz"],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "62ca106be173579c0a167deb23358fdfe71ffa1e4cfdddf5582af26520f1c66f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.23.0/bazel-gazelle-v0.23.0.tar.gz",
    ],
)

http_archive(
    name = "io_k8s_repo_infra",
    sha256 = "5baf1b698bc20080ba714401c34ddc6ffa447757896c26399c00ab8bb39cc980",
    strip_prefix = "repo-infra-0.1.8",
    urls = [
        "https://github.com/kubernetes/repo-infra/archive/v0.1.8.tar.gz",
    ],
)

http_archive(
    name = "io_bazel_rules_jsonnet",
    sha256 = "7f51f859035cd98bcf4f70dedaeaca47fe9fbae6b199882c516d67df416505da",
    strip_prefix = "rules_jsonnet-0.3.0",
    urls = ["https://github.com/bazelbuild/rules_jsonnet/archive/0.3.0.tar.gz"],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.16.2")

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

## Load kubernetes repo-infra for tools like kazel

# Load repositories from external files
# gazelle:repository_macro hack/build/repos.bzl%go_repositories
load("//hack/build:repos.bzl", "go_repositories")

go_repositories()

## Load rules_docker and depdencies, for working with docker images
# Download the rules_docker repository at release v0.14.4

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
    #tag = "3.11",
    digest = "sha256:ddba4d27a7ffc3f86dd6c2f92041af252a1f23a8e742c90e6e1297bfa1bc0c45",
    registry = "index.docker.io",
    repository = "library/alpine",
)

## Setup jsonnet

load("@io_bazel_rules_jsonnet//jsonnet:jsonnet.bzl", "jsonnet_repositories")

jsonnet_repositories()

load("@jsonnet_go//bazel:repositories.bzl", "jsonnet_go_repositories")

jsonnet_go_repositories()

load("@jsonnet_go//bazel:deps.bzl", "jsonnet_go_dependencies")

jsonnet_go_dependencies()

git_repository(
    name = "com_github_jemdiggity_rules_os_dependent_http_archive",
    commit = "b1e3ed2fd829dfd1602bc31df4804ff34149f659",
    remote = "https://github.com/jemdiggity/rules_os_dependent_http_archive.git",
)

load("@com_github_jemdiggity_rules_os_dependent_http_archive//:os_dependent_http_archive.bzl", "os_dependent_http_archive")

# Packer binary dependencies
PACKER_VERSION = "1.5.4"

http_archive(
    name = "packer_linux_amd64_bin",
    build_file_content = '''filegroup(
    name="bin",
    srcs=["packer"],
    visibility = ["//visibility:public"],
)''',
    sha256 = "c7277f64d217c7d9ccfd936373fe352ea935454837363293f8668f9e42d8d99d",
    urls = ["https://releases.hashicorp.com/packer/%s/packer_%s_linux_amd64.zip" % (PACKER_VERSION, PACKER_VERSION)],
)

http_archive(
    name = "packer_darwin_amd64_bin",
    build_file_content = '''filegroup(
    name="bin",
    srcs=["packer"],
    visibility = ["//visibility:public"],
)''',
    sha256 = "dab5ab9d4801da5206755856bc3f026942ce18391419202a1b0b442c1c2e591d",
    urls = ["https://releases.hashicorp.com/packer/%s/packer_%s_darwin_amd64.zip" % (PACKER_VERSION, PACKER_VERSION)],
)

# kubebuilder for testing our controllers
http_archive(
    name = "kubebuilder_linux_amd64_bin",
    build_file_content = '''filegroup(
    name="bin",
    srcs=["etcd","kubectl","kube-apiserver", "kubebuilder"],
    visibility = ["//visibility:public"],
)''',
    sha256 = "9ef35a4a4e92408f7606f1dd1e68fe986fa222a88d34e40ecc07b6ffffcc8c12",
    strip_prefix = "kubebuilder_2.2.0_linux_amd64/bin/",
    urls = ["https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.2.0/kubebuilder_2.2.0_linux_amd64.tar.gz"],
)

http_archive(
    name = "kubebuilder_darwin_amd64_bin",
    build_file_content = '''filegroup(
    name="bin",
    srcs=["etcd","kubectl","kube-apiserver", "kubebuilder"],
    visibility = ["//visibility:public"],
)''',
    sha256 = "5ccb9803d391e819b606b0c702610093619ad08e429ae34401b3e4d448dd2553",
    strip_prefix = "kubebuilder_2.2.0_darwin_amd64/bin/",
    urls = ["https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.2.0/kubebuilder_2.2.0_darwin_amd64.tar.gz"],
)

# kubectl binary dependencies
KUBECTL_VERSION = "1.17.2"

http_file(
    name = "kubectl_linux_amd64_bin",
    downloaded_file_path = "kubectl",
    executable = True,
    sha256 = "7732548b9c353114b0dfa173bc7bcdedd58a607a5b4ca49d867bdb4c05dc25a1",
    urls = ["https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubectl" % KUBECTL_VERSION],
)

http_file(
    name = "kubectl_darwin_amd64_bin",
    downloaded_file_path = "kubectl",
    executable = True,
    sha256 = "5d5bd9f88cc77fc51057641c46a2a73e6490550efa7c808f2d2e27a90cfe0c6e",
    urls = ["https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubectl" % KUBECTL_VERSION],
)
