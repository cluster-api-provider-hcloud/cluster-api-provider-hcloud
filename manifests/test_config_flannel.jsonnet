// TODO: Update test
/*
local config = import 'config.jsonnet';

local defaults = {
  hcloudToken: 'my-token',
  hcloudNetwork: 'my-network',
  network+: {
    flannel+: {
      backend: 'wireguard',
    },
  },
};

local flannelManifests(x) = std.filter(
  function(y) std.startsWith(y, 'cilium') || std.startsWith(y, 'flannel'),
  std.objectFields(x),
);

{
  testDefaults:: config.new(defaults),

  keysDefaults: flannelManifests($.testDefaults),
}
*/
{}