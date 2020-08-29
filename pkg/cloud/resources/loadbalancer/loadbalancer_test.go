package loadbalancer_test

import (
	"context"
	"flag"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hetznercloud/hcloud-go/hcloud"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	loadbalancer "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/loadbalancer"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope"
	mock_scope "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope/mock"
)

func TestLoadBalancers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LoadBalancer Suite")
}

type hcloudCreateOptsMatcher struct {
}

var _ = Describe("LoadBalancers", func() {

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
						Name: "cluster-dev",
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

	Context("No load balancers desired, no load balancer existing", func() {
		It("should reconcile", func() {
			scp := newClusterScope()
			hclient := hcloud.NewClient()
			loadBalancerType, _, _ := hclient.LoadBalancerType.GetByName(context.Background(), "lb11")
			alg := hcloud.LoadBalancerAlgorithm{Type: hcloud.LoadBalancerAlgorithmTypeRoundRobin}
			mockClient.EXPECT().ListLoadBalancers(gomock.Any(), gomock.Any()).Return(
				[]*hcloud.LoadBalancer{{
					Name:             "test_lb",
					LoadBalancerType: loadBalancerType,
					Algorithm:        alg,
					Labels:           map[string]string{"unimportant": "for-sure"},
				}},
				nil,
			)
			svc := loadbalancer.NewService(scp)
			Expect(svc.Reconcile(context.TODO())).NotTo(HaveOccurred())
			Expect(len(scp.HcloudCluster.Status.ControlPlaneLoadBalancers)).To(Equal(0))
		})
	})

	Context("No loadbalancers desired, existing Loadbalancer but none recorded in status", func() {
		It("should reconcile and delete Loadbalancers", func() {
			scp := newClusterScope()
			hclient := hcloud.NewClient()
			loadBalancerType, _, _ := hclient.LoadBalancerType.GetByName(context.Background(), "lb11")
			alg := hcloud.LoadBalancerAlgorithm{Type: hcloud.LoadBalancerAlgorithmTypeRoundRobin}
			gomock.InOrder(
				mockClient.EXPECT().ListLoadBalancers(gomock.Any(), gomock.Any()).Return(
					[]*hcloud.LoadBalancer{
						{
							Name:             "test_lb1",
							LoadBalancerType: loadBalancerType,
							Algorithm:        alg,
							Labels:           map[string]string{"unimportant": "for-sure"},
						},
						{
							Name:             "test_lb2",
							LoadBalancerType: loadBalancerType,
							Algorithm:        alg,
							Labels:           map[string]string{"unimportant": "for-sure"},
						},
					},
					nil,
				),
				mockClient.EXPECT().ListLoadBalancers(gomock.Any(), gomock.Any()).Return(
					[]*hcloud.LoadBalancer{
						{
							Name:             "test_lb3",
							LoadBalancerType: loadBalancerType,
							Algorithm:        alg,
							Labels:           map[string]string{"unimportant": "for-sure"},
						},
					},
					nil,
				),
			)
			mockClient.EXPECT().DeleteLoadBalancer(gomock.Any(), gomock.Eq(&hcloud.LoadBalancer{Name: "test_lb2"})).Return(nil, nil)
			svc := loadbalancer.NewService(scp)
			Expect(svc.Reconcile(context.TODO())).NotTo(HaveOccurred())
			//Expect(len(scp.HcloudCluster.Status.ControlPlaneLoadBalancers)).To(Equal(1))
		})
	})
})
