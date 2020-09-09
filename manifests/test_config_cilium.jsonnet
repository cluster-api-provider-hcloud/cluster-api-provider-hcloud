// TODO: Update this test
/*
local config = import 'config.jsonnet';

local defaults = {
  hcloudToken: 'my-token',
  hcloudNetwork: 'my-network',
  network: {
    cilium: {
    },
  },
};

local ciliumManifests(x) = std.filter(
  function(y) std.startsWith(y, 'cilium') || std.startsWith(y, 'flannel'),
  std.objectFields(x),
);

{
  testDefaults::
    config.new(defaults),

  testWithIPSec::
    config.new(defaults {
      network+: {
        cilium+: {
          ipSecKeys: 'test123',
        },
      },
    }),

  testWithFlannel::
    config.new(defaults {
      network+: {
        cilium+: {
        },
        flannel+: {
        },
      },
    }),

  keysDefaults: ciliumManifests($.testDefaults),
  keysWithIPSec: ciliumManifests($.testWithIPSec),
  keysWithFlannel: ciliumManifests($.testWithFlannel),
}

EXPECCTED OUTPUT IN test_config_cilium.json

{
   "keysDefaults": [
      "cilium",
      "cilium-config",
      "cilium-operator"
   ],
   "keysWithFlannel": [
      "cilium",
      "cilium-config",
      "cilium-operator",
      "flannel"
   ],
   "keysWithIPSec": [
      "cilium",
      "cilium-config",
      "cilium-operator"
   ]
}

*/
{}