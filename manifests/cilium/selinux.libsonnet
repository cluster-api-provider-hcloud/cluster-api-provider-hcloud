{
  // This enables cilium SELinux setup for Centos
  cilium+: {
    daemonSet+: {
      spec+: {
        template+: {
          spec+: {
            hostPID: true,
            initContainers: [
              {
                name: 'install-selinux-module',
                image: 'simonswine/cilium-selinux-module:0.1',
                imagePullPolicy: 'IfNotPresent',
                resources: {
                  requests: {
                    cpu: '10m',
                  },
                },
                securityContext: {
                  privileged: true,
                },
                volumeMounts: [
                  {
                    mountPath: '/etc/selinux',
                    name: 'etc-selinux',
                  },
                  {
                    mountPath: '/var/lib/selinux',
                    name: 'var-lib-selinux',
                  },
                ],
              },
            ] + super.initContainers,
            volumes+: [
              {
                hostPath: {
                  path: '/etc/selinux',
                  type: 'Directory',
                },
                name: 'etc-selinux',
              },
              {
                hostPath: {
                  path: '/var/lib/selinux',
                  type: 'Directory',
                },
                name: 'var-lib-selinux',
              },
            ],
          },
        },
        updateStrategy: {
          rollingUpdate: {
            maxUnavailable: 1,
          },
          type: 'RollingUpdate',
        },
      },
    },
  },
}
