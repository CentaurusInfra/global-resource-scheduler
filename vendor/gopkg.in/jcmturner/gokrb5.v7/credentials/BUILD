load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "ccache.go",
        "credentials.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/gokrb5.v7/credentials",
    importpath = "gopkg.in/jcmturner/gokrb5.v7/credentials",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/hashicorp/go-uuid:go_default_library",
        "//vendor/github.com/jcmturner/gofork/encoding/asn1:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/nametype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/keytab:go_default_library",
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
