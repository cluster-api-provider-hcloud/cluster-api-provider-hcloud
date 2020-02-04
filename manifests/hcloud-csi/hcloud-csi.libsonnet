local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

upstream {
  _config+:: {
    hcloudTokenRef: null,
  },

  'hcloud-csi-controller'+:
    if $._config.hcloudTokenRef != null then
      {
        statefulSet: utils.recursiveEnvReplace(super.statefulSet, { name: 'HCLOUD_TOKEN' } + $._config.hcloudTokenRef),
      }
    else {},

  'hcloud-csi-node'+:
    if $._config.hcloudTokenRef != null then
      {
        daemonSet: utils.recursiveEnvReplace(super.daemonSet, { name: 'HCLOUD_TOKEN' } + $._config.hcloudTokenRef),
      }
    else {},
}
