local calico = import 'calico.libsonnet';

{
  tree:: calico {
    _config+:: {
      podsCIDRBlock: '10.123.0.0/17',
    },
  },

  values:
    std.filter(
      function(x)
        if x.name == 'CALICO_IPV4POOL_CIDR' || x.name == 'FELIX_XDPENABLED' then true else false,
      $.tree['calico-node'].daemonSet.spec.template.spec.containers[0].env,
    ),
}
