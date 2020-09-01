local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

upstream
{
  _config+:: {
    // token secret reference
    hcloudTokenRef+: {},

    images+: {
      'hcloud-metallb-floater': 'docker.io/simonswine/hcloud-metallb-floater@sha256:4e0de7d4b20a8b052db77ef3833c862fbab40159ae58f1b9e73e4b9a8a294e9d',  // v0.1.0
    },
  },


  'hcloud-metallb-floater-controller'+: {
    local this = self,


    // make inner container accessible
    container:: super.deployment.spec.template.spec.containers[0] {
      image: $._config.images['hcloud-metallb-floater'],
      imagePullPolicy: 'IfNotPresent',
      env: if $._config.hcloudTokenRef != null then
        utils.mergeEnv(super.env, [{
          name: 'HCLOUD_TOKEN',
        } + $._config.hcloudTokenRef])
      else super.env,
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
