load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["registry.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/algorithmprovider",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/framework/plugins/defaultbinder:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/flavor:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/locationandoperator:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/queuesort:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/regionandaz:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
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
