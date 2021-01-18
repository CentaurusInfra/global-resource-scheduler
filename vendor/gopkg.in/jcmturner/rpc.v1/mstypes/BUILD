load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "claims.go",
        "common.go",
        "filetime.go",
        "group_membership.go",
        "kerb_sid_and_attributes.go",
        "reader.go",
        "rpc_unicode_string.go",
        "sid.go",
        "user_session_key.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/rpc.v1/mstypes",
    importpath = "gopkg.in/jcmturner/rpc.v1/mstypes",
    visibility = ["//visibility:public"],
    deps = ["//vendor/gopkg.in/jcmturner/rpc.v1/ndr:go_default_library"],
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
