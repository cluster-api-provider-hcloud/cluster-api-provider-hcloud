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

local recursiveEnvReplaceFound(obj, value) = std.map(
  function(x)
    if x.name == value.name then
      value
    else
      x,
  obj,
)
;


{
  convertManifests(obj):: std.foldl(appendKind, obj, {}),
  mergeEnv(env, newValues):: std.foldl(setOrCreateKey, newValues, env),

  // set environment variables recursivly if they exist
  recursiveEnvReplace(obj, value)::
    if std.type(obj) == 'object' then
      std.mapWithKey(
        function(name, field)
          if name == 'env' && std.type(field) == 'array' then
            recursiveEnvReplaceFound(field, value)
          else
            $.recursiveEnvReplace(field, value),
        obj,
      )
    else if std.type(obj) == 'array' then
      std.map(function(x) $.recursiveEnvReplace(x, value), obj)
    else
      obj,
}
