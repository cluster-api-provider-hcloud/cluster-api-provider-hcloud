local calico = import 'calico/calico.libsonnet';
local hcloudCloudControllerManager = import 'hcloud-cloud-controller-manager/hcloud-cloud-controller-manager.libsonnet';
local hcloudCSI = import 'hcloud-csi/hcloud-csi.libsonnet';
local hcloudMetalLBFloater = import 'hcloud-metallb-floater/hcloud-metallb-floater.libsonnet';
local metalLB = import 'metallb/metallb.libsonnet';
local metricsServer = import 'metrics-server/metrics-server.libsonnet';

local newControlPlaneService(pos, ip) = {
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: 'kube-apiserver-%d' % pos,
    namespace: 'kube-system',
  },
  spec: {
    selector: {
      component: 'kube-apiserver',
      tier: 'control-plane',
    },
    ports: [
      {
        protocol: 'TCP',
        port: 6443,
        targetPort: 6443,
      },
    ],
    type: 'LoadBalancer',
    loadBalancerIP: ip,
    externalTrafficPolicy: 'Local',
  },
};


{
  _config:: {
    hcloudNetworkRef: {
      valueFrom: {
        secretKeyRef: {
          key: 'network',
          name: 'hcloud',
        },
      },
    },
    hcloudTokenRef: {
      valueFrom: {
        secretKeyRef: {
          key: 'token',
          name: 'hcloud',
        },
      },
    },
    podsCIDRBlock: '192.168.0.0/16',
    hcloudToken: 'xx',
    hcloudNetwork: 'yy',
    hcloudFloatingIPs: ['1.1.1.1', '2.2.2.2'],
  },

  hcloudSecret: {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'hcloud',
      namespace: 'kube-system',
    },
    type: 'Opaque',
    data: {
      network: std.base64($._config.hcloudNetwork),
      token: std.base64($._config.hcloudToken),
    },
  },

  controlPlaneServices: std.mapWithIndex(newControlPlaneService, $._config.hcloudFloatingIPs),

  manifests:
    calico +
    hcloudCloudControllerManager +
    hcloudCSI +
    metricsServer +
    hcloudMetalLBFloater +
    metalLB +
    {
      _config+:: $._config,
    },
}
