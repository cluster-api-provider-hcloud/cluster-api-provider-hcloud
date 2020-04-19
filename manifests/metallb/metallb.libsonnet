local utils = import '../utils.libsonnet';
local upstream = utils.convertManifests(import 'manifests.json');

upstream {
  _config+:: {
    hcloudFloatingIPs: [],

    images+: {
      'metallb-controller': 'docker.io/metallb/controller@sha256:c0c6f8655f9c855bc6e10c9a9975413d253b91063e91732021f07eca140797eb',  // v0.9.3
      'metallb-speaker': 'docker.io/simonswine/metallb-speaker@sha256:ec255f41e5ccd8cbaefe5d341056457ce2d16528a08b51d31fb9e55036266d6e',  // v0.9.3-simonswine2
    },

    config: {
      'address-pools': [{
        name: 'kube-apiserver',
        protocol: 'layer2',
        addresses: [
          x + '/32'
          for x in $._config.hcloudFloatingIPs
        ],
      }],
    },
  },

  '10-namespace': {
    apiVersion: 'v1',
    kind: 'Namespace',
    metadata: {
      name: $.controller.deployment.metadata.namespace,
      labels: {
        name: $.controller.deployment.metadata.namespace,
      },
    },
  },

  configMap: {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: $.controller.deployment.metadata {
      name: 'config',
    },
    data: {
      config: std.manifestYamlDoc($._config.config),
    },
  },

  controller+: {
    local this = self,
    container:: super.deployment.spec.template.spec.containers[0] {
      image: $._config.images['metallb-controller'],
      imagePullPolicy: 'IfNotPresent',
    },

    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [this.container],
            tolerations: [
              {
                effect: 'NoSchedule',
                key: 'node-role.kubernetes.io/master',
              },
            ],
          },
        },
      },
    },
  },


  speaker+: {
    local this = self,
    container:: super.daemonSet.spec.template.spec.containers[0] {
      image: $._config.images['metallb-speaker'],
      imagePullPolicy: 'IfNotPresent',
      env: utils.mergeEnv(super.env, [
        // This remove the member list parsing for the speakers
        { name: 'METALLB_ML_BIND_ADDR', value: '' },
        { name: 'METALLB_ML_SECRET_KEY', value: '' },
      ]),
    },

    daemonSet+: {
      spec+: {
        template+: {
          spec+: {
            containers: [this.container],
            tolerations+: [{
              key: 'node.kubernetes.io/unreachable',
              operator: 'Exists',
              effect: 'NoExecute',
              tolerationSeconds: 60,
            }],
          },
        },
      },
    },
  },

  'metallb-system:speaker'+: {
    clusterRole+: {
      rules+: [
        {
          apiGroups: [
            '',
          ],
          resources: [
            'services',
          ],
          verbs: [
            'update',
            'patch',
          ],
        },
      ],
    },
  },


}
