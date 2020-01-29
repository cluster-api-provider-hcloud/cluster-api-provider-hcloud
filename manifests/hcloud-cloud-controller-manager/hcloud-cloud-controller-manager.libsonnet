local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

std.prune(upstream {
  metadata:: {
    namespace: 'kube-system',
  },
})
