local config = import 'config.jsonnet';
local utils = import 'utils.libsonnet';

local myConfig = {
  hcloudToken: std.extVar('hcloud-token'),
  robotUserName: std.extVar('robot-username'),
  robotPassword: std.extVar('robot-password'),
  hcloudNetwork: std.extVar('hcloud-network'),
  kubeAPIServerIPv4: std.extVar('kube-apiserver-ip'),
  kubeAPIServerDomain: std.extVar('kube-apiserver-domain'),
  port: std.parseInt(std.extVar('port')),
  caCrt: std.extVar('ca-crt'),
  caKey: std.extVar('ca-key'),
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
