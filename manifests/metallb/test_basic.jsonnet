local metallb = import 'metallb.libsonnet';

{
  tree:: metallb {
    _config+:: {
      podsCIDRBlock: '10.123.0.0/17',
    },
  },

  values:
    std.filter(
      function(x)
        if x.name == 'METALLB_ML_BIND_ADDR' || x.name == 'METALLB_ML_SECRET_KEY' then true else false,
      $.tree.speaker.daemonSet.spec.template.spec.containers[0].env,
    ),
}
