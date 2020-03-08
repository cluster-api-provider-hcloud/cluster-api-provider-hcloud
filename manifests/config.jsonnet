local calico = import 'calico/calico.libsonnet';
local hcloudCloudControllerManager = import 'hcloud-cloud-controller-manager/hcloud-cloud-controller-manager.libsonnet';
local hcloudCSI = import 'hcloud-csi/hcloud-csi.libsonnet';
local metricsServer = import 'metrics-server/metrics-server.libsonnet';
local apiServerKeepalived = (import 'kube-keepalived-vip/kube-keepalived-vip.libsonnet') {
  _config+:: {
    name: 'kube-apiserver-keepalived-vip',
    backendService: 'default/kubernetes',
  },
  daemonSet+: {
    spec+: {
      template+: {
        spec+: {
          nodeSelector+: {
            'node-role.kubernetes.io/master': '',
          },
          tolerations+: [
            {
              key: 'CriticalAddonsOnly',
              operator: 'Exists',
            },
            {
              operator: 'Exists',
            },
          ],
        },
      },
    },
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
    hcloudFloatingIPs: ['1.1.1.1'],
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

  manifests:
    calico +
    hcloudCloudControllerManager +
    hcloudCSI +
    metricsServer +
    apiServerKeepalived +
    {
      _config+:: $._config,
    },
}
