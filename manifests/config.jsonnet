local calico = import 'calico/calico.libsonnet';
local flannel = import 'flannel/flannel.libsonnet';
local hcloudCloudControllerManager = import 'hcloud-cloud-controller-manager/hcloud-cloud-controller-manager.libsonnet';
local hcloudCSI = import 'hcloud-csi/hcloud-csi.libsonnet';
local metricsServer = import 'metrics-server/metrics-server.libsonnet';

local AllManifests = {
    "calico": import "calico/calico.libsonnet",
    "cilium": import 'cilium/cilium.libsonnet',
    "hcloudCSI": import "hcloud-csi/hcloud-csi.libsonnet",
    "metricsServer": import "metrics-server/metrics-server.libsonnet",
    "hcloudCloudControllerManager": import "hcloud-cloud-controller-manager/hcloud-cloud-controller-manager.libsonnet"
};

local getManifestFromKey(x) = 
  if std.objectHas(AllManifests,x) then
    AllManifests[x] 
  else
    std.trace("Manifest key not found: " + x, {});

local join_objects(objs) =
  local aux(arr, i, running) =
    if i >= std.length(arr) then
      running
    else
      aux(arr, i + 1, running + arr[i]) tailstrict;
  aux(objs, 0, {});

local getManifests(keys) = join_objects(std.map(getManifestFromKey, keys));

local defaultConfig = {
  hcloudNetworkRef: {
    valueFrom: {
      secretKeyRef: {
        key: 'network',
        name: 'hcloud',
      },
    },
  },
  hcloudTokenRef: {
    valueFrom: {
      secretKeyRef: {
        key: 'token',
        name: 'hcloud',
      },
    },
  },
  podsCIDRBlock: '192.168.0.0/16',
  hcloudToken: 'xx',
  hcloudNetwork: 'yy',
  hcloudLoadBalancerIPv4s: ['1.1.1.1', '2.2.2.2'],
};

local specs(ip, domain) =
  if (domain == "") then {
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
  } else {
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
    hostName: domain,
    externalIPs: [
      ip,
    ],
};


local newControlPlaneService(ip, domain) = {
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: 'kube-apiserver,
    namespace: 'kube-system',
  },
  spec: specs(ip, domain),
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

  controlPlaneServices: newControlPlaneService($._config.kubeAPIServerIPv4, $._config.kubeAPIServerDomain),

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
    _config+:: defaultConfig,
  } +
  (if std.objectHas(c, 'manifests') then
    getManifests(c.manifests)
  else 
    {}) +
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
