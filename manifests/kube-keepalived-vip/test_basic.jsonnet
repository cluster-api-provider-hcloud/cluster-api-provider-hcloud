local kubeKeepalivedVIP = import 'kube-keepalived-vip.libsonnet';

{
  tree:: kubeKeepalivedVIP {
    _config+:: {
      backendService: 'default/kubernetes',
      hcloudFloatingIPs: ['1.2.3.4'],
      hcloudTokenRef: {
        valueFrom: {
          secretKeyRef: {
            key: 'token-123',
            name: 'hcloud-123',
          },
        },
      },
    },
  },

  configMapData: $.tree.configMap.data,

  local env = $.tree.daemonSet.spec.template.spec.containers[0].env,

  envHcloudTokenRef: std.filter(
    function(x)
      if x.name == 'NOTIFY_HCLOUD_TOKEN' then true else false,
    env
  ),

  envFloatingIP: std.filter(
    function(x)
      if x.name == 'NOTIFY_FLOATING_IPS' then true else false,
    env
  ),
}
