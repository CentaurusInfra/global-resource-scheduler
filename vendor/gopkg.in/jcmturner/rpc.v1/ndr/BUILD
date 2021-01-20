load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "arrays.go",
        "decoder.go",
        "error.go",
        "header.go",
        "pipe.go",
        "primitives.go",
        "rawbytes.go",
        "strings.go",
        "tags.go",
        "union.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/rpc.v1/ndr",
    importpath = "gopkg.in/jcmturner/rpc.v1/ndr",
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
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
