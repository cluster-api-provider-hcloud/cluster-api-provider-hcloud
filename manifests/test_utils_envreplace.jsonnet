local utils = import 'utils.libsonnet';

{
  input:: {
    containers: [
      { env: [{ name: 'NAME', value: 'VALUEOLD' }] },
      { env: [{ name: 'NAMEX', value: 'VALUEOLD' }] },
    ],
  },
  test1: utils.recursiveEnvReplace($.input, { name: 'NAMEX', value: 'VALUENEW' }),
  test2: utils.recursiveEnvReplace($.input, { name: 'NAME', value: 'VALUENEW' }),
  test3: utils.recursiveEnvReplace($.input, { name: 'NAMEY', value: 'VALUENEW' }),
}
