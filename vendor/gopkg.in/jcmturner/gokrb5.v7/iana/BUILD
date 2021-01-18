load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["constants.go"],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/gokrb5.v7/iana",
    importpath = "gopkg.in/jcmturner/gokrb5.v7/iana",
    visibility = ["//visibility:public"],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/addrtype:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/adtype:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/asnAppTag:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/chksumtype:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/errorcode:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/etypeID:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/flags:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/keyusage:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/msgtype:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/nametype:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/patype:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
