package floatingip_test

import (
	"context"
	"flag"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hetznercloud/hcloud-go/hcloud"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"

	infrav1 "github.com/simonswine/cluster-api-provider-hetzner/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/resources/floatingip"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/scope"
	mock_scope "github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/scope/mock"
)

func TestFloatingIPs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FloatingIP Suite")
}

type hcloudCreateOptsMatcher struct {
}

var _ = Describe("FloatingIPs", func() {

	BeforeSuite(func() {
		klog.InitFlags(nil)
		if testing.Verbose() {
			flag.Set("v", "5")
		}
	})

	var (
		mockCtrl      *gomock.Controller
		mockClient    *mock_scope.MockHetznerClient
		clientFactory = func(ctx context.Context) (scope.HetznerClient, error) {
			return mockClient, nil
		}
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock_scope.NewMockHetznerClient(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("No floating IPs desired, no floating IP existing", func() {
		It("should reconcile", func() {
			scp, err := scope.NewClusterScope(scope.ClusterScopeParams{
				Cluster:              &clusterv1.Cluster{},
				HetznerCluster:       &infrav1.HetznerCluster{},
				HetznerClientFactory: clientFactory,
			})
			mockClient.EXPECT().ListFloatingIPs(gomock.Any(), gomock.Any()).Return(
				[]*hcloud.FloatingIP{&hcloud.FloatingIP{
					ID:     123,
					Type:   hcloud.FloatingIPTypeIPv4,
					IP:     net.IPv4(0x01, 0x01, 0x01, 0x01),
					Labels: map[string]string{"unimportant": "for-sure"},
				}},
				nil,
			)

			Expect(err).NotTo(HaveOccurred())
			svc := floatingip.NewService(scp)
			Expect(svc.Reconcile(context.TODO())).NotTo(HaveOccurred())
			Expect(len(scp.HetznerCluster.Status.ControlPlaneFloatingIPs)).To(Equal(0))
		})
	})

	Context("No floating IPs desired, existing IPs but none recorded in status", func() {
		It("should reconcile and delete floating IP", func() {
			scp, err := scope.NewClusterScope(scope.ClusterScopeParams{
				Cluster: &clusterv1.Cluster{},
				HetznerCluster: &infrav1.HetznerCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "christian-dev",
					},
				},
				HetznerClientFactory: clientFactory,
			})
			scp.HetznerCluster.Name = "christian-dev"
			clusterTagKey := infrav1.ClusterTagKey(scp.HetznerCluster.Name)
			gomock.InOrder(
				mockClient.EXPECT().ListFloatingIPs(gomock.Any(), gomock.Any()).Return(
					[]*hcloud.FloatingIP{
						&hcloud.FloatingIP{
							ID:     123,
							Type:   hcloud.FloatingIPTypeIPv4,
							IP:     net.IPv4(0x01, 0x01, 0x01, 0x01),
							Labels: map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)},
						},
						&hcloud.FloatingIP{
							ID:     124,
							Type:   hcloud.FloatingIPTypeIPv4,
							IP:     net.IPv4(0x01, 0x00, 0x00, 0x01),
							Labels: map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleShared)},
						},
					},
					nil,
				),
				mockClient.EXPECT().ListFloatingIPs(gomock.Any(), gomock.Any()).Return(
					[]*hcloud.FloatingIP{
						&hcloud.FloatingIP{
							ID:     124,
							Type:   hcloud.FloatingIPTypeIPv4,
							IP:     net.IPv4(0x01, 0x00, 0x00, 0x01),
							Labels: map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleShared)},
						},
					},
					nil,
				),
			)
			mockClient.EXPECT().DeleteFloatingIP(gomock.Any(), gomock.Eq(&hcloud.FloatingIP{ID: 123})).Return(nil, nil)
			Expect(err).NotTo(HaveOccurred())
			svc := floatingip.NewService(scp)
			Expect(svc.Reconcile(context.TODO())).NotTo(HaveOccurred())
			//Expect(len(scp.HetznerCluster.Status.ControlPlaneFloatingIPs)).To(Equal(1))
		})
	})

	Context("A v4 and v6 floating IP desired, none existing", func() {
		It("should reconcile and create floating IPs", func() {
			scp, err := scope.NewClusterScope(scope.ClusterScopeParams{
				Cluster: &clusterv1.Cluster{},
				HetznerCluster: &infrav1.HetznerCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "christian-dev",
					},
					Spec: infrav1.HetznerClusterSpec{
						ControlPlaneFloatingIPs: []infrav1.HetznerFloatingIPSpec{
							{Type: infrav1.HetznerFloatingIPTypeIPv4},
							{Type: infrav1.HetznerFloatingIPTypeIPv6},
						},
					},
					Status: infrav1.HetznerClusterStatus{
						Location: "myhome",
					},
				},
				HetznerClientFactory: clientFactory,
			})
			scp.HetznerCluster.Spec.ControlPlaneFloatingIPs = []infrav1.HetznerFloatingIPSpec{
				{Type: infrav1.HetznerFloatingIPTypeIPv4},
				{Type: infrav1.HetznerFloatingIPTypeIPv6},
			}
			clusterTagKey := infrav1.ClusterTagKey(scp.HetznerCluster.Name)
			_, deadBeefNetwork, err := net.ParseCIDR("2001:dead:beef::/64")
			Expect(err).NotTo(HaveOccurred())
			fipIPV4 := &hcloud.FloatingIP{
				ID:     124,
				Type:   hcloud.FloatingIPTypeIPv4,
				IP:     net.IPv4(0x01, 0x00, 0x00, 0x01),
				Labels: map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)},
			}
			fipIPV6 := &hcloud.FloatingIP{
				ID:      125,
				Type:    hcloud.FloatingIPTypeIPv6,
				Network: deadBeefNetwork,
				Labels:  map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)},
			}

			gomock.InOrder(
				mockClient.EXPECT().ListFloatingIPs(gomock.Any(), gomock.Any()).Return(
					[]*hcloud.FloatingIP{},
					nil,
				),
				mockClient.EXPECT().ListFloatingIPs(gomock.Any(), gomock.Any()).Return(
					[]*hcloud.FloatingIP{
						fipIPV4,
						fipIPV6,
					},
					nil,
				),
			)

			gomock.InOrder(
				mockClient.EXPECT().CreateFloatingIP(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opts hcloud.FloatingIPCreateOpts) (hcloud.FloatingIPCreateResult, *hcloud.Response, error) {
					Expect(opts.HomeLocation.Name).To(Equal(string(scp.HetznerCluster.Status.Location)))
					Expect(opts.Type).To(Equal(hcloud.FloatingIPTypeIPv4))
					return hcloud.FloatingIPCreateResult{
						FloatingIP: fipIPV4,
					}, nil, nil
				}),
				mockClient.EXPECT().CreateFloatingIP(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opts hcloud.FloatingIPCreateOpts) (hcloud.FloatingIPCreateResult, *hcloud.Response, error) {
					Expect(opts.HomeLocation.Name).To(Equal(string(scp.HetznerCluster.Status.Location)))
					Expect(opts.Type).To(Equal(hcloud.FloatingIPTypeIPv6))
					return hcloud.FloatingIPCreateResult{
						FloatingIP: fipIPV6,
					}, nil, nil
				}),
			)

			Expect(err).NotTo(HaveOccurred())
			svc := floatingip.NewService(scp)
			Expect(svc.Reconcile(context.TODO())).NotTo(HaveOccurred())
			Expect(len(scp.HetznerCluster.Status.ControlPlaneFloatingIPs)).To(Equal(2))

			mockClient.EXPECT().ListFloatingIPs(gomock.Any(), gomock.Any()).Return(
				[]*hcloud.FloatingIP{
					fipIPV4,
					fipIPV6,
				},
				nil,
			)
			Expect(svc.Reconcile(context.TODO())).NotTo(HaveOccurred())
		})
	})
})
