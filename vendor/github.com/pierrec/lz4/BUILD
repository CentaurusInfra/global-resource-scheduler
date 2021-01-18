load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "block.go",
        "debug_stub.go",
        "decode_amd64.go",
        "decode_amd64.s",
        "decode_other.go",
        "errors.go",
        "lz4.go",
        "lz4_go1.10.go",
        "lz4_notgo1.10.go",
        "reader.go",
        "writer.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/pierrec/lz4",
    importpath = "github.com/pierrec/lz4",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/pierrec/lz4/internal/xxh32:go_default_library"],
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
        "//vendor/github.com/pierrec/lz4/internal/xxh32:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
