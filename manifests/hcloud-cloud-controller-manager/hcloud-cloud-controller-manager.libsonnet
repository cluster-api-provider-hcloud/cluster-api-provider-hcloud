local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

upstream {

  'hcloud-cloud-controller-manager'+: {
    deployment+: {
      spec+: {
        template+: {
          spec+: {
            tolerations+: [{
              key: 'node.kubernetes.io/unreachable',
              operator: 'Exists',
              effect: 'NoExecute',
              tolerationSeconds: 60,
            }],
          },
        },
      },
    },
  },

}
