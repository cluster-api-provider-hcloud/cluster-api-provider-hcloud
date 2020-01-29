local calico = import 'calico/calico.libsonnet';
local hcloudCloudControllerManager = import 'hcloud-cloud-controller-manager/hcloud-cloud-controller-manager.libsonnet';
local hcloudIPFloater = import 'hcloud-ip-floater/hcloud-ip-floater.libsonnet';

{
  _config:: {
    token: 'xx',
    podCIDR: '192.168.0.0/16',
  },
  manifests: {
    calico: calico,
    hcloudCloudControllerManager: hcloudCloudControllerManager,
    hcloudIPFloater: hcloudIPFloater,
  },
}
