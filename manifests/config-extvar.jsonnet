local config = import 'config.jsonnet';

config {
  _config+:: {
    hcloudToken: std.extVar('hcloud-token'),
    hcloudNetwork: std.extVar('hcloud-network'),
    podsCIDRBlock: std.extVar('pod-cidr-block'),
  },
}
