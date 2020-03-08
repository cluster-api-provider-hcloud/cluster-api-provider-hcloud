module github.com/simonswine/cluster-api-provider-hcloud

go 1.12

require (
	github.com/bronze1man/yaml2json v0.0.0-20190501122504-861f66b7262b
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/fatih/color v1.7.0
	github.com/go-logr/logr v0.1.0
	github.com/golang/mock v1.4.1
	github.com/google/go-jsonnet v0.13.0
	github.com/hetznercloud/hcloud-go v1.17.0
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.17.0 // indirect
	k8s.io/apimachinery v0.17.0
	k8s.io/apiserver v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/klog v1.0.0
	sigs.k8s.io/cluster-api v0.2.10
	sigs.k8s.io/cluster-api-bootstrap-provider-kubeadm v0.1.6
	sigs.k8s.io/controller-runtime v0.4.0
)
