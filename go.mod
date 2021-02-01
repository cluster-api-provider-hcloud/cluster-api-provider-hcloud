module github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud

go 1.15

require (
	github.com/bronze1man/yaml2json v0.0.0-20190501122504-861f66b7262b
	github.com/fatih/color v1.9.0
	github.com/go-logr/logr v0.3.0
	github.com/golang/mock v1.4.3
	github.com/google/go-jsonnet v0.16.0
	github.com/hetznercloud/hcloud-go v1.22.0
	github.com/nl2go/hrobot-go v0.1.3
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/tcnksm/ghr v0.13.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/sys v0.0.0-20201005172224-997123666555 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/apiserver v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/cluster-api v0.3.9
	sigs.k8s.io/controller-runtime v0.7.2
	sigs.k8s.io/controller-tools v0.3.0
	sigs.k8s.io/yaml v1.2.0
)
