package server

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var userDataExample = []byte(`## template: jinja
#cloud-config

write_files:
  - content: |
        ---
        apiServer:
          extraArgs:
            cloud-provider: external
        apiVersion: kubeadm.k8s.io/v1beta1
        certificatesDir: /etc/kubernetes/pki
        clusterName: christian-dev
        controlPlaneEndpoint: ""
        controllerManager:
          extraArgs:
            cloud-provider: external
        dns:
          type: ""
        etcd: {}
        imageRepository: ""
        kind: ClusterConfiguration
        kubernetesVersion: v1.16.6
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
    owner: root:root
    path: /tmp/kubeadm.yaml
    permissions: "0640"
runcmd:
  - kubeadm init --config /tmp/kubeadm.yaml
`)

var userDataWithMountExample = []byte(`## template: jinja
#cloud-config

write_files:
  - content: |
        ---
        apiServer:
          extraArgs:
            cloud-provider: external
        apiVersion: kubeadm.k8s.io/v1beta1
        certificatesDir: /etc/kubernetes/pki
        clusterName: christian-dev
        controlPlaneEndpoint: ""
        controllerManager:
          extraArgs:
            cloud-provider: external
        dns:
          type: ""
        etcd: {}
        imageRepository: ""
        kind: ClusterConfiguration
        kubernetesVersion: v1.16.6
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
    owner: root:root
    path: /tmp/kubeadm.yaml
    permissions: "0640"
  - path: /etc/systemd/system/var-lib-data.mount
    content: |
        [Mount]
        What=/dev/disk/by-id/scsi-0HC_Volume_123
        Where=/var/lib/data
        Type=ext4
        Options=discard,defaults

        [Install]
        WantedBy=local-fs.target
    permissions: "0644"
    owner: root:root
runcmd:
  - kubeadm init --config /tmp/kubeadm.yaml
`)

var userDataWithAppenedKubeadm = []byte(`## template: jinja
#cloud-config

write_files:
  - content: |
        ---
        apiServer:
          extraArgs:
            cloud-provider: external
        apiVersion: kubeadm.k8s.io/v1beta1
        certificatesDir: /etc/kubernetes/pki
        clusterName: christian-dev
        controlPlaneEndpoint: ""
        controllerManager:
          extraArgs:
            cloud-provider: external
        dns:
          type: ""
        etcd: {}
        imageRepository: ""
        kind: ClusterConfiguration
        kubernetesVersion: v1.16.6
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
    owner: root:root
    path: /tmp/kubeadm.yaml
    permissions: "0640"
runcmd:
  - kubeadm init --config /tmp/kubeadm.yaml
`)

func TestUserData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UserData Suite")
}

var _ = Describe("UserData", func() {
	var (
		s *Service
	)

	BeforeEach(func() {
		s = &Service{}
	})
	Context("No changes done", func() {
		It("should match", func() {
			u, err := s.parseUserData(userDataExample)
			Expect(err).NotTo(HaveOccurred())
			output, err := u.output()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal(string(userDataExample)))
		})
	})
	Context("Add volume mount", func() {
		It("should have an extra systemd unit", func() {
			u, err := s.parseUserData(userDataExample)
			Expect(err).NotTo(HaveOccurred())
			err = u.addVolumeMount(123, "/var/lib/data")
			Expect(err).NotTo(HaveOccurred())
			output, err := u.output()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal(string(userDataWithMountExample)))
		})
	})
	Context("Append to kubeadm config", func() {
		It("should have an extra systemd unit", func() {
			u, err := s.parseUserData(userDataExample)
			Expect(err).NotTo(HaveOccurred())
			err = u.appendKubeadmConfig(`---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
rotateCertificates: true
serverTLSBootstrap: true
`)
			Expect(err).NotTo(HaveOccurred())
			output, err := u.output()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal(string(userDataWithAppenedKubeadm)))
		})
	})
})
