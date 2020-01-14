module sigs.k8s.io/cluster-api-provider-hetzner

go 1.12

require (
	github.com/aws/aws-sdk-go v1.25.43 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/hetznercloud/hcloud-go v1.17.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/apiserver v0.0.0-20190918200908-1e17798da8c1
	k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	sigs.k8s.io/cluster-api v0.2.7 // indirect
	sigs.k8s.io/cluster-api-provider-aws v0.4.6 // indirect
	sigs.k8s.io/controller-runtime v0.3.0
)
