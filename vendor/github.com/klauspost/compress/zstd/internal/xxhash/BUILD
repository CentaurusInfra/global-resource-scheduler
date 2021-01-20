load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "xxhash.go",
        "xxhash_amd64.go",
        "xxhash_amd64.s",
        "xxhash_other.go",
        "xxhash_safe.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/klauspost/compress/zstd/internal/xxhash",
    importpath = "github.com/klauspost/compress/zstd/internal/xxhash",
    visibility = ["//vendor/github.com/klauspost/compress/zstd:__subpackages__"],
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
