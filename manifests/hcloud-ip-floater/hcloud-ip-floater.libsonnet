local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

std.prune(upstream {
  _config:: {
    // token secret reference
    tokenRef:: null,
  },

  // remove namespace
  namespace: null,

  'hcloud-ip-floater'+: {
    local this = self,


    // make inner container accessible
    container:: super.deployment.spec.template.spec.containers[0] {
      envFrom: null,
      env: [{
        name: '',
        secret: $._config.tokenRef,
      }],
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
})
