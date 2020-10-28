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

local newHcloudSecret(network, token, ip, domain) = 
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

local newHrobotSecret(user, password) = 
  if (user =="") then {
    } else {
    hrobotSecret: {
      apiVersion: 'v1',
      kind: 'Secret',
      metadata: {
        name: 'hrobot',
        namespace: 'kube-system',
      },
      type: 'Opaque',
      data: {
        'robot-user': std.base64(user),
        'robot-password': std.base64(password),
      },
    },
};

local newCASecret(caCrt, caKey) =
  if (caCrt =="") then {
    } else {
  vcManagerNamespace: {
    apiVersion: 'v1',
    kind: 'Namespace',
    metadata: {
      name: 'vc-manager',
    },
  },
  caSecret: {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'vc-kubelet-client',
      namespace: 'vc-manager',
    },
    type: 'Opaque',
    data: {
      'client.crt': std.base64(caCrt),
      'client.key': std.base64(caKey),
    },
  },
};

local addons = {
  secrets: newHcloudSecret($._config.hcloudNetwork, $._config.hcloudToken, $._config.kubeAPIServerIPv4, $._config.kubeAPIServerDomain)
  + newHrobotSecret($._config.robotUserName, $._config.robotPassword) + newCASecret($._config.caCrt, $._config.caKey),
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
