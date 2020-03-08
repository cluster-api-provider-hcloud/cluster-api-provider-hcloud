package server

import (
	"context"
	"fmt"
	"reflect"

	"github.com/coreos/go-systemd/unit"
	"github.com/pkg/errors"
	bootstrapv1 "sigs.k8s.io/cluster-api-bootstrap-provider-kubeadm/api/v1alpha2"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api-bootstrap-provider-kubeadm/kubeadm/v1beta1"
)

const (
	kubeadmConfigPath              = "/tmp/kubeadm.yaml"
	kubeadmConfigPartsPathTemplate = "/tmp/kubeadm.parts.%s.yaml"
)

const systemdMountTemplate = `[Mount]
What=/dev/disk/by-id/scsi-0HC_Volume_%d
Where=%s
Type=ext4
Options=discard,defaults

[Install]
WantedBy=local-fs.target
`
const systemdUnitAfterRequires = `[Unit]
After=%s
Requires=%s
`

type kubeadmConfig struct {
	s *Service

	configBefore *bootstrapv1.KubeadmConfig
}

func (s *Service) newKubeadmConfig(c *bootstrapv1.KubeadmConfig) *kubeadmConfig {
	return &kubeadmConfig{
		s:            s,
		configBefore: c.DeepCopy(),
	}
}

func kubeadmConfigPartsPath(priority int, name string) string {
	return fmt.Sprintf(
		fmt.Sprintf(kubeadmConfigPartsPathTemplate, "%02d-%s"),
		priority,
		name,
	)
}

func (k *kubeadmConfig) addWaitForMount(unitPath string, mountPath string) {
	unitName := fmt.Sprintf("%s.mount", unit.UnitNamePathEscape(mountPath))
	k.addOrUpdateFile(bootstrapv1.File{
		Path:        fmt.Sprintf("/etc/systemd/system/%s", unitPath),
		Content:     fmt.Sprintf(systemdUnitAfterRequires, unitName, unitName),
		Permissions: "0644",
		Owner:       "root:root",
	})
}

func (k *kubeadmConfig) addVolumeMount(id int64, mountPath string) {
	k.addOrUpdateFile(bootstrapv1.File{
		Path:        fmt.Sprintf("/etc/systemd/system/%s.mount", unit.UnitNamePathEscape(mountPath)),
		Content:     fmt.Sprintf(systemdMountTemplate, id, mountPath),
		Permissions: "0644",
		Owner:       "root:root",
	})
}

func (k *kubeadmConfig) addOrUpdateFile(f bootstrapv1.File) {
	for pos := range k.s.scope.KubeadmConfig.Spec.Files {
		if k.s.scope.KubeadmConfig.Spec.Files[pos].Path == f.Path {
			k.s.scope.KubeadmConfig.Spec.Files[pos] = f
			return
		}
	}
	k.s.scope.KubeadmConfig.Spec.Files = append(
		k.s.scope.KubeadmConfig.Spec.Files,
		f,
	)
}

// update updates the kubeadmConfig if its spec has changed. If the object has
// changed, it will return a resource version
func (k *kubeadmConfig) update(ctx context.Context) (*string, error) {
	// if the config is the same return now
	if reflect.DeepEqual(k.configBefore.Spec, k.s.scope.KubeadmConfig.Spec) {
		return nil, nil
	}

	resourceVersionBefore := k.s.scope.KubeadmConfig.ObjectMeta.ResourceVersion
	if err := k.s.scope.Client.Update(ctx, k.s.scope.KubeadmConfig); err != nil {
		return nil, errors.Wrap(err, "error updating changed KubeadmConfig")
	}

	k.s.scope.V(1).Info(
		"kubeadmConfig updated",
		"resourceVersionBefore",
		resourceVersionBefore,
		"resourceVersionAfter",
		k.s.scope.KubeadmConfig.ObjectMeta.ResourceVersion,
	)

	return &k.s.scope.KubeadmConfig.ObjectMeta.ResourceVersion, nil
}

func stringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (k *kubeadmConfig) ensureKubeadmConfigParts() {
	cmdMoveConfig := fmt.Sprintf(
		`mv "%s" "%s"`,
		kubeadmConfigPath,
		kubeadmConfigPartsPath(20, "base"),
	)
	if !stringSliceContains(k.s.scope.KubeadmConfig.Spec.PreKubeadmCommands, cmdMoveConfig) {
		k.s.scope.KubeadmConfig.Spec.PreKubeadmCommands = append(
			[]string{cmdMoveConfig},
			k.s.scope.KubeadmConfig.Spec.PreKubeadmCommands...,
		)
	}

	cmdAggregateConfig := fmt.Sprintf(
		`cat "%s" > "%s"`,
		fmt.Sprintf(kubeadmConfigPartsPathTemplate, `"*"`),
		kubeadmConfigPath,
	)
	if !stringSliceContains(k.s.scope.KubeadmConfig.Spec.PreKubeadmCommands, cmdAggregateConfig) {
		k.s.scope.KubeadmConfig.Spec.PreKubeadmCommands = append(
			k.s.scope.KubeadmConfig.Spec.PreKubeadmCommands,
			cmdAggregateConfig,
		)
	}
}

func (k *kubeadmConfig) addControlPlaneConfig() {
	// TODO: set external cloud-provider if not set
	// remove the controlPlaneEndpoint to make api server connect locally
	cmdRemoveControlPlane := fmt.Sprintf(
		"sed -i '/controlPlaneEndpoint: /d' '%s'",
		kubeadmConfigPath,
	)
	if !stringSliceContains(k.s.scope.KubeadmConfig.Spec.PreKubeadmCommands, cmdRemoveControlPlane) {
		k.s.scope.KubeadmConfig.Spec.PreKubeadmCommands = append(
			k.s.scope.KubeadmConfig.Spec.PreKubeadmCommands,
			cmdRemoveControlPlane,
		)
	}

	// add extra dns names to apiserver certificate
	if k.s.scope.KubeadmConfig.Spec.ClusterConfiguration == nil {
		k.s.scope.KubeadmConfig.Spec.ClusterConfiguration = &kubeadmv1beta1.ClusterConfiguration{}
	}
	apiServer := &k.s.scope.KubeadmConfig.Spec.ClusterConfiguration.APIServer

	// add localhost and floating point addresses
	hosts := []string{"127.0.0.1", "localhost"}
	for _, apiEndpoint := range k.s.scope.Cluster.Status.APIEndpoints {
		hosts = append(hosts, apiEndpoint.Host)
	}
	for _, host := range hosts {
		if !stringSliceContains(apiServer.CertSANs, host) {
			apiServer.CertSANs = append(
				apiServer.CertSANs,
				host,
			)
		}
	}
}

func (k *kubeadmConfig) addKubeletConfigTLSBootstrap() {
	k.ensureKubeadmConfigParts()
	k.addOrUpdateFile(bootstrapv1.File{
		Path:        kubeadmConfigPartsPath(80, "kubelet-tls"),
		Permissions: "0644",
		Owner:       "root:root",
		Content: `---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
rotateCertificates: true
serverTLSBootstrap: true
`,
	})
}
