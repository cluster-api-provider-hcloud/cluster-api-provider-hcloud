local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

upstream {
  'metrics-server'+: {
    local this = self,

    container:: super.deployment.spec.template.spec.containers[0] {
      args+: [
        '--metric-resolution=30s',
        '--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname',
      ],
    },

    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [this.container],
          },
        },
      },
    },
  },
}
