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
    go_repository(
        name = "com_github_onsi_ginkgo",
        importpath = "github.com/onsi/ginkgo",
        sum = "h1:mFwc4LvZ0xpSvDZ3E+k8Yte0hLOMxXUlP+yXtJqkYfQ=",
        version = "v1.12.1",
    )
    go_repository(
        name = "com_github_onsi_gomega",
        importpath = "github.com/onsi/gomega",
        sum = "h1:o0+MgICZLuZ7xjH7Vx6zS/zcu93/BEp1VwkIW1mEXCE=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_github_hashicorp_go_version",
        importpath = "github.com/hashicorp/go-version",
        sum = "h1:3vNe/fWF5CBgRIguda1meWhsZHy3m8gCJ5wx+dIzX/E=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_mitchellh_colorstring",
        importpath = "github.com/mitchellh/colorstring",
        sum = "h1:62I3jR2EmQ4l5rM/4FEfDWcRD+abF5XlKShorW5LRoQ=",
        version = "v0.0.0-20190213212951-d06e56a500db",
    )
    go_repository(
        name = "com_github_songmu_retry",
        importpath = "github.com/Songmu/retry",
        sum = "h1:hPA5xybQsksLR/ry/+t/7cFajPW+dqjmjhzZhioBILA=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_tcnksm_ghr",
        importpath = "github.com/tcnksm/ghr",
        sum = "h1:a5ZbaUAfiaiw6rEDJVUEDYA9YreZOkh3XAfXHWn8zu8=",
        version = "v0.13.0",
    )
    go_repository(
        name = "com_github_tcnksm_go_gitconfig",
        importpath = "github.com/tcnksm/go-gitconfig",
        sum = "h1:iiDhRitByXAEyjgBqsKi9QU4o2TNtv9kPP3RgPgXBPw=",
        version = "v0.1.2",
    )
    go_repository(
        name = "com_github_tcnksm_go_latest",
        importpath = "github.com/tcnksm/go-latest",
        sum = "h1:IWllFTiDjjLIf2oeKxpIUmtiDV5sn71VgeQgg6vcE7k=",
        version = "v0.0.0-20170313132115-e3007ae9052e",
    )
