load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "ASExchange.go",
        "TGSExchange.go",
        "cache.go",
        "client.go",
        "network.go",
        "passwd.go",
        "session.go",
        "settings.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/gokrb5.v7/client",
    importpath = "gopkg.in/jcmturner/gokrb5.v7/client",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/config:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/credentials:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/etype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/errorcode:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/flags:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/keyusage:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/nametype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/patype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/kadmin:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/keytab:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/krberror:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/messages:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/types:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
