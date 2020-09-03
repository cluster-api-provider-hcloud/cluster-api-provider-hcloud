local config = import 'config.jsonnet';
local utils = import 'utils.libsonnet';

local myConfig = {
  hcloudToken: std.extVar('hcloud-token'),
  hcloudNetwork: std.extVar('hcloud-network'),
  hcloudLoadBalancerIPv4s: std.split(std.extVar('hcloud-loadbalancer'), ','),
  podsCIDRBlock: std.extVar('pod-cidr-block'),
  local networkConfig = std.parseJson(std.extVar('network')),
  network+: {
    [k]+: networkConfig[k]
    for k in std.objectFields(networkConfig)
  },
};

local addLabelIfNotExists(key, value) =
  function(obj) if
    std.objectHas(obj, 'metadata') &&
    std.objectHas(obj.metadata, 'labels') &&
    std.objectHas(obj.metadata.labels, key) then
    obj else
    obj {
      metadata+: {
        labels+: {
          [key]: value,
        },
      },
    };

utils.mapPerRessource(
  config.new(myConfig),
  addLabelIfNotExists(
    'cluster-api-provider-hcloud.capihc.com/manifests',
    'Reconcile',
  ),
)
