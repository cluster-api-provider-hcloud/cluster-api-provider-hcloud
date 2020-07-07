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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/floatingip"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope"
	mock_scope "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope/mock"
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
		mockClient    *mock_scope.MockHcloudClient
		mockPacker    *mock_scope.MockPacker
		mockManifests *mock_scope.MockManifests
		clientFactory = func(ctx context.Context) (scope.HcloudClient, error) {
			return mockClient, nil
		}
		newClusterScope func() *scope.ClusterScope
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock_scope.NewMockHcloudClient(mockCtrl)
		mockPacker = mock_scope.NewMockPacker(mockCtrl)
		mockManifests = mock_scope.NewMockManifests(mockCtrl)
		newClusterScope = func() *scope.ClusterScope {
			scp, err := scope.NewClusterScope(scope.ClusterScopeParams{
				Cluster: &clusterv1.Cluster{},
				HcloudCluster: &infrav1.HcloudCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "christian-dev",
					},
				},
				HcloudClientFactory: clientFactory,
				Packer:              mockPacker,
				Manifests:           mockManifests,
			})
			Expect(err).NotTo(HaveOccurred())
			return scp
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("No floating IPs desired, no floating IP existing", func() {
		It("should reconcile", func() {
			scp := newClusterScope()
			mockClient.EXPECT().ListFloatingIPs(gomock.Any(), gomock.Any()).Return(
				[]*hcloud.FloatingIP{{
					ID:     123,
					Type:   hcloud.FloatingIPTypeIPv4,
					IP:     net.IPv4(0x01, 0x01, 0x01, 0x01),
					Labels: map[string]string{"unimportant": "for-sure"},
				}},
				nil,
			)
			svc := floatingip.NewService(scp)
			Expect(svc.Reconcile(context.TODO())).NotTo(HaveOccurred())
			Expect(len(scp.HcloudCluster.Status.ControlPlaneFloatingIPs)).To(Equal(0))
		})
	})

	Context("No floating IPs desired, existing IPs but none recorded in status", func() {
		It("should reconcile and delete floating IP", func() {
			scp := newClusterScope()
			clusterTagKey := infrav1.ClusterTagKey(scp.HcloudCluster.Name)
			gomock.InOrder(
				mockClient.EXPECT().ListFloatingIPs(gomock.Any(), gomock.Any()).Return(
					[]*hcloud.FloatingIP{
						{
							ID:     123,
							Type:   hcloud.FloatingIPTypeIPv4,
							IP:     net.IPv4(0x01, 0x01, 0x01, 0x01),
							Labels: map[string]string{clusterTagKey: string(infrav1.ResourceLifecycleOwned)},
						},
						{
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
						{
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
			svc := floatingip.NewService(scp)
			Expect(svc.Reconcile(context.TODO())).NotTo(HaveOccurred())
			//Expect(len(scp.HcloudCluster.Status.ControlPlaneFloatingIPs)).To(Equal(1))
		})
	})

	Context("A v4 and v6 floating IP desired, none existing", func() {
		It("should reconcile and create floating IPs", func() {
			scp := newClusterScope()
			scp.HcloudCluster.Spec.ControlPlaneFloatingIPs = []infrav1.HcloudFloatingIPSpec{
				{Type: infrav1.HcloudFloatingIPTypeIPv4},
				{Type: infrav1.HcloudFloatingIPTypeIPv6},
			}
			scp.HcloudCluster.Status = infrav1.HcloudClusterStatus{
				Locations: []infrav1.HcloudLocation{infrav1.HcloudLocation("myhome")},
			}
			clusterTagKey := infrav1.ClusterTagKey(scp.HcloudCluster.Name)
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
					Expect(opts.HomeLocation.Name).To(Equal(string(scp.HcloudCluster.Status.Locations[0])))
					Expect(opts.Type).To(Equal(hcloud.FloatingIPTypeIPv4))
					return hcloud.FloatingIPCreateResult{
						FloatingIP: fipIPV4,
					}, nil, nil
				}),
				mockClient.EXPECT().CreateFloatingIP(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, opts hcloud.FloatingIPCreateOpts) (hcloud.FloatingIPCreateResult, *hcloud.Response, error) {
					Expect(opts.HomeLocation.Name).To(Equal(string(scp.HcloudCluster.Status.Locations[0])))
					Expect(opts.Type).To(Equal(hcloud.FloatingIPTypeIPv6))
					return hcloud.FloatingIPCreateResult{
						FloatingIP: fipIPV6,
					}, nil, nil
				}),
			)

			Expect(err).NotTo(HaveOccurred())
			svc := floatingip.NewService(scp)
			Expect(svc.Reconcile(context.TODO())).NotTo(HaveOccurred())
			Expect(len(scp.HcloudCluster.Status.ControlPlaneFloatingIPs)).To(Equal(2))

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
