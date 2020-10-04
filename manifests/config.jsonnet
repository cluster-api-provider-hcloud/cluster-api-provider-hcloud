local specs(ip, domain, port) =
  if (domain == "") then {
    selector: {
      component: 'kube-apiserver',
      tier: 'control-plane',
    },
    ports: [
      {
        protocol: 'TCP',
        port: port,
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
        port: port,
        targetPort: 6443,
      },
    ],
    type: 'LoadBalancer',
    loadBalancerIP: ip,
    externalTrafficPolicy: 'Local',
    externalName: domain,
    externalIPs: [
      ip,
    ],
};

local newControlPlaneService(ip, domain, port) = {
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: 'kube-apiserver',
    namespace: 'kube-system',
  },
  spec: specs(ip, domain, port),
};

local newSecrets(network, token, ip, domain) = 
  if (domain =="") then {
    hcloudSecret: {
      apiVersion: 'v1',
      kind: 'Secret',
      metadata: {
        name: 'hcloud',
        namespace: 'kube-system',
      },
      type: 'Opaque',
      data: {
        network: std.base64(network),
        token: std.base64(token),
        apiserver: std.base64(ip),
      },
    },
    } else {
    hcloudSecret: {
      apiVersion: 'v1',
      kind: 'Secret',
      metadata: {
        name: 'hcloud',
        namespace: 'kube-system',
      },
      type: 'Opaque',
      data: {
        network: std.base64(network),
        token: std.base64(token),
        apiserver: std.base64(domain),
      },
    },
};

local addons = {
  secrets: newSecrets($._config.hcloudNetwork, $._config.hcloudToken, $._config.kubeAPIServerIPv4, $._config.kubeAPIServerDomain),
  controlPlaneServices: newControlPlaneService($._config.kubeAPIServerIPv4, $._config.kubeAPIServerDomain, $._config.port),
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
