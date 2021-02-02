load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["collector.go"],
    importpath = "k8s.io/kubernetes/cmd/resource-collector",
    visibility = ["//visibility:private"],
    deps = [
        "//cmd/resource-collector/app:go_default_library",
        "//resourcecollector/pkg/collector/common/config:go_default_library",
    ],
)

go_binary(
    name = "resource-collector",
    embed = [":go_default_library"],
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
        "//cmd/resource-collector/app:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
