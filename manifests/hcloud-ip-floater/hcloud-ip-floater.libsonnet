local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

std.prune(upstream {
  // remove namespace
  namespace: null,
}) {
  _config+:: {
    // token secret reference
    hcloudTokenRef:: null,
  },


  'hcloud-ip-floater'+: {
    local this = self,


    // make inner container accessible
    container:: super.deployment.spec.template.spec.containers[0] {
      envFrom: null,
      env+:
        if $._config.hcloudTokenRef != null then [{
          name: 'HCLOUD_IP_FLOATER_HCLOUD_TOKEN',
        } + $._config.hcloudTokenRef]
        else [],
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
