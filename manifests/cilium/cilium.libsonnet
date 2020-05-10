local utils = import '../utils.libsonnet';
local ipsec = import './ipsec.libsonnet';
local selinux = import './selinux.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

upstream {
  _config+:: {
    //  podsCIDRBlock: '192.168.0.0/17',
    //  mtu: 1430,
    network+: {
      cilium+: {
      },
    },
  },

  'cilium-config'+: {
    configMap+: {
      data+: {
        'enable-external-ips': 'true',
        'enable-node-port': 'true',
        'kube-proxy-replacement': 'disabled',
      } + (if std.objectHas($._config.network, 'flannel')
           then {
             'enable-endpoint-health-checking': 'false',
             'enable-local-node-route': 'false',
             tunnel: 'disabled',
             masquerade: 'false',
             'flannel-master-device': 'cni0',
             'flannel-uninstall-on-exit': 'false',
           }
           else {
             'cni-chaining-mode': 'portmap',
           }),
    },
  },
} + ipsec + selinux
