load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["file_handler.go"],
    importmap = "k8s.io/kubernetes/vendor/github.com/go-chassis/go-archaius/sources/utils",
    importpath = "github.com/go-chassis/go-archaius/sources/utils",
    visibility = ["//visibility:public"],
    deps = ["//vendor/gopkg.in/yaml.v2:go_default_library"],
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
