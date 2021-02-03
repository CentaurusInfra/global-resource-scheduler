load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["server.go"],
    importpath = "k8s.io/kubernetes/resourcecollector/pkg/collector/common/apiserver",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler:go_default_library",
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//resourcecollector/pkg/collector/common/config:go_default_library",
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
