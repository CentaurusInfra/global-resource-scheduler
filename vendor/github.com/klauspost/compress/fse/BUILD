load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "bitreader.go",
        "bitwriter.go",
        "bytereader.go",
        "compress.go",
        "decompress.go",
        "fse.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/klauspost/compress/fse",
    importpath = "github.com/klauspost/compress/fse",
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
