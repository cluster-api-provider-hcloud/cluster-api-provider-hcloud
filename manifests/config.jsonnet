
local newControlPlaneService(pos, ip) = {
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: 'kube-apiserver-%d' % pos,
    namespace: 'kube-system',
  },
  spec: {
    selector: {
      component: 'kube-apiserver',
      tier: 'control-plane',
    },
    ports: [
      {
        protocol: 'TCP',
        port: 6443,
        targetPort: 6443,
      },
    ],
    type: 'LoadBalancer',
    loadBalancerIP: ip,
    externalTrafficPolicy: 'Local',
    externalIPs: [
      ip,
    ],
  },
};

local addons = {
  hcloudSecret: {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'hcloud',
      namespace: 'kube-system',
    },
    type: 'Opaque',
    data: {
      network: std.base64($._config.hcloudNetwork),
      token: std.base64($._config.hcloudToken),
    },
  },

  controlPlaneServices: std.mapWithIndex(newControlPlaneService, $._config.hcloudLoadBalancerIPv4s),

  workarounds: {
    // This fixes a problem join v1.18 node to a v1.17 control plane
    // https://github.com/kubernetes/kubeadm/issues/2079
    'upgrade-hotfix-v1.18': {
      clusterRole: {
        apiVersion: 'rbac.authorization.k8s.io/v1',
        kind: 'ClusterRole',
        metadata: {
          name: 'kubeadm:get-nodes',
        },
        rules: [
          {
            apiGroups: [
              '',
            ],
            resources: [
              'nodes',
            ],
            verbs: [
              'get',
            ],
          },
        ],
      },
      clusterRoleBinding: {
        apiVersion: 'rbac.authorization.k8s.io/v1',
        kind: 'ClusterRoleBinding',
        metadata: {
          name: 'kubeadm:get-nodes',
        },
        roleRef: {
          apiGroup: 'rbac.authorization.k8s.io',
          kind: 'ClusterRole',
          name: 'kubeadm:get-nodes',
        },
        subjects: [
          {
            apiGroup: 'rbac.authorization.k8s.io',
            kind: 'Group',
            name: 'system:bootstrappers:kubeadm:default-node-token',
          },
        ],
      },
    },
  },
};


local new(c) = (
  {
    _config+:: c,
  } +
  addons
);

{
  new(config)::
    new(config),

  example: new({}),
}
