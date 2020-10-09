package userdata

// TODO: Fix error when running this test
/*
import (
	"bytes"
	"strings"
	"testing"

	"gotest.tools/assert"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
)

var userdataControlPlaneInit = `## template: jinja
#cloud-config

write_files:
  - path: /etc/kubernetes/pki/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/etcd/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/etcd/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/front-proxy-ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/front-proxy-ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/sa.pub
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN PUBLIC KEY-----
        -----END PUBLIC KEY-----
  - path: /etc/kubernetes/pki/sa.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /tmp/kubeadm.yaml
    owner: root:root
    permissions: '0640'
    content: |
        ---
        apiServer:
          extraArgs:
            cloud-provider: external
        apiVersion: kubeadm.k8s.io/v1beta1
        clusterName: cluster-dev
        controlPlaneEndpoint: 1.2.3.4:6443
        controllerManager:
          extraArgs:
            cloud-provider: external
        dns: {}
        etcd: {}
        kind: ClusterConfiguration
        kubernetesVersion: v1.18.1
        networking:
          dnsDomain: cluster.local
          podSubnet: 192.168.0.0/16
          serviceSubnet: 172.16.0.0/12
        scheduler: {}
        ---
        apiVersion: kubeadm.k8s.io/v1beta1
        kind: InitConfiguration
        localAPIEndpoint:
          advertiseAddress: ""
          bindPort: 0
        nodeRegistration:
          kubeletExtraArgs:
            cloud-provider: external
runcmd:
  - 'kubeadm init --config /tmp/kubeadm.yaml '
`

var userdataChangesControlPlane = `## template: jinja
#cloud-config

write_files:
  - path: /etc/kubernetes/pki/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/etcd/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/etcd/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/front-proxy-ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/front-proxy-ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/sa.pub
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN PUBLIC KEY-----
        -----END PUBLIC KEY-----
  - path: /etc/kubernetes/pki/sa.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /tmp/kubeadm.yaml
    owner: root:root
    permissions: '0640'
    content: |
        ---
        apiServer:
          extraArgs:
            cloud-provider: external
        apiVersion: kubeadm.k8s.io/v1beta1
        clusterName: cluster-dev
        controllerManager:
          extraArgs:
            cloud-provider: external
        dns: {}
        etcd: {}
        kind: ClusterConfiguration
        kubernetesVersion: v1.18.1
        networking:
          dnsDomain: cluster.local
          podSubnet: 192.168.0.0/16
          serviceSubnet: 172.16.0.0/12
        scheduler: {}
        ---
        apiVersion: kubeadm.k8s.io/v1beta1
        kind: InitConfiguration
        localAPIEndpoint:
          advertiseAddress: ""
          bindPort: 0
        nodeRegistration:
          kubeletExtraArgs:
            cloud-provider: external
        ---
        apiVersion: kubelet.config.k8s.io/v1beta1
        kind: KubeletConfiguration
        rotateCertificates: true
        serverTLSBootstrap: true
runcmd:
  - 'kubeadm init --config /tmp/kubeadm.yaml '
`

var userdataSetOrAddFiles = `## template: jinja
#cloud-config

write_files:
  - path: /etc/kubernetes/pki/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/etcd/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/etcd/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/front-proxy-ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/front-proxy-ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/sa.pub
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN PUBLIC KEY-----
        -----END PUBLIC KEY-----
  - path: /etc/kubernetes/pki/sa.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /tmp/kubeadm.yaml
    owner: noroot:noroot
    permissions: '0444'
    content: |
        D
        E
        F
        G
  - content: |
        A
        B
        C
        D
    owner: root:root
    path: /tmp/new
    permissions: "0440"
runcmd:
  - 'kubeadm init --config /tmp/kubeadm.yaml '
`

var userdataControlPlaneJoin = `## template: jinja
#cloud-config

write_files:
  - path: /etc/kubernetes/pki/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/etcd/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/etcd/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/front-proxy-ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/front-proxy-ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/sa.pub
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN PUBLIC KEY-----
        -----END PUBLIC KEY-----
  - path: /etc/kubernetes/pki/sa.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /tmp/kubeadm-join-config.yaml
    owner: root:root
    permissions: '0640'
    content: |
        apiVersion: kubeadm.k8s.io/v1beta1
        controlPlane:
          localAPIEndpoint:
            advertiseAddress: ""
            bindPort: 0
        discovery:
          bootstrapToken:
            apiServerEndpoint: 1.2.3.4:6443
            caCertHashes:
            - sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
            token: xxxxxx.yyyyyyyyyyyyyyyy
            unsafeSkipCAVerification: false
        kind: JoinConfiguration
        nodeRegistration: {}
runcmd:
  - kubeadm join --config /tmp/kubeadm-join-config.yaml
`

var userdataControlPlaneJoinChanged = `## template: jinja
#cloud-config

write_files:
  - path: /etc/kubernetes/pki/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/etcd/ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/etcd/ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/front-proxy-ca.crt
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN CERTIFICATE-----
        -----END CERTIFICATE-----
  - path: /etc/kubernetes/pki/front-proxy-ca.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /etc/kubernetes/pki/sa.pub
    owner: root:root
    permissions: '0640'
    content: |
        -----BEGIN PUBLIC KEY-----
        -----END PUBLIC KEY-----
  - path: /etc/kubernetes/pki/sa.key
    owner: root:root
    permissions: '0600'
    content: |
        -----BEGIN RSA PRIVATE KEY-----
        -----END RSA PRIVATE KEY-----
  - path: /tmp/kubeadm-join-config.yaml
    owner: root:root
    permissions: '0640'
    content: |
        ---
        apiVersion: kubeadm.k8s.io/v1beta1
        controlPlane:
          localAPIEndpoint:
            advertiseAddress: ""
            bindPort: 0
        discovery:
          bootstrapToken:
            apiServerEndpoint: 1.2.3.4:6443
            caCertHashes:
            - sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
            token: xxxxxx.yyyyyyyyyyyyyyyy
            unsafeSkipCAVerification: false
        kind: JoinConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            test: value
runcmd:
  - kubeadm join --config /tmp/kubeadm-join-config.yaml
`

func TestUserData_Basic_Init(t *testing.T) {
	r := strings.NewReader(userdataControlPlaneInit)
	u, err := NewFromReader(r)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	b := bytes.NewBuffer(nil)
	if err := u.WriteYAML(b); err != nil {
		t.Errorf("unexpected error: %w", err)
	}
	assert.Equal(t, userdataControlPlaneInit, b.String(), "they should be equal")
}

// This tests a rewrite if the command to skip kube-proxy
func TestUserData_Basic_SkipKubeProxy(t *testing.T) {
	r := strings.NewReader(userdataControlPlaneInit)
	u, err := NewFromReader(r)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	if err := u.SkipKubeProxy(); err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	b := bytes.NewBuffer(nil)
	if err := u.WriteYAML(b); err != nil {
		t.Errorf("unexpected error: %w", err)
	}
	data := strings.Split(userdataControlPlaneInit, "\n")
	data = data[:len(data)-2]
	data = append(data, "  - 'kubeadm init --config /tmp/kubeadm.yaml --skip-phases=addon/kube-proxy'")
	assert.Equal(t, strings.Join(data, "\n")+"\n", b.String(), "they should be equal")
}

// This tests roughly what is necessary for updating the control plane kubeadm
// init config
func TestUserData_UpdateKubeadmConfig_Init(t *testing.T) {
	r := strings.NewReader(userdataControlPlaneInit)
	u, err := NewFromReader(r)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	k, err := u.GetKubeadmConfig()
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	k.ClusterConfiguration.ControlPlaneEndpoint = ""

	err = u.SetKubeadmConfig(k)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	b := bytes.NewBuffer(nil)
	if err := u.WriteYAML(b); err != nil {
		t.Errorf("unexpected error: %w", err)
	}
	assert.Equal(t, userdataChangesControlPlane, b.String(), "they should be equal")
}

func TestUserData_Basic_Join(t *testing.T) {
	r := strings.NewReader(userdataControlPlaneJoin)
	u, err := NewFromReader(r)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	b := bytes.NewBuffer(nil)
	if err := u.WriteYAML(b); err != nil {
		t.Errorf("unexpected error: %w", err)
	}
	assert.Equal(t, userdataControlPlaneJoin, b.String(), "they should be equal")
}

func TestUserData_UpdateKubeadmConfig_Join(t *testing.T) {
	r := strings.NewReader(userdataControlPlaneJoin)
	u, err := NewFromReader(r)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	k, err := u.GetKubeadmConfig()
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	err = u.SetKubeadmConfig(k)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	b := bytes.NewBuffer(nil)
	if err := u.WriteYAML(b); err != nil {
		t.Errorf("unexpected error: %w", err)
	}
	assert.Equal(t, userdataControlPlaneJoinChanged, b.String(), "they should be equal")
}

// This tests updating and adding a file
func TestUserData_SetOrUpdateFile(t *testing.T) {
	r := strings.NewReader(userdataControlPlaneInit)
	u, err := NewFromReader(r)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	err = u.SetOrUpdateFile(bootstrapv1.File{
		Path:        "/tmp/new",
		Content:     "A\nB\nC\nD\n",
		Owner:       "root:root",
		Permissions: "0440",
	})
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	err = u.SetOrUpdateFile(bootstrapv1.File{
		Path:        "/tmp/kubeadm.yaml",
		Content:     "D\nE\nF\nG\n",
		Owner:       "noroot:noroot",
		Permissions: "0444",
	})
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	b := bytes.NewBuffer(nil)
	if err := u.WriteYAML(b); err != nil {
		t.Errorf("unexpected error: %w", err)
	}
	assert.Equal(t, userdataSetOrAddFiles, b.String(), "they should be equal")
}
*/
