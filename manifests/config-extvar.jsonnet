local config = import 'config.jsonnet';

config {
  _config+:: {
    hcloudToken: std.extVar('hcloud-token'),
    hcloudNetwork: std.extVar('hcloud-network'),
    hcloudFloatingIPs: std.split(std.extVar('hcloud-floating-ips'), ','),
    podsCIDRBlock: std.extVar('pod-cidr-block'),
  },
}
