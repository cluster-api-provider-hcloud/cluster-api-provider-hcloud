local config = import 'config.jsonnet';

{
  test:: config {
    _config+: {
      hcloudToken: 'my-token',
      hcloudNetwork: 'my-network',
    },
  },

  keys: std.objectFields($.test),
}
