load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "bitreader.go",
        "bitwriter.go",
        "bytereader.go",
        "compress.go",
        "decompress.go",
        "huff0.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/klauspost/compress/huff0",
    importpath = "github.com/klauspost/compress/huff0",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/klauspost/compress/fse:go_default_library"],
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
