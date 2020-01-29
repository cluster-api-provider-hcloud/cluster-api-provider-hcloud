module sigs.k8s.io/cluster-api-provider-hetzner

go 1.12

require (
	github.com/aws/aws-sdk-go v1.25.43 // indirect
	github.com/bronze1man/go-yaml2json v0.0.0-20150129175009-f6f64b738964
	github.com/bronze1man/yaml2json v0.0.0-20190501122504-861f66b7262b // indirect
	github.com/coreos/go-systemd v0.0.0-20180511133405-39ca1b05acc7
	github.com/go-delve/delve v1.3.2 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/golang/mock v1.3.1
	github.com/hetznercloud/hcloud-go v1.17.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	gopkg.in/yaml.v2 v2.2.4
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/apiserver v0.0.0-20190918200908-1e17798da8c1
	k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	k8s.io/klog v1.0.0
	sigs.k8s.io/cluster-api v0.2.7
	sigs.k8s.io/cluster-api-provider-aws v0.4.6 // indirect
	sigs.k8s.io/controller-runtime v0.3.0
)
