local config = import 'config.jsonnet';

{
  test:: config.new({
    hcloudToken: 'my-token',
    hcloudNetwork: 'my-network',
  }),

  keys: std.objectFields($.test),
}
