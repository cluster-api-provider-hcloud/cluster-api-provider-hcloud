load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_repositories():
    go_repository(
        name = "com_github_alecthomas_template",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alecthomas/template",
        sum = "h1:cAKDfWh5VpdgMhJosfJnn5/FoN2SRZ4p7fJNX58YPaU=",
        version = "v0.0.0-20160405071501-a0175ee3bccc",
    )
    go_repository(
        name = "com_github_alecthomas_units",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alecthomas/units",
        sum = "h1:qet1QNfXsQxTZqLG4oE62mJzwPIB8+Tee4RNCL9ulrY=",
        version = "v0.0.0-20151022065526-2efee857e7cf",
    )
    go_repository(
        name = "com_github_armon_consul_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/armon/consul-api",
        sum = "h1:G1bPvciwNyF7IUmKXNt9Ak3m6u9DE1rF+RmtIkBpVdA=",
        version = "v0.0.0-20180202201655-eb2c6b5be1b6",
    )
    go_repository(
        name = "com_github_asaskevich_govalidator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/asaskevich/govalidator",
        sum = "h1:eg0MeVzsP1G42dRafH3vf+al2vQIJU0YHX+1Tw87oco=",
        version = "v0.0.0-20180720115003-f9ffefc3facf",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/aws/aws-sdk-go",
        sum = "h1:R5YqHQFIulYVfgRySz9hvBRTWBjudISa+r0C8XQ1ufg=",
        version = "v1.25.43",
    )
    go_repository(
        name = "com_github_awslabs_goformation",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/awslabs/goformation",
        sum = "h1:1y3J+mle+xAD+J7IlHuY2fK5iLbZqGOm+isHdrW7kLQ=",
        version = "v0.0.0-20190310235947-776555df5a6d",
    )
    go_repository(
        name = "com_github_azure_go_ansiterm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-ansiterm",
        sum = "h1:w+iIsaOQNcT7OZ575w+acHgRric5iCyQh+xv+KJ4HB8=",
        version = "v0.0.0-20170929234023-d6e3b3328b78",
    )
    go_repository(
        name = "com_github_azure_go_autorest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/Azure/go-autorest",
        sum = "h1:viZ3tV5l4gE2Sw0xrasFHytCGtzYCrT+um/rrSQ1BfA=",
        version = "v11.1.2+incompatible",
    )
    go_repository(
        name = "com_github_beorn7_perks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/beorn7/perks",
        sum = "h1:VlbKKnNfV8bJzeqoa4cOKqO6bYr3WgKZxO8Z16+hsOM=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_blang_semver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/blang/semver",
        sum = "h1:CGxCgetQ64DKk7rdZ++Vfnb1+ogGNnB17OJKJXD2Cfs=",
        version = "v3.5.0+incompatible",
    )
    go_repository(
        name = "com_github_bronze1man_go_yaml2json",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bronze1man/go-yaml2json",
        sum = "h1:YtNfk92cLL9IKGS7pNxnxN4OXWw5S2efi9/6jcwzTdY=",
        version = "v0.0.0-20150129175009-f6f64b738964",
    )
    go_repository(
        name = "com_github_bronze1man_yaml2json",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bronze1man/yaml2json",
        sum = "h1:H7QXuITFovnVHZcSbOWCUM0Xs7oDsioAClU5ezTzf/M=",
        version = "v0.0.0-20190501122504-861f66b7262b",
    )
    go_repository(
        name = "com_github_burntsushi_toml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/BurntSushi/toml",
        sum = "h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_burntsushi_xgb",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/BurntSushi/xgb",
        sum = "h1:1BDTz0u9nC3//pOCMdNH+CiXJVYJh5UQNCOBG7jbELc=",
        version = "v0.0.0-20160522181843-27f122750802",
    )
    go_repository(
        name = "com_github_coreos_bbolt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/bbolt",
        sum = "h1:uTXKg9gY70s9jMAKdfljFQcuh4e/BXOM+V+d00KFj3A=",
        version = "v1.3.1-coreos.6",
    )
    go_repository(
        name = "com_github_coreos_etcd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/etcd",
        sum = "h1:8F3hqu9fGYLBifCmRCJsicFqDx/D68Rt3q1JMazcgBQ=",
        version = "v3.3.13+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_etcd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-etcd",
        sum = "h1:bXhRBIXoTm9BYHS3gE0TtQuyNZyeEMux2sDi4oo5YOo=",
        version = "v2.0.0+incompatible",
    )
    go_repository(
        name = "com_github_coreos_go_oidc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-oidc",
        sum = "h1:X+JQSgXg3CcxgcBoMAqU8NoS0fch8zHxjiKWcXclxaI=",
        version = "v0.0.0-20180117170138-065b426bd416",
    )
    go_repository(
        name = "com_github_coreos_go_semver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-semver",
        sum = "h1:3Jm3tLmsgAYcjC+4Up7hJrFBPr+n7rAqYeSw/SZazuY=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_coreos_go_systemd",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/go-systemd",
        sum = "h1:u9SHYsPQNyt5tgDm3YN7+9dYrpK96E5wFilTFWIDZOM=",
        version = "v0.0.0-20180511133405-39ca1b05acc7",
    )
    go_repository(
        name = "com_github_coreos_pkg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/coreos/pkg",
        sum = "h1:n2Ltr3SrfQlf/9nOna1DoGKxLx3qTSI8Ttl6Xrqp6mw=",
        version = "v0.0.0-20180108230652-97fdf19511ea",
    )
    go_repository(
        name = "com_github_cosiner_argv",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cosiner/argv",
        sum = "h1:rIXlvz2IWiupMFlC45cZCXZFvKX/ExBcSLrDy2G0Lp8=",
        version = "v0.0.0-20170225145430-13bacc38a0a5",
    )
    go_repository(
        name = "com_github_cpuguy83_go_md2man",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cpuguy83/go-md2man",
        sum = "h1:BSKMNlYxDvnunlTymqtgONjNnaRV1sTpcovwwjF22jk=",
        version = "v1.0.10",
    )
    go_repository(
        name = "com_github_davecgh_go_spew",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_dgrijalva_jwt_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/dgrijalva/jwt-go",
        sum = "h1:NyywMz59neOoVRFDz+ccfKWxn784fiHMDnZSy6T+JXY=",
        version = "v0.0.0-20160705203006-01aeca54ebda",
    )
    go_repository(
        name = "com_github_docker_docker",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/docker",
        sum = "h1:w3NnFcKR5241cfmQU5ZZAsf0xcpId6mWOupTvJlUX2U=",
        version = "v0.7.3-0.20190327010347-be7ac8be2ae0",
    )
    go_repository(
        name = "com_github_docker_go_units",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/go-units",
        sum = "h1:Xk8S3Xj5sLGlG5g67hJmYMmUgXv5N4PhkjJHHqrwnTk=",
        version = "v0.3.3",
    )
    go_repository(
        name = "com_github_docker_spdystream",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docker/spdystream",
        sum = "h1:cenwrSVm+Z7QLSV/BsnenAOcDXdX4cMv4wP0B/5QbPg=",
        version = "v0.0.0-20160310174837-449fdfce4d96",
    )
    go_repository(
        name = "com_github_docopt_docopt_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/docopt/docopt-go",
        sum = "h1:bWDMxwH3px2JBh6AyO7hdCn/PkvCZXii8TGj7sbtEbQ=",
        version = "v0.0.0-20180111231733-ee0de3bc6815",
    )
    go_repository(
        name = "com_github_elazarl_goproxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/elazarl/goproxy",
        sum = "h1:p1yVGRW3nmb85p1Sh1ZJSDm4A4iKLS5QNbvUHMgGu/M=",
        version = "v0.0.0-20170405201442-c4fc26588b6e",
    )
    go_repository(
        name = "com_github_emicklei_go_restful",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/emicklei/go-restful",
        sum = "h1:H2pdYOb3KQ1/YsqVWoWNLQO+fusocsw354rqGTZtAgw=",
        version = "v0.0.0-20170410110728-ff4f55a20633",
    )
    go_repository(
        name = "com_github_evanphx_json_patch",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/evanphx/json-patch",
        sum = "h1:ouOWdg56aJriqS0huScTkVXPC5IcNrDCXZ6OoTAWu7M=",
        version = "v4.5.0+incompatible",
    )
    go_repository(
        name = "com_github_fsnotify_fsnotify",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/fsnotify/fsnotify",
        sum = "h1:IXs+QLmnXW2CcXuY+8Mzv/fWEsPGWxqefPtCP5CnV9I=",
        version = "v1.4.7",
    )
    go_repository(
        name = "com_github_ghodss_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ghodss/yaml",
        sum = "h1:bRzFpEzvausOAt4va+I/22BZ1vXDtERngp0BNYDKej0=",
        version = "v0.0.0-20180820084758-c7ce16629ff4",
    )
    go_repository(
        name = "com_github_globalsign_mgo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/globalsign/mgo",
        sum = "h1:DujepqpGd1hyOd7aW59XpK7Qymp8iy83xq74fLr21is=",
        version = "v0.0.0-20181015135952-eeefdecb41b8",
    )
    go_repository(
        name = "com_github_go_delve_delve",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-delve/delve",
        sum = "h1:K8VjV+Q2YnBYlPq0ctjrvc9h7h03wXszlszzfGW5Tog=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_go_kit_kit",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-kit/kit",
        sum = "h1:Wz+5lgoB0kkuqLEc6NVmwRknTKP6dTGbSqvhZtBI/j0=",
        version = "v0.8.0",
    )
    go_repository(
        name = "com_github_go_logfmt_logfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-logfmt/logfmt",
        sum = "h1:8HUsc87TaSWLKwrnumgC8/YconD2fJQsRJAsWaPg2ic=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_go_logr_logr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-logr/logr",
        sum = "h1:M1Tv3VzNlEHg6uyACnRdtrploV2P7wZqH8BoQMtz0cg=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_go_logr_zapr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-logr/zapr",
        sum = "h1:h+WVe9j6HAA01niTJPA/kKH0i7e0rLZBCwauQFcRE54=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_go_openapi_analysis",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/analysis",
        sum = "h1:eYp14J1o8TTSCzndHBtsNuckikV1PfZOSnx4BcBeu0c=",
        version = "v0.17.2",
    )
    go_repository(
        name = "com_github_go_openapi_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/errors",
        sum = "h1:azEQ8Fnx0jmtFF2fxsnmd6I0x6rsweUF63qqSO1NmKk=",
        version = "v0.17.2",
    )
    go_repository(
        name = "com_github_go_openapi_jsonpointer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/jsonpointer",
        sum = "h1:FTUMcX77w5rQkClIzDtTxvn6Bsa894CcrzNj2MMfeg8=",
        version = "v0.19.0",
    )
    go_repository(
        name = "com_github_go_openapi_jsonreference",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/jsonreference",
        sum = "h1:BqWKpV1dFd+AuiKlgtddwVIFQsuMpxfBDBHGfM2yNpk=",
        version = "v0.19.0",
    )
    go_repository(
        name = "com_github_go_openapi_loads",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/loads",
        sum = "h1:tEXYu6Xc0pevpzzQx5ghrMN9F7IVpN/+u4iD3rkYE5o=",
        version = "v0.17.2",
    )
    go_repository(
        name = "com_github_go_openapi_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/runtime",
        sum = "h1:/ZK67ikFhQAMFFH/aPu2MaGH7QjP4wHBvHYOVIzDAw0=",
        version = "v0.17.2",
    )
    go_repository(
        name = "com_github_go_openapi_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/spec",
        sum = "h1:eb2NbuCnoe8cWAxhtK6CfMWUYmiFEZJ9Hx3Z2WRwJ5M=",
        version = "v0.17.2",
    )
    go_repository(
        name = "com_github_go_openapi_strfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/strfmt",
        sum = "h1:1isAxYf//QDTnVzbLAMrUK++0k1EjeLJU/gTOR0o3Mc=",
        version = "v0.17.0",
    )
    go_repository(
        name = "com_github_go_openapi_swag",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/swag",
        sum = "h1:K/ycE/XTUDFltNHSO32cGRUhrVGJD64o8WgAIZNyc3k=",
        version = "v0.17.2",
    )
    go_repository(
        name = "com_github_go_openapi_validate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-openapi/validate",
        sum = "h1:PVXYcP1GkTl+XIAJnyJxOmK6CSG5Q1UcvoCvNO++5Kg=",
        version = "v0.18.0",
    )
    go_repository(
        name = "com_github_go_stack_stack",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-stack/stack",
        sum = "h1:5SgMzNM5HxrEjV0ww2lTmX6E2Izsfxas4+YHWRs3Lsk=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_github_gogo_protobuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gogo/protobuf",
        sum = "h1:3PaI8p3seN09VjbTYC/QWlUZdZ1qS1zGjy7LH2Wt07I=",
        version = "v1.2.2-0.20190723190241-65acae22fc9d",
    )
    go_repository(
        name = "com_github_golang_glog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/glog",
        sum = "h1:VKtxabqXZkF25pY9ekfRL6a582T4P37/31XEstQ5p58=",
        version = "v0.0.0-20160126235308-23def4e6c14b",
    )
    go_repository(
        name = "com_github_golang_groupcache",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/groupcache",
        sum = "h1:veQD95Isof8w9/WXiA+pa3tz3fJXkt5B7QaRBrM62gk=",
        version = "v0.0.0-20190129154638-5b532d6fd5ef",
    )
    go_repository(
        name = "com_github_golang_mock",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/mock",
        sum = "h1:qGJ6qTW+x6xX/my+8YUVl4WNpX9B7+/l2tRsHGZ7f2s=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_golang_protobuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/protobuf",
        sum = "h1:6nsPYzhq5kReh6QImI3k5qWzO4PEbvbIW2cwSfR/6xs=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_google_btree",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/btree",
        sum = "h1:JHB7F/4TJCrYBW8+GZO8VkWDj1jxcWuCl6uxKODiyi4=",
        version = "v0.0.0-20160524151835-7d79101e329e",
    )
    go_repository(
        name = "com_github_google_go_cmp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-cmp",
        sum = "h1:Xye71clBPdm5HgqGwUkwhbynsUJZhDbS20FvLhQ2izg=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_google_gofuzz",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/gofuzz",
        sum = "h1:A8PeW59pxE9IoFRqBp37U+mSNaQoZ46F1f0f863XSXw=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_google_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/uuid",
        sum = "h1:b4Gk+7WdP/d3HZH8EJsZpvV7EtDOgaZLtnaNGIu1adA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_googleapis_gnostic",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/googleapis/gnostic",
        sum = "h1:WeAefnSUHlBb0iJKwxFDZdbfGwkd7xRNuV+IpXMJhYk=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_gophercloud_gophercloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gophercloud/gophercloud",
        sum = "h1:6sjpKIpVwRIIwmcEGp+WwNovNsem+c+2vm6oxshRpL8=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_gorilla_websocket",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gorilla/websocket",
        sum = "h1:Lh2aW+HnU2Nbe1gqD9SOJLJxW1jBMmQOktN2acDyJk8=",
        version = "v0.0.0-20170926233335-4201258b820c",
    )
    go_repository(
        name = "com_github_gregjones_httpcache",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gregjones/httpcache",
        sum = "h1:6TSoaYExHper8PYsJu23GWVNOyYRCSnIFyxKgLSZ54w=",
        version = "v0.0.0-20170728041850-787624de3eb7",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_middleware",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/go-grpc-middleware",
        sum = "h1:lR9ssWAqp9qL0bALxqEEkuudiP1eweOdv9jsRK3e7lE=",
        version = "v0.0.0-20190222133341-cfaf5686ec79",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_prometheus",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
        sum = "h1:f5vL2EW5pL274ztMNnizZAEa457nKyKPEaN/sm/kdBk=",
        version = "v0.0.0-20170330212424-2500245aa611",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_grpc_gateway",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/grpc-gateway",
        sum = "h1:HJtP6RRwj2EpPCD/mhAWzSvLL/dFTdPm1UrWwanoFos=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_hashicorp_golang_lru",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/golang-lru",
        sum = "h1:YPkqC67at8FYaadspW/6uE0COsBxS2656RLEr8Bppgk=",
        version = "v0.5.3",
    )
    go_repository(
        name = "com_github_hashicorp_hcl",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hashicorp/hcl",
        sum = "h1:0Anlzjpi4vEasTeNFn2mLJgTSwt0+6sfsiTG8qcWGx4=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hetznercloud_hcloud_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hetznercloud/hcloud-go",
        sum = "h1:IKH0GLLoTEfgMuBY+GaaVTwjYChecrHFVo4/t0sIkGU=",
        version = "v1.17.0",
    )
    go_repository(
        name = "com_github_hpcloud_tail",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/hpcloud/tail",
        sum = "h1:nfCOvKYfkgYP8hkirhJocXT2+zOD8yUNjXaWfTlyFKI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_imdario_mergo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/imdario/mergo",
        sum = "h1:CGgOkSJeqMRmt0D9XLWExdT4m4F1vd3FV3VPt+0VxkQ=",
        version = "v0.3.8",
    )
    go_repository(
        name = "com_github_inconshreveable_mousetrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/inconshreveable/mousetrap",
        sum = "h1:Z8tu5sraLXCXIcARxBp/8cbvlwVa7Z1NHg9XEKhtSvM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jmespath_go_jmespath",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jmespath/go-jmespath",
        sum = "h1:pmfjZENx5imkbgOkpRUYLnmbU7UEFbjtDA2hxJ1ichM=",
        version = "v0.0.0-20180206201540-c2b33e8439af",
    )
    go_repository(
        name = "com_github_jonboulle_clockwork",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/jonboulle/clockwork",
        sum = "h1:XpRROA6ssPlTwJI8/pH+61uieOkcJhmAFz25cu0B94Y=",
        version = "v0.0.0-20141017032234-72f9bd7c4e0c",
    )
    go_repository(
        name = "com_github_json_iterator_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/json-iterator/go",
        sum = "h1:KfgG9LzI+pYjr4xvmz/5H4FXjokeP+rlHLhv3iH62Fo=",
        version = "v1.1.7",
    )
    go_repository(
        name = "com_github_julienschmidt_httprouter",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/julienschmidt/httprouter",
        sum = "h1:TDTW5Yz1mjftljbcKqRcrYhd4XeOoI98t+9HbQbYf7g=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_kisielk_errcheck",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kisielk/errcheck",
        sum = "h1:reN85Pxc5larApoH1keMBiu2GWtPqXQ1nc9gx+jOU+E=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_kisielk_gotool",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kisielk/gotool",
        sum = "h1:AV2c/EiW3KqPNT9ZKl07ehoAGi4C5/01Cfbblndcapg=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_konsorten_go_windows_terminal_sequences",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/konsorten/go-windows-terminal-sequences",
        sum = "h1:mweAR1A6xJ3oS2pRaGiHgQ4OO8tzTaLawm8vnODuwDk=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_kr_logfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kr/logfmt",
        sum = "h1:T+h1c/A9Gawja4Y9mFVWj2vyii2bbUNDw3kt9VxK2EY=",
        version = "v0.0.0-20140226030751-b84e30acd515",
    )
    go_repository(
        name = "com_github_kr_pretty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kr/pretty",
        sum = "h1:L/CwN0zerZDmRFUapSPitk6f+Q3+0za1rQkzVuMiMFI=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_kr_pty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kr/pty",
        sum = "h1:VkoXIwSboBpnk99O/KFauAEILuNHv5DVFKZMBN/gUgw=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_kr_text",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/kr/text",
        sum = "h1:45sCR5RtlFHMR4UwH9sdQ5TC8v0qDQCHnXt+kaKSTVE=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_magiconair_properties",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/magiconair/properties",
        sum = "h1:LLgXmsheXeRoUOBOjtwPQCWIYqM/LU1ayDtDePerRcY=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_github_mailru_easyjson",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mailru/easyjson",
        sum = "h1:2gxZ0XQIU/5z3Z3bUBu+FXuk2pFbkN6tcwi/pjyaDic=",
        version = "v0.0.0-20180823135443-60711f1a8329",
    )
    go_repository(
        name = "com_github_mattn_go_colorable",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-colorable",
        sum = "h1:isR/L+BIZ+rqODWYR/f526ygrBMGKZYFhaaFRDGvuZ8=",
        version = "v0.0.0-20170327083344-ded68f7a9561",
    )
    go_repository(
        name = "com_github_mattn_go_isatty",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mattn/go-isatty",
        sum = "h1:ns/ykhmWi7G9O+8a448SecJU3nSMBXJfqQkl0upE1jI=",
        version = "v0.0.3",
    )
    go_repository(
        name = "com_github_matttproud_golang_protobuf_extensions",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/matttproud/golang_protobuf_extensions",
        sum = "h1:4hp9jkHxhMHkqkrB3Ix0jegS5sx/RkqARlsWZ6pIwiU=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_mitchellh_go_homedir",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/go-homedir",
        sum = "h1:lukF9ziXFxDFPkA1vsr5zpc1XuPDn/wFntq5mG+4E0Y=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_mitchellh_mapstructure",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mitchellh/mapstructure",
        sum = "h1:fmNYVwqnSfB9mZU6OS2O6GsXM+wcskZDuKQzvN1EDeE=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_modern_go_concurrent",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/modern-go/concurrent",
        sum = "h1:TRLaZ9cD/w8PVh93nsPXa1VrQ6jlwL5oN8l14QlcNfg=",
        version = "v0.0.0-20180306012644-bacd9c7ef1dd",
    )
    go_repository(
        name = "com_github_modern_go_reflect2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/modern-go/reflect2",
        sum = "h1:9f412s+6RmYXLWZSEzVVgPGK7C2PphHj5RJrvfx9AWI=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_munnerz_goautoneg",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/munnerz/goautoneg",
        sum = "h1:7PxY7LVfSZm7PEeBTyK1rj1gABdCO2mbri6GKO1cMDs=",
        version = "v0.0.0-20120707110453-a547fc61f48d",
    )
    go_repository(
        name = "com_github_mwitkow_go_conntrack",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mwitkow/go-conntrack",
        sum = "h1:F9x/1yl3T2AeKLr2AMdilSD8+f9bvMnNN8VS5iDtovc=",
        version = "v0.0.0-20161129095857-cc309e4a2223",
    )
    go_repository(
        name = "com_github_mxk_go_flowrate",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/mxk/go-flowrate",
        sum = "h1:y5//uYreIhSUg3J1GEMiLbxo1LJaP8RfCpH6pymGZus=",
        version = "v0.0.0-20140419014527-cca7078d478f",
    )
    go_repository(
        name = "com_github_natefinch_lumberjack",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/natefinch/lumberjack",
        sum = "h1:4QJd3OLAMgj7ph+yZTuX13Ld4UpgHp07nNdFX7mqFfM=",
        version = "v2.0.0+incompatible",
    )
    go_repository(
        name = "com_github_nytimes_gziphandler",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/NYTimes/gziphandler",
        sum = "h1:lsxEuwrXEAokXB9qhlbKWPpo3KMLZQ5WB5WLQRW1uq0=",
        version = "v0.0.0-20170623195520-56545f4a5d46",
    )
    go_repository(
        name = "com_github_onsi_ginkgo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/ginkgo",
        sum = "h1:q/mM8GF/n0shIN8SaAZ0V+jnLPzen6WIVZdiwrRlMlo=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_github_onsi_gomega",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/onsi/gomega",
        sum = "h1:XPnZz8VVBHjVsy1vzJmRwIcSwiUO+JFfrv/xGiigmME=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_github_pborman_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pborman/uuid",
        sum = "h1:J7Q5mO4ysT1dv8hyrUGHb9+ooztCXu1D8MY8DZYsu3g=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_pelletier_go_toml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pelletier/go-toml",
        sum = "h1:T5zMGML61Wp+FlcbWjRDT7yAxhJNAiPPLOFECq181zc=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_peterbourgon_diskv",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/peterbourgon/diskv",
        sum = "h1:UBdAOUP5p4RWqPBg048CAvpKN+vxiaj6gdUUzhl4XmI=",
        version = "v2.0.1+incompatible",
    )
    go_repository(
        name = "com_github_peterh_liner",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/peterh/liner",
        sum = "h1:8uaXtUkxiy+T/zdLWuxa/PG4so0TPZDZfafFNNSaptE=",
        version = "v0.0.0-20170317030525-88609521dc4b",
    )
    go_repository(
        name = "com_github_pkg_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pkg/errors",
        sum = "h1:iURUrRGxPUNPdy5/HRSm+Yj6okJ6UtLINN0Q9M4+h3I=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_github_pkg_profile",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pkg/profile",
        sum = "h1:wBza4Dlm/NCQF572oSGNZ69flNFxlwIHjtwS6oy3Rvw=",
        version = "v0.0.0-20170413231811-06b906832ed0",
    )
    go_repository(
        name = "com_github_pmezard_go_difflib",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_pquerna_cachecontrol",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/pquerna/cachecontrol",
        sum = "h1:0XM1XL/OFFJjXsYXlG30spTkV/E9+gmd5GD1w2HE8xM=",
        version = "v0.0.0-20171018203845-0dec1b30a021",
    )
    go_repository(
        name = "com_github_prometheus_client_golang",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/client_golang",
        sum = "h1:vrDKnkGzuGvhNAL56c7DBz29ZL+KxnoR0x7enabFceM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_prometheus_client_model",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/client_model",
        sum = "h1:S/YWwWx/RA8rT8tKFRuGUZhuA90OyIBpPCXkcbwU8DE=",
        version = "v0.0.0-20190129233127-fd36f4220a90",
    )
    go_repository(
        name = "com_github_prometheus_common",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/common",
        sum = "h1:K0MGApIoQvMw27RTdJkPbr3JZ7DNbtxQNyi5STVM6Kw=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_prometheus_procfs",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/prometheus/procfs",
        sum = "h1:3+auTFlqw+ZaQYJARz6ArODtkaIwtvBTx3N2NehQlL8=",
        version = "v0.0.5",
    )
    go_repository(
        name = "com_github_puerkitobio_purell",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/PuerkitoBio/purell",
        sum = "h1:rmGxhojJlM0tuKtfdvliR84CFHljx9ag64t2xmVkjK4=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_puerkitobio_urlesc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/PuerkitoBio/urlesc",
        sum = "h1:d+Bc7a5rLufV/sSk/8dngufqelfh6jnri85riMAaF/M=",
        version = "v0.0.0-20170810143723-de5bf2ad4578",
    )
    go_repository(
        name = "com_github_remyoudompheng_bigfft",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/remyoudompheng/bigfft",
        sum = "h1:/NRJ5vAYoqz+7sG51ubIDHXeWO8DlTSrToPu6q11ziA=",
        version = "v0.0.0-20170806203942-52369c62f446",
    )
    go_repository(
        name = "com_github_russross_blackfriday",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/russross/blackfriday",
        sum = "h1:HyvC0ARfnZBqnXwABFeSZHpKvJHJJfPz81GNueLj0oo=",
        version = "v1.5.2",
    )
    go_repository(
        name = "com_github_sanathkr_go_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sanathkr/go-yaml",
        sum = "h1:jUK33OXuZP/l6babJtnLo1qsGvq6G9so9KMflGAm4YA=",
        version = "v0.0.0-20170819195128-ed9d249f429b",
    )
    go_repository(
        name = "com_github_sanathkr_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sanathkr/yaml",
        sum = "h1:fOCp11H0yuyAt2wqlbJtbyPzSgaxHTv8uN1pMpkG1t8=",
        version = "v0.0.0-20170819201035-0056894fa522",
    )
    go_repository(
        name = "com_github_sergi_go_diff",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sergi/go-diff",
        sum = "h1:Kpca3qRNrduNnOQeazBd0ysaKrUJiIuISHxogkT9RPQ=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_sirupsen_logrus",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:juTguoYk5qI21pwyTXY3B3Y5cOTH3ZUyZCg1v/mihuo=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_soheilhy_cmux",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/soheilhy/cmux",
        sum = "h1:09wy7WZk4AqO03yH85Ex1X+Uo3vDsil3Fa9AgF8Emss=",
        version = "v0.1.3",
    )
    go_repository(
        name = "com_github_spf13_afero",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/afero",
        sum = "h1:5jhuqJyZCZf2JRofRvN/nIFgIWNzPa3/Vz8mYylgbWc=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_spf13_cast",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/cast",
        sum = "h1:oget//CVOEoFewqQxwr0Ej5yjygnqGkvggSE/gB35Q8=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_spf13_cobra",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/cobra",
        sum = "h1:f0B+LkLX6DtmRH1isoNA9VTtNUK9K8xYd28JNNfOv/s=",
        version = "v0.0.5",
    )
    go_repository(
        name = "com_github_spf13_jwalterweatherman",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/jwalterweatherman",
        sum = "h1:XHEdyB+EcvlqZamSM4ZOMGlc93t6AcsBEu9Gc1vn7yk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_spf13_pflag",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/pflag",
        sum = "h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=",
        version = "v1.0.5",
    )
    go_repository(
        name = "com_github_spf13_viper",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/spf13/viper",
        sum = "h1:VUFqw5KcqRf7i70GOzW7N+Q7+gxVBkSSqiXB12+JQ4M=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_stretchr_objx",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/stretchr/objx",
        sum = "h1:2vfRuCMp5sSVIDSqO8oNnWJq7mPa6KVP3iPIwFBuy8A=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_stretchr_testify",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/stretchr/testify",
        sum = "h1:2E4SXV/wtOkTonXsotYi4li6zVWxYlZuYNCXe9XRJyk=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_tmc_grpc_websocket_proxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tmc/grpc-websocket-proxy",
        sum = "h1:ndzgwNDnKIqyCvHTXaCqh9KlOWKvBry6nuXMJmonVsE=",
        version = "v0.0.0-20170815181823-89b8d40f7ca8",
    )
    go_repository(
        name = "com_github_ugorji_go_codec",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ugorji/go/codec",
        sum = "h1:3SVOIvH7Ae1KRYyQWRjXWJEA9sS/c/pjvH++55Gr648=",
        version = "v0.0.0-20181204163529-d75b2dcb6bc8",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonpointer",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xeipuuv/gojsonpointer",
        sum = "h1:0TnXeVP6mx+A4CBf8cQVkQfkhyGBQCmJcT4g6zKzm7M=",
        version = "v0.0.0-20170225233418-6fe8760cad35",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonreference",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xeipuuv/gojsonreference",
        sum = "h1:XZWnr3bsDQWAZg4Ne+cPoXRPILrNlPNQfxBuwLl43is=",
        version = "v0.0.0-20150808065054-e02fc20de94c",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonschema",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xeipuuv/gojsonschema",
        sum = "h1:yhqBHs09SmmUoNOHc9jgK4a60T3XFRtPAkYxVnqgY50=",
        version = "v0.0.0-20181112162635-ac52e6811b56",
    )
    go_repository(
        name = "com_github_xiang90_probing",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xiang90/probing",
        sum = "h1:MPPkRncZLN9Kh4MEFmbnK4h3BD7AUmskWv2+EeZJCCs=",
        version = "v0.0.0-20160813154853-07dd2e8dfe18",
    )
    go_repository(
        name = "com_github_xordataexchange_crypt",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/xordataexchange/crypt",
        sum = "h1:ESFSdwYZvkeru3RtdrYueztKhOBCSAAzS4Gf+k0tEow=",
        version = "v0.0.3-0.20170626215501-b2862e3d0a77",
    )
    go_repository(
        name = "com_google_cloud_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "cloud.google.com/go",
        sum = "h1:eOI3/cP2VTU6uZLDYAoic+eyzzB9YyGmJ7eIjl8rOPg=",
        version = "v0.34.0",
    )
    go_repository(
        name = "in_gopkg_airbrake_gobrake_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/airbrake/gobrake.v2",
        sum = "h1:7z2uVWwn7oVeeugY1DtlPAy5H+KYgB1KeKTnqjNatLo=",
        version = "v2.0.9",
    )
    go_repository(
        name = "in_gopkg_alecthomas_kingpin_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/alecthomas/kingpin.v2",
        sum = "h1:jMFz6MfLP0/4fUyZle81rXUoxOBFi19VUFKVDOQfozc=",
        version = "v2.2.6",
    )
    go_repository(
        name = "in_gopkg_check_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/check.v1",
        sum = "h1:YR8cESwS4TdDjEe65xsg0ogRM/Nc3DYOhEAlW+xobZo=",
        version = "v1.0.0-20190902080502-41f04d3bba15",
    )
    go_repository(
        name = "in_gopkg_fsnotify_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/fsnotify.v1",
        sum = "h1:xOHLXZwVvI9hhs+cLKq5+I5onOuwQLhQwiu63xxlHs4=",
        version = "v1.4.7",
    )
    go_repository(
        name = "in_gopkg_gemnasium_logrus_airbrake_hook_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/gemnasium/logrus-airbrake-hook.v2",
        sum = "h1:OAj3g0cR6Dx/R07QgQe8wkA9RNjB2u4i700xBkIT4e0=",
        version = "v2.1.2",
    )
    go_repository(
        name = "in_gopkg_inf_v0",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/inf.v0",
        sum = "h1:73M5CoZyi3ZLMOyDlQh031Cx6N9NDJ2Vvfl76EDAgDc=",
        version = "v0.9.1",
    )
    go_repository(
        name = "in_gopkg_natefinch_lumberjack_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/natefinch/lumberjack.v2",
        sum = "h1:986b60BAz5vO2Vaf48yQaq+wb2bU4JsXxKu1+itW6x8=",
        version = "v2.0.0-20150622162204-20b71e5b60d7",
    )
    go_repository(
        name = "in_gopkg_square_go_jose_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/square/go-jose.v2",
        sum = "h1:ELQJ5WuT+ydETLCpWvAuw8iGBQRGoJq+A3RAbbAcZUY=",
        version = "v2.0.0-20180411045311-89060dee6a84",
    )
    go_repository(
        name = "in_gopkg_tomb_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/tomb.v1",
        sum = "h1:uRGJdciOHaEIrze2W8Q3AKkepLTh2hOroT7a+7czfdQ=",
        version = "v1.0.0-20141024135613-dd632973f1e7",
    )
    go_repository(
        name = "in_gopkg_yaml_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/yaml.v1",
        sum = "h1:POO/ycCATvegFmVuPpQzZFJ+pGZeX22Ufu6fibxDVjU=",
        version = "v1.0.0-20140924161607-9f9df34309c0",
    )
    go_repository(
        name = "in_gopkg_yaml_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/yaml.v2",
        sum = "h1:/eiJrUcujPVeJ3xlSWaiNi3uSVmDGBK1pDHUHAnao1I=",
        version = "v2.2.4",
    )
    go_repository(
        name = "in_gopkg_yaml_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gopkg.in/yaml.v3",
        sum = "h1:Xe2gvTZUJpsvOWUnvmL/tmhVBZUmHSvLbMjRj6NUUKo=",
        version = "v3.0.0-20200121175148-a6ecf24a6d71",
    )
    go_repository(
        name = "io_k8s_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/api",
        sum = "h1:bkwe5LsuANqyOwsBng5Qc4S91D2Tv0JHctAztt3YTQs=",
        version = "v0.0.0-20190918195907-bd6ac527cfd2",
    )
    go_repository(
        name = "io_k8s_apiextensions_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apiextensions-apiserver",
        sum = "h1:Kl/sh+wWzYK2hWFZtwvuFECup1SbE2kXfMnhGZsoO5M=",
        version = "v0.0.0-20190918201827-3de75813f604",
    )
    go_repository(
        name = "io_k8s_apimachinery",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apimachinery",
        sum = "h1:7Kns6qqhMAQWvGkxYOLSLRZ5hJO0/5pcE5lPGP2fxUw=",
        version = "v0.0.0-20190817020851-f2f3a405f61d",
    )
    go_repository(
        name = "io_k8s_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/apiserver",
        sum = "h1:vXeVIqZsVgwfUENTAeybgfNVW7zZD7GPZJGtwReK+hI=",
        version = "v0.0.0-20190918200908-1e17798da8c1",
    )
    go_repository(
        name = "io_k8s_client_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/client-go",
        sum = "h1:huOvPq1vO7dkuw9rZPYsLGpFmyGvy6L8q6mDItgkdQ4=",
        version = "v0.0.0-20190918200256-06eb1244587a",
    )
    go_repository(
        name = "io_k8s_cluster_bootstrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/cluster-bootstrap",
        sum = "h1:5wvjieVoU4oovHlkeD256q2M2YYi2P01zk6wxSR2zk0=",
        version = "v0.0.0-20190516232516-d7d78ab2cfe7",
    )
    go_repository(
        name = "io_k8s_code_generator",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/code-generator",
        sum = "h1:p+PRuwXWwk5e+UYvicGiavEupapqM5NOxUl3y1GkD6c=",
        version = "v0.0.0-20190612205613-18da4a14b22b",
    )
    go_repository(
        name = "io_k8s_component_base",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/component-base",
        sum = "h1:YicOHTLJZz/TIpJcqhVYJI2LyuM7VMkYiiG6FZfJmzY=",
        version = "v0.0.0-20190918200425-ed2f0867c778",
    )
    go_repository(
        name = "io_k8s_gengo",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/gengo",
        sum = "h1:4s3/R4+OYYYUKptXPhZKjQ04WJ6EhQQVFdjOFvCazDk=",
        version = "v0.0.0-20190128074634-0689ccc1d7d6",
    )
    go_repository(
        name = "io_k8s_klog",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/klog",
        sum = "h1:Pt+yjF5aB1xDSVbau4VsWe+dQNzA0qv1LlXdC2dF6Q8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "io_k8s_kube_openapi",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kube-openapi",
        sum = "h1:EYm5AW/UUDbnmnI+gK0TJDVK9qPLhM+sRHYanNKw0EQ=",
        version = "v0.0.0-20190816220812-743ec37842bf",
    )
    go_repository(
        name = "io_k8s_kubernetes",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/kubernetes",
        sum = "h1:Gdq2hPpttbaJBoClIanCE6WSu4IZReA54yhkZtvPUOo=",
        version = "v1.14.2",
    )
    go_repository(
        name = "io_k8s_sigs_cluster_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/cluster-api",
        sum = "h1:WjhtuvyjnMgo62kKlVizhI/nYs4DJxHNf+ZMSk/uUsM=",
        version = "v0.2.7",
    )
    go_repository(
        name = "io_k8s_sigs_cluster_api_bootstrap_provider_kubeadm",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/cluster-api-bootstrap-provider-kubeadm",
        sum = "h1:GMpyX3FYSeHmXK+6VHq+khIKV5KkqGXfkbtAJN0Rm04=",
        version = "v0.1.3",
    )
    go_repository(
        name = "io_k8s_sigs_cluster_api_provider_aws",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/cluster-api-provider-aws",
        sum = "h1:WMFAQ7fxOVHvcSGPZnFnNG3Rn+r5qZ39FHkPjOIbwPA=",
        version = "v0.4.6",
    )
    go_repository(
        name = "io_k8s_sigs_controller_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/controller-runtime",
        sum = "h1:ZtdgqJXVHsIytjdmDuk0QjagnzyLq9FjojXRqIp+dU4=",
        version = "v0.3.0",
    )
    go_repository(
        name = "io_k8s_sigs_structured_merge_diff",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/structured-merge-diff",
        sum = "h1:4Z09Hglb792X0kfOBBJUPFEyvVfQWrYT/l8h5EKA6JQ=",
        version = "v0.0.0-20190525122527-15d366b2352e",
    )
    go_repository(
        name = "io_k8s_sigs_testing_frameworks",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/testing_frameworks",
        sum = "h1:GtDhkj3cF4A4IW+A9LScsuxvJqA9DE7G7PGH1f8B07U=",
        version = "v0.1.2-0.20190130140139-57f07443c2d4",
    )
    go_repository(
        name = "io_k8s_sigs_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "sigs.k8s.io/yaml",
        sum = "h1:4A07+ZFc2wgJwo8YNlQpr1rVlgUDlxXHhPJciaPY5gs=",
        version = "v1.1.0",
    )
    go_repository(
        name = "io_k8s_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "k8s.io/utils",
        sum = "h1:uy5HAgt4Ha5rEMbhZA+aM1j2cq5LmR6LQ71EYC2sVH4=",
        version = "v0.0.0-20190809000727-6c36bc71fc4a",
    )
    go_repository(
        name = "net_starlark_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.starlark.net",
        sum = "h1:lkYv5AKwvvduv5XWP6szk/bvvgO6aDeUujhZQXIFTes=",
        version = "v0.0.0-20190702223751-32f345186213",
    )
    go_repository(
        name = "org_golang_google_appengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/appengine",
        sum = "h1:KxkO13IPW4Lslp2bz+KHP2E3gtFlrIGNThxkZQ3g+4c=",
        version = "v1.5.0",
    )
    go_repository(
        name = "org_golang_google_genproto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/genproto",
        sum = "h1:72GtwBPfq6av9X0Ru2HtAopsPW+d+vh1K1zaxanTdE8=",
        version = "v0.0.0-20170731182057-09f6ed296fc6",
    )
    go_repository(
        name = "org_golang_google_grpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/grpc",
        sum = "h1:bHIbVsCwmvbArgCJmLdgOdHFXlKqTOVjbibbS19cXHc=",
        version = "v1.13.0",
    )
    go_repository(
        name = "org_golang_x_arch",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/arch",
        sum = "h1:TP7YcWHbnFq4v8/3wM2JwgM0SRRtsYJ7Z6Oj0arz2bs=",
        version = "v0.0.0-20171004143515-077ac972c2e4",
    )
    go_repository(
        name = "org_golang_x_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/crypto",
        sum = "h1:0hQKqeLdqlt5iIwVOBErRisrHJAN57yOiPRQItI20fU=",
        version = "v0.0.0-20190911031432-227b76d455e7",
    )
    go_repository(
        name = "org_golang_x_exp",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/exp",
        sum = "h1:I6A9Ag9FpEKOjcKrRNjQkPHawoXIhKyTGfvvjFAiiAk=",
        version = "v0.0.0-20190312203227-4b39c73a6495",
    )
    go_repository(
        name = "org_golang_x_image",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/image",
        sum = "h1:KYGJGHOQy8oSi1fDlSpcZF0+juKwk/hEMv5SiwHogR0=",
        version = "v0.0.0-20190227222117-0694c2d4d067",
    )
    go_repository(
        name = "org_golang_x_mobile",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/mobile",
        sum = "h1:Tus/Y4w3V77xDsGwKUC8a/QrV7jScpU557J77lFffNs=",
        version = "v0.0.0-20190312151609-d3739f865fa6",
    )
    go_repository(
        name = "org_golang_x_net",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/net",
        sum = "h1:XfVGCX+0T4WOStkaOsJRllbsiImhB2jgVBGc9L0lPGc=",
        version = "v0.0.0-20190909003024-a7b16738d86b",
    )
    go_repository(
        name = "org_golang_x_oauth2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/oauth2",
        sum = "h1:SVwTIAaPC2U/AvvLNZ2a7OVsmBpC8L5BlwK1whH3hm0=",
        version = "v0.0.0-20190604053449-0f29369cfe45",
    )
    go_repository(
        name = "org_golang_x_sync",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/sync",
        sum = "h1:8gQV6CLnAEikrhgkHFbMAEhagSSnXWGV915qUMm9mrU=",
        version = "v0.0.0-20190423024810-112230192c58",
    )
    go_repository(
        name = "org_golang_x_sys",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/sys",
        sum = "h1:SW/0nsKCUaozCUtZTakri5laocGx/5bkDSSLrFUsa5s=",
        version = "v0.0.0-20190911201528-7ad0cfa0b7b5",
    )
    go_repository(
        name = "org_golang_x_text",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/text",
        sum = "h1:tW2bmiBqwgJj/UpqtC8EpXEZVYOwU0yG4iWbprSVAcs=",
        version = "v0.3.2",
    )
    go_repository(
        name = "org_golang_x_time",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/time",
        sum = "h1:SvFZT6jyqRaOeXpc5h/JSfZenJ2O330aBsf7JfSUXmQ=",
        version = "v0.0.0-20190308202827-9d24e82272b4",
    )
    go_repository(
        name = "org_golang_x_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/tools",
        sum = "h1:qsl9y/CJx34tuA7QCPNp86JNJe4spst6Ff8MjvPUdPg=",
        version = "v0.0.0-20190425150028-36563e24a262",
    )
    go_repository(
        name = "org_golang_x_xerrors",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/xerrors",
        sum = "h1:9zdDQZ7Thm29KFXgAX/+yaf3eVbP7djjWp/dXAppNCc=",
        version = "v0.0.0-20190717185122-a985d3407aa7",
    )
    go_repository(
        name = "org_gonum_v1_gonum",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gonum.org/v1/gonum",
        sum = "h1:OB/uP/Puiu5vS5QMRPrXCDWUPb+kt8f1KW8oQzFejQw=",
        version = "v0.0.0-20190331200053-3d26580ed485",
    )
    go_repository(
        name = "org_gonum_v1_netlib",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gonum.org/v1/netlib",
        sum = "h1:jRyg0XfpwWlhEV8mDfdNGBeSJM2fuyh9Yjrnd8kF2Ts=",
        version = "v0.0.0-20190331212654-76723241ea4e",
    )
    go_repository(
        name = "org_modernc_cc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "modernc.org/cc",
        sum = "h1:nPibNuDEx6tvYrUAtvDTTw98rx5juGsa5zuDnKwEEQQ=",
        version = "v1.0.0",
    )
    go_repository(
        name = "org_modernc_golex",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "modernc.org/golex",
        sum = "h1:wWpDlbK8ejRfSyi0frMyhilD3JBvtcx2AdGDnU+JtsE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "org_modernc_mathutil",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "modernc.org/mathutil",
        sum = "h1:93vKjrJopTPrtTNpZ8XIovER7iCIH1QU7wNbOQXC60I=",
        version = "v1.0.0",
    )
    go_repository(
        name = "org_modernc_strutil",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "modernc.org/strutil",
        sum = "h1:XVFtQwFVwc02Wk+0L/Z/zDDXO81r5Lhe6iMKmGX3KhE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "org_modernc_xc",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "modernc.org/xc",
        sum = "h1:7ccXrupWZIS3twbUGrtKmHS2DXY6xegFua+6O3xgAFU=",
        version = "v1.0.0",
    )
    go_repository(
        name = "org_uber_go_atomic",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/atomic",
        sum = "h1:cxzIVoETapQEqDhQu3QfnvXAV4AlzcvUCxkVUFw3+EU=",
        version = "v1.4.0",
    )
    go_repository(
        name = "org_uber_go_multierr",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/multierr",
        sum = "h1:HoEmRHQPVSqub6w2z2d2EOVs2fjyFRGyofhKuyDq0QI=",
        version = "v1.1.0",
    )
    go_repository(
        name = "org_uber_go_zap",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "go.uber.org/zap",
        sum = "h1:ORx85nbTijNz8ljznvCMR1ZBIPKFn3jQrag10X2AsuM=",
        version = "v1.10.0",
    )
    go_repository(
        name = "tools_gotest",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gotest.tools",
        sum = "h1:VsBPFP1AI068pPrMxtb/S8Zkgf9xEmTLJjfM+P5UIEo=",
        version = "v2.2.0+incompatible",
    )
    go_repository(
        name = "xyz_gomodules_jsonpatch_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "gomodules.xyz/jsonpatch/v2",
        sum = "h1:xyiBuvkD2g5n7cYzx6u2sxQvsAy4QJsZFCzGVdzOXZ0=",
        version = "v2.0.1",
    )
