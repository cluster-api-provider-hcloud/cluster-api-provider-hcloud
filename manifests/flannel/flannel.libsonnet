local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

local makeCPrivileged(array) = [
  c { securityContext+: { privileged: true } }
  for c in array
];

local updateDaemonSet(config, obj) = obj {
  spec+: {
    template+: {
      spec+: {
        containers: makeCPrivileged(super.containers),
        initContainers: if std.objectHas(config.network, 'cilium')
        then []
        else makeCPrivileged(super.initContainers),
      },
    },
  },
};

upstream {
  _config+:: {
    podsCIDRBlock: '192.168.0.0/17',
    //  mtu: 1430,
    network+: {
      flannel+: {
        backend: 'vxlan',
      },
    },
  },

  backendWireguard:: {
    Type: 'extension',
    PreStartupCommand: 'wg genkey | tee privatekey | wg pubkey',
    PostStartupCommand: "export SUBNET_IP=`echo $SUBNET | cut -d'/' -f 1`; ip link del flannel-wg 2>/dev/null; ip link add flannel-wg type wireguard && wg set flannel-wg listen-port 51820 private-key privatekey && ip addr add $SUBNET_IP/32 dev flannel-wg && ip link set flannel-wg up && ip route add $NETWORK dev flannel-wg",
    ShutdownCommand: 'ip link del flannel-wg',
    SubnetAddCommand: 'read PUBLICKEY; wg set flannel-wg peer $PUBLICKEY endpoint $PUBLIC_IP:51820 allowed-ips $SUBNET',
    SubnetRemoveCommand: 'read PUBLICKEY; wg set flannel-wg peer $PUBLICKEY remove',
  },

  // build cni config
  config+:: {
    cni+::
      {
        name: 'cbr0',
        cniVersion: '0.3.1',
        plugins: [
          {
            type: 'flannel',
            delegate: {
              hairpinMode: true,
              isDefaultGateway: true,
            },
          },
          {
            type: 'portmap',
            capabilities: {
              portMappings: true,
            },
          },
        ],
      },
    net+::
      {
        Network: $._config.podsCIDRBlock,
        Backend: if $._config.network.flannel.backend == 'wireguard' then $.backendWireguard else {
          Type: $._config.network.flannel.backend,
        },
      },
  },

  // selinux needs privileged enabled
  'psp.flannel.unprivileged'+: {
    podSecurityPolicy+: {
      spec+: {
        privileged: true,
      },
    },
  },

  'kube-flannel-cfg'+: {
    configMap+: {
      data+: {
        'cni-conf.json': std.manifestJsonEx($.config.cni, '  '),
        'net-conf.json': std.manifestJsonEx($.config.net, '  '),
      },
    },
  },

  'kube-flannel-ds-amd64'+: {
    daemonSet: updateDaemonSet($._config, super.daemonSet),
  },
  'kube-flannel-ds-arm'+: {
    daemonSet: updateDaemonSet($._config, super.daemonSet),
  },
  'kube-flannel-ds-arm64'+: {
    daemonSet: updateDaemonSet($._config, super.daemonSet),
  },
  'kube-flannel-ds-ppc64le'+: {
    daemonSet: updateDaemonSet($._config, super.daemonSet),
  },
  'kube-flannel-ds-s390x'+: {
    daemonSet: updateDaemonSet($._config, super.daemonSet),
  },

}
