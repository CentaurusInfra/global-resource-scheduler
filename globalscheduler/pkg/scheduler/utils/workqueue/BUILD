load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["parallelizer.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/workqueue",
    visibility = ["//visibility:public"],
    deps = ["//globalscheduler/pkg/scheduler/utils/runtime:go_default_library"],
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
