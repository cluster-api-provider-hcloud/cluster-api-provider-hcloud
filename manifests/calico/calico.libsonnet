local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

upstream {
  _config:: {
    podsCIDRBlock: '192.168.0.0/17',
  },


  'calico-node'+: {
    local this = self,

    container:: super.daemonSet.spec.template.spec.containers[0] {
      env: utils.mergeEnv(super.env, [
        // Set pod cidr
        { name: 'CALICO_IPV4POOL_CIDR', value: $._config.podsCIDRBlock },
        // Disable XDP (prevented by SELinux)
        { name: 'FELIX_XDPENABLED', value: 'false' },
      ]),
    },

    daemonSet+: {
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
