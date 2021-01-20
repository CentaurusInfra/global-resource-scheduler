load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["xxh32zero.go"],
    importmap = "k8s.io/kubernetes/vendor/github.com/pierrec/lz4/internal/xxh32",
    importpath = "github.com/pierrec/lz4/internal/xxh32",
    visibility = ["//vendor/github.com/pierrec/lz4:__subpackages__"],
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
