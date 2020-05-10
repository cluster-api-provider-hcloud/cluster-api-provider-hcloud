{
  _config+:: {
    network+: {
      cilium+: {
      },
    },
  },

  _ipSecEnabled:: std.objectHas($._config.network.cilium, 'ipSecKeys'),

  // setup secret mount
  cilium+: if $._ipSecEnabled then {
    daemonSet+: {
      spec+: {
        template+: {
          spec+: {
            containers: std.map(
              function(x)
                if x.name == 'cilium-agent' then
                  x {
                    volumeMounts+: [{
                      mountPath: '/etc/ipsec',
                      name: 'cilium-ipsec-secrets',
                    }],
                  }
                else x,
              super.containers,
            ),
            volumes+: [
              {
                name: 'cilium-ipsec-secrets',
                secret: {
                  secretName: $.cilium.ipSecKeysSecret.metadata.name,
                },
              },
            ],
          },
        },
      },
    },
    ipSecKeysSecret+: {
      apiVersion: 'v1',
      kind: 'Secret',
      metadata: {
        name: 'cilium-ipsec-keys',
        namespace: 'kube-system',
      },
      type: 'Opaque',
      data: {
        keys: std.base64($._config.network.cilium.ipSecKeys),
      },
    },
  }
  else {},

  // enable config option
  'cilium-config'+: if $._ipSecEnabled then {
    configMap+: {
      data+: {
        'enable-ipsec': 'true',
        'ipsec-key-file': '/etc/ipsec/keys',
      },
    },
  }
  else {},
}
