load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_repositories():
    go_repository(
        name = "com_github_drone_envsubst",
        importpath = "github.com/drone/envsubst",
        sum = "h1:pf3CyiWgjOLL7cjFos89AEOPCWSOoQt7tgbEk/SvBAg=",
        version = "v1.0.3-0.20200709223903-efdb65b94e5a",
    )
    go_repository(
        name = "com_github_nxadm_tail",
        importpath = "github.com/nxadm/tail",
        sum = "h1:DQuhQpB1tVlglWS2hLQ5OV6B5r8aGxSrPc5Qo6uTN78=",
        version = "v1.4.4",
    )
    go_repository(
        name = "io_k8s_klog_v2",
        importpath = "k8s.io/klog/v2",
        sum = "h1:Foj74zO6RbjjP4hBEKjnYtjjAhGg4jNynUdYF6fJrok=",
        version = "v2.0.0",
    )
    go_repository(
        name = "io_k8s_sigs_structured_merge_diff_v2",
        importpath = "sigs.k8s.io/structured-merge-diff/v2",
        sum = "h1:I0h4buiCqDtPztO3NOiyoNMtqSIfld49D4Wj3UBXYZA=",
        version = "v2.0.1",
    )
    go_repository(
        name = "org_golang_google_protobuf",
        importpath = "google.golang.org/protobuf",
        sum = "h1:4MY060fB1DLGMB/7MBTLnwQUY6+F09GEiz6SsrNqyzM=",
        version = "v1.23.0",
    )
