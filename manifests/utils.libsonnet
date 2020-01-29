local lowercaseFirstChar(s) =
  std.asciiLower(s[0]) + std.substr(s, 1, std.length(s));

local appendKind(m, o) =
  if o == null then
    m
  else
    local kind = lowercaseFirstChar(o.kind);
    local name = o.metadata.name;
    if kind == 'customResourceDefinition' then
      m {
        [kind]+: {
          [name]+: o,
        },
      }
    else
      m {
        [name]+: {
          [kind]+: o,
        },
      }
;

local setOrCreateKey(env, value) =
  local names = std.map(function(x) x.name, env);
  local matches = std.find(value.name, names);
  if std.length(matches) > 0 then
    std.map(
      function(x)
        if x.name == value.name then
          value
        else
          x,
      env
    )
  else
    env + [value];


{
  convertManifests(obj):: std.foldl(appendKind, obj, {}),
  mergeEnv(env, newValues):: std.foldl(setOrCreateKey, newValues, env),
}
