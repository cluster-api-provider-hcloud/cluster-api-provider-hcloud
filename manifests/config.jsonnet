local calico = import 'calico/calico.libsonnet';
local hcloudCloudControllerManager = import 'hcloud-cloud-controller-manager/hcloud-cloud-controller-manager.libsonnet';
local hcloudCSI = import 'hcloud-csi/hcloud-csi.libsonnet';
local hcloudIPFloater = import 'hcloud-ip-floater/hcloud-ip-floater.libsonnet';
local metricsServer = import 'metrics-server/metrics-server.libsonnet';

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
    hcloudIPFloater +
    metricsServer +
    {
      _config+:: $._config,
    },
}
