module github.com/simonswine/cluster-api-provider-hcloud

go 1.12

require (
	github.com/bronze1man/yaml2json v0.0.0-20190501122504-861f66b7262b
	github.com/fatih/color v1.7.0
	github.com/go-logr/logr v0.1.0
	github.com/golang/mock v1.4.1
	github.com/google/go-jsonnet v0.13.0
	github.com/hetznercloud/hcloud-go v1.17.0
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.6
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/apiserver v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/cluster-api v0.3.3
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/yaml v1.2.0
)
