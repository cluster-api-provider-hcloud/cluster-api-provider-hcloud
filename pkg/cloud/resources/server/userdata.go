package server

import (
	"bytes"
	"fmt"
	"text/template"

	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
)

var ipTablesProxyTemplate = template.Must(template.New("").Parse(`ExecStartPre=/bin/sh -c "iptables -t nat -C OUTPUT -d {{.destination}} -p tcp -m tcp --dport {{.destinationPort}} -j DNAT --to-destination 127.0.0.1:{{.localPort}} || iptables -t nat -I OUTPUT -d {{.destination}} -p tcp -m tcp --dport {{.destinationPort}} -j DNAT --to-destination 127.0.0.1:{{.localPort}}"
`))

func (s *Service) getIPTablesProxyFile() (bootstrapv1.File, error) {
	b := bytes.NewBuffer([]byte("[Service]\n"))
	port := s.scope.ControlPlaneAPIEndpointPort()

	for _, floatingIP := range s.scope.HcloudCluster.Status.ControlPlaneFloatingIPs {
		if floatingIP.Type == infrav1.HcloudFloatingIPTypeIPv4 {
			if err := ipTablesProxyTemplate.Execute(b, map[string]interface{}{
				"destination":     fmt.Sprintf("%s/32", floatingIP.IP),
				"localPort":       port,
				"destinationPort": port,
			}); err != nil {
				return bootstrapv1.File{}, err
			}
		}
	}

	return bootstrapv1.File{
		Content:     b.String(),
		Path:        "/etc/systemd/system/kubelet.service.d/20-iptables-redirect.conf",
		Owner:       "root:root",
		Permissions: "0644",
	}, nil
}
