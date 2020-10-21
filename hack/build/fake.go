package fake

// Commented out import ERROR: /home/runner/work/cluster-api-provider-hcloud/cluster-api-provider-hcloud/hack/build/BUILD.bazel:17:11: in go_library rule //hack/build:go_default_library: target '@com_github_tcnksm_ghr//:go_default_library' is not visible from target '//hack/build:go_default_library'. Check the visibility declaration of the former target if you think the dependency is legitimate
// +build ignore
import (
	_ "github.com/tcnksm/ghr"
	_ "sigs.k8s.io/controller-tools/pkg/version"
)

func init() {
}
